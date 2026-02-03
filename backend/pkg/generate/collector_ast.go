package generate

// This file handles Go AST parsing and type extraction.

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"log/slog"
	"maps"
	"os"
	"reflect"
	"slices"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/tools/go/packages"
)

const DEPRECATED_PREFIX = "deprecated:"

// parseGoTypesDir parses Go type definitions from a directory using go/packages.
func (g *OpenAPICollector) parseGoTypesDir(goTypesDirPath string) (*GoParser, error) {
	g.l.Debug("Parsing Go types directory", slog.String("path", goTypesDirPath))

	if _, err := os.Stat(goTypesDirPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to parse Go types directory: path %s does not exist", goTypesDirPath)
	}

	// Use go/packages to load and type-check the package
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo,
		Dir: goTypesDirPath,
	}

	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to load package from directory %s: %w", goTypesDirPath, err)
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages found in directory: %s", goTypesDirPath)
	}

	pkg := pkgs[0]

	// Fail if there are any errors in the loaded package
	if len(pkg.Errors) > 0 {
		var errMsgs []string
		for _, e := range pkg.Errors {
			errMsgs = append(errMsgs, e.Error())
		}

		return nil, fmt.Errorf("package %s loading failed from directory %s with errors: %s", pkg.Name, goTypesDirPath, strings.Join(errMsgs, "; "))
	}

	g.l.Debug("Go types parsed successfully",
		slog.String("package", pkg.Name),
		slog.Int("fileCount", len(pkg.Syntax)))

	return &GoParser{
		fset:  pkg.Fset,
		files: pkg.Syntax,
		pkg:   pkg,
	}, nil
}

// extractAllTypesFromGo walks the Go AST and extracts all type information in one pass.
func (g *OpenAPICollector) extractAllTypesFromGo(goParser *GoParser) error {
	g.l.Debug("Starting type extraction from Go AST")

	errs := make([]error, 0, len(goParser.files)*2)

	// Walk all files and extract type declarations
	for _, file := range goParser.files {
		// First pass: extract all type declarations
		errs = append(errs, g.extractTypeDeclarations(file)...)

		// Second pass: extract enums from const blocks
		errs = append(errs, g.extractConstDeclarations(file)...)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	g.l.Debug("Completed type extraction", slog.Int("typeCount", len(g.types)))

	return nil
}

// extractTypeDeclarations extracts all type declarations from a single AST file.
func (g *OpenAPICollector) extractTypeDeclarations(file *ast.File) []error {
	var errs []error

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			typeName := typeSpec.Name.Name

			typeInfo, err := g.extractTypeFromSpec(typeName, typeSpec, genDecl)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to extract type %s: %w", typeName, err))

				continue
			}

			g.types[typeName] = typeInfo
		}
	}

	return errs
}

// extractTypeFromSpec extracts TypeInfo from a Go type spec.
func (g *OpenAPICollector) extractTypeFromSpec(name string, typeSpec *ast.TypeSpec, genDecl *ast.GenDecl) (*TypeInfo, error) {
	g.l.Debug("Extracting type", slog.String("name", name))

	// Store the AST node for later Go source generation
	g.typeASTs[name] = genDecl

	// Extract comments
	desc := g.extractCommentsFromDoc(genDecl.Doc)

	deprecated, cleanedDesc, err := g.parseDeprecation(desc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse deprecation info for type %s: %w", name, err)
	}

	typeInfo := &TypeInfo{
		Name:        name,
		Description: cleanedDesc,
		Deprecated:  deprecated,
	}

	// Determine type based on the type expression
	switch t := typeSpec.Type.(type) {
	case *ast.StructType:
		return g.extractStructType(name, t, typeInfo)
	case *ast.Ident:
		// Type alias to another type (e.g., type MyString string)
		typeInfo.Kind = TypeKindAlias

		// Extract the underlying type information
		underlyingType, refs, err := g.analyzeGoType(t)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze underlying type for alias %s: %w", name, err)
		}

		typeInfo.UnderlyingType = &underlyingType
		typeInfo.References = refs

		return typeInfo, nil
	case *ast.ArrayType, *ast.MapType:
		// Arrays and maps as top-level types are treated as aliases
		typeInfo.Kind = TypeKindAlias

		// Extract the underlying type information
		underlyingType, refs, err := g.analyzeGoType(typeSpec.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze underlying type for alias %s: %w", name, err)
		}

		typeInfo.UnderlyingType = &underlyingType
		typeInfo.References = refs

		return typeInfo, nil
	default:
		return nil, fmt.Errorf("unsupported type %s: %T (please use struct, type alias, or basic types)", name, typeSpec.Type)
	}
}

// extractStructType extracts struct type information.
func (g *OpenAPICollector) extractStructType(name string, structType *ast.StructType, typeInfo *TypeInfo) (*TypeInfo, error) {
	typeInfo.Kind = TypeKindObject
	typeInfo.Fields = []FieldInfo{}
	typeInfo.References = []string{}

	refs := make(map[string]struct{})

	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			return nil, fmt.Errorf("embedded fields are not supported in struct type %s", name)
		}

		for _, fieldName := range field.Names {
			if !fieldName.IsExported() {
				// Skip unexported fields
				continue
			}

			fieldInfo, fieldRefs, err := g.extractFieldInfo(name, fieldName.Name, field)
			if err != nil {
				// Skip fields that should be ignored (e.g., json:"-" tags)
				if errors.Is(err, ErrFieldSkipped) {
					continue
				}

				return nil, err
			}

			typeInfo.Fields = append(typeInfo.Fields, fieldInfo)

			// Collect references
			for _, ref := range fieldRefs {
				refs[ref] = struct{}{}
			}
		}
	}

	typeInfo.References = []string{}
	// Convert refs map to sorted slice (single pass)
	if len(refs) > 0 {
		typeInfo.References = slices.Collect(maps.Keys(refs))
		slices.Sort(typeInfo.References)
	}

	g.l.Debug("Extracted struct type", slog.String("name", name), slog.Int("fieldCount", len(typeInfo.Fields)))

	return typeInfo, nil
}

// extractFieldInfo extracts field information from a struct field.
func (g *OpenAPICollector) extractFieldInfo(parentName, fieldName string, field *ast.Field) (FieldInfo, []string, error) {
	// Parse JSON tag
	tagInfo := parseJSONTag(field, fieldName)
	if tagInfo.skip {
		return FieldInfo{}, nil, ErrFieldSkipped
	}

	// Analyze field type
	fieldType, refs, err := g.analyzeGoType(field.Type)
	if err != nil {
		return FieldInfo{}, nil, fmt.Errorf("failed to analyze field type for %s.%s (type: %T): %w", parentName, fieldName, field.Type, err)
	}

	// Determine if field is required
	// In Go: pointer types (*T) are optional, non-pointer types are required unless omitempty is set
	required := !fieldType.Nullable && !tagInfo.omitempty
	fieldType.Required = required

	// Extract field documentation
	fieldDesc := g.extractCommentsFromDoc(field.Doc)

	fieldDeprecated, cleanedFieldDesc, err := g.parseDeprecation(fieldDesc)
	if err != nil {
		return FieldInfo{}, nil, fmt.Errorf("failed to parse deprecation info for field %s.%s: %w", parentName, fieldName, err)
	}

	fieldInfo := FieldInfo{
		Name:        tagInfo.name,
		DisplayType: generateDisplayType(fieldType),
		TypeInfo:    fieldType,
		Description: cleanedFieldDesc,
		Deprecated:  fieldDeprecated,
	}

	return fieldInfo, refs, nil
}

// analyzeGoType analyzes a Go type expression and returns FieldType and referenced types.
func (g *OpenAPICollector) analyzeGoType(expr ast.Expr) (FieldType, []string, error) {
	refs := []string{}

	switch t := expr.(type) {
	case *ast.Ident:
		// Simple type reference (string, int, MyType, etc.)
		typeName := t.Name

		// Check for primitives - map Go types to OpenAPI/JSON Schema types
		if primitiveType, ok := g.primitiveTypeMapping[typeName]; ok {
			return primitiveType, refs, nil
		}

		// Reject 'any' explicitly
		if typeName == "any" {
			return FieldType{}, nil, errors.New("type 'any' is not allowed in API types - use concrete types instead. Check struct fields and type aliases for 'any' usage")
		}

		// Check if it's a defined type in our types map (will be populated after first pass)
		// For now, treat as reference
		refs = append(refs, typeName)

		return FieldType{
			Kind: FieldKindReference,
			Type: typeName,
		}, refs, nil

	case *ast.StarExpr:
		return g.analyzePointerType(t)

	case *ast.ArrayType:
		return g.analyzeArrayType(t)

	case *ast.MapType:
		return g.analyzeMapType(t)

	case *ast.SelectorExpr:
		return g.analyzeSelectorType(t)

	default:
		return FieldType{}, nil, fmt.Errorf("unsupported type expression: %T (check for unsupported Go language features like interfaces, channels, or functions)", expr)
	}
}

// analyzePointerType handles pointer types (*T) which become nullable.
func (g *OpenAPICollector) analyzePointerType(t *ast.StarExpr) (FieldType, []string, error) {
	inner, innerRefs, err := g.analyzeGoType(t.X)
	if err != nil {
		return FieldType{}, nil, err
	}

	inner.Nullable = true

	return inner, innerRefs, nil
}

// analyzeArrayType handles array/slice types ([]T).
func (g *OpenAPICollector) analyzeArrayType(t *ast.ArrayType) (FieldType, []string, error) {
	elemType, elemRefs, err := g.analyzeGoType(t.Elt)
	if err != nil {
		return FieldType{}, nil, err
	}

	return FieldType{
		Kind:      FieldKindArray,
		Type:      "array",
		ItemsType: &elemType,
	}, elemRefs, nil
}

// analyzeMapType handles map types (map[K]V).
func (g *OpenAPICollector) analyzeMapType(t *ast.MapType) (FieldType, []string, error) {
	valueType, valueRefs, err := g.analyzeGoType(t.Value)
	if err != nil {
		return FieldType{}, nil, err
	}

	return FieldType{
		Kind:                 FieldKindObject,
		Type:                 "object",
		AdditionalProperties: &valueType,
	}, valueRefs, nil
}

// analyzeSelectorType handles external types (e.g., time.Time, types.URL).
func (g *OpenAPICollector) analyzeSelectorType(t *ast.SelectorExpr) (FieldType, []string, error) {
	pkgIdent, ok := t.X.(*ast.Ident)
	if !ok {
		return FieldType{}, nil, fmt.Errorf("unsupported selector expression with base type %T - expected package.Type format", t.X)
	}

	fullType := pkgIdent.Name + "." + t.Sel.Name

	// Check for known external types
	format, exists := g.externalTypeFormats[fullType]
	if !exists {
		// Try with full package path
		fullTypePath := "http-mqtt-boilerplate/backend/pkg/" + pkgIdent.Name + "." + t.Sel.Name
		format, exists = g.externalTypeFormats[fullTypePath]

		if !exists {
			return FieldType{}, nil, fmt.Errorf("unknown external type %s - please add it to externalTypeFormats map in NewOpenAPICollector", fullType)
		}
	}

	return FieldType{
		Kind:   FieldKindPrimitive,
		Type:   "string",
		Format: format,
	}, nil, nil
}

// extractCommentsFromDoc extracts text from a comment group.
func (g *OpenAPICollector) extractCommentsFromDoc(doc *ast.CommentGroup) string {
	if doc == nil {
		return ""
	}

	var builder strings.Builder

	for i, comment := range doc.List {
		if i > 0 {
			builder.WriteString(" ")
		}

		// Trim leading slashes and whitespace
		text := strings.TrimPrefix(comment.Text, "//")
		text = strings.TrimSpace(text)

		if text == "" {
			continue
		}

		// Skip nolint comments
		if strings.HasPrefix(text, "nolint:") {
			continue
		}

		builder.WriteString(strings.TrimSpace(text))
	}

	return strings.TrimSpace(builder.String())
}

// parseDeprecation extracts deprecation info from comments and returns cleaned description.
// It looks for "Deprecated:" anywhere in the text (case-insensitive) and captures the message.
// Returns (deprecationInfo, cleanedDescription, error).
func (g *OpenAPICollector) parseDeprecation(comments string) (string, string, error) {
	if comments == "" {
		return "", "", nil
	}

	// Look for "Deprecated:" anywhere in the text (case-insensitive)
	lowerComments := strings.ToLower(comments)
	idx := strings.Index(lowerComments, DEPRECATED_PREFIX)

	if idx == -1 {
		return "", comments, nil
	}

	// Extract the message after "Deprecated:"
	// Start from the original string to preserve casing
	message := strings.TrimSpace(comments[idx+len(DEPRECATED_PREFIX):])

	// Clean the description by removing the deprecation text
	cleanedDesc := strings.TrimSpace(comments[:idx])

	if message == "" {
		return "", cleanedDesc, errors.New("deprecation message is empty - when using 'Deprecated:' comment, provide a message explaining why it's deprecated and what to use instead")
	}

	return message, cleanedDesc, nil
}

// generateDisplayType creates a human-readable type string from FieldType.
func generateDisplayType(ft FieldType) string {
	switch ft.Kind {
	case FieldKindReference, FieldKindEnum:
		return ft.Type
	case FieldKindPrimitive:
		caser := cases.Title(language.English)

		return caser.String(ft.Type)

	case FieldKindArray:
		if ft.ItemsType != nil {
			itemDisplay := generateDisplayType(*ft.ItemsType)

			return itemDisplay + "[]"
		}

		return "Array"

	case FieldKindObject:
		return "Object"

	default:
		panic(fmt.Sprintf("unexpected field kind: %s, should have been caught by type analysis", ft.Kind))
	}
}

// extractTypeNameFromValue extracts the type name from a Go value using reflection.
// If the value is nil, typeName is set to empty string and no error is returned.
func extractTypeNameFromValue(value any) (string, error) {
	if value == nil {
		return "", nil
	}

	rt := reflect.TypeOf(value)
	if rt == nil {
		return "", nil
	}

	// Handle pointers
	for rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}

	name := rt.Name()
	if name == "" {
		return "", errors.New("anonymous type not supported")
	}

	return name, nil
}

func getPrimitiveTypeMappings() map[string]FieldType {
	return map[string]FieldType{
		"string":  {Kind: FieldKindPrimitive, Type: "string"},
		"byte":    {Kind: FieldKindPrimitive, Type: "string"},
		"rune":    {Kind: FieldKindPrimitive, Type: "string"},
		"bool":    {Kind: FieldKindPrimitive, Type: "boolean"},
		"int":     {Kind: FieldKindPrimitive, Type: "integer"},
		"int8":    {Kind: FieldKindPrimitive, Type: "integer"},
		"int16":   {Kind: FieldKindPrimitive, Type: "integer"},
		"uint":    {Kind: FieldKindPrimitive, Type: "integer"},
		"uint8":   {Kind: FieldKindPrimitive, Type: "integer"},
		"uint16":  {Kind: FieldKindPrimitive, Type: "integer"},
		"int32":   {Kind: FieldKindPrimitive, Type: "integer", Format: "int32"},
		"uint32":  {Kind: FieldKindPrimitive, Type: "integer", Format: "int32"},
		"int64":   {Kind: FieldKindPrimitive, Type: "integer", Format: "int64"},
		"uint64":  {Kind: FieldKindPrimitive, Type: "integer", Format: "int64"},
		"float32": {Kind: FieldKindPrimitive, Type: "number", Format: "float"},
		"float64": {Kind: FieldKindPrimitive, Type: "number", Format: "double"},
	}
}

// jsonTagInfo holds parsed JSON struct tag information.
type jsonTagInfo struct {
	name      string
	omitempty bool
	skip      bool
}

// parseJSONTag parses a JSON struct tag and returns the field name, omitempty flag, and skip flag.
func parseJSONTag(field *ast.Field, defaultName string) jsonTagInfo {
	info := jsonTagInfo{
		name: defaultName,
	}

	if field.Tag == nil {
		return info
	}

	// Use reflect.StructTag to properly parse struct tags
	tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))

	jsonTag, ok := tag.Lookup("json")
	if !ok {
		return info
	}

	parts := strings.Split(jsonTag, ",")
	if parts[0] == "-" {
		info.skip = true

		return info
	}

	if parts[0] != "" {
		info.name = parts[0]
	}

	if slices.Contains(parts[1:], "omitempty") {
		info.omitempty = true
	}

	return info
}
