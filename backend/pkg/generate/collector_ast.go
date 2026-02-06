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

// parseGoTypesDirs parses Go type definitions from multiple directories using go/packages.
// All packages are loaded together so they can reference each other properly.
func (g *OpenAPICollector) parseGoTypesDirs(goTypesDirPaths []string) (*GoParser, error) {
	g.l.Debug("Parsing Go types directories", slog.Any("paths", goTypesDirPaths))

	// Validate all paths exist
	for _, path := range goTypesDirPaths {
		if _, err := os.Stat(path); err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to parse Go types directory: path %s does not exist", path)
			}

			return nil, fmt.Errorf("failed to parse Go types directory: path %s does not exist", path)
		}
	}

	// Use a shared file set for all packages
	fset := token.NewFileSet()

	// Use go/packages to load and type-check all packages
	// We load them all at once so they can properly reference each other
	cfg := &packages.Config{
		Fset: fset, // Use our shared file set
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedImports,
	}

	pkgs, err := packages.Load(cfg, goTypesDirPaths...)
	if err != nil {
		return nil, fmt.Errorf("failed to load packages from directories: %w", err)
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages found in directories: %v", goTypesDirPaths)
	}

	// Collect all files from all packages and check for errors
	var allFiles []*ast.File

	for _, pkg := range pkgs {
		// Fail if there are any errors in the loaded package
		if len(pkg.Errors) > 0 {
			var errMsgs []string
			for _, e := range pkg.Errors {
				errMsgs = append(errMsgs, e.Error())
			}

			return nil, fmt.Errorf("package %s loading failed with errors: %s", pkg.Name, strings.Join(errMsgs, "; "))
		}

		g.l.Debug("Loaded package",
			slog.String("package", pkg.Name),
			slog.Int("fileCount", len(pkg.Syntax)))

		// Add all AST files from this package
		allFiles = append(allFiles, pkg.Syntax...)
	}

	g.l.Debug("Go types parsed successfully",
		slog.Int("packageCount", len(pkgs)),
		slog.Int("totalFileCount", len(allFiles)))

	// Return a GoParser with all files from all packages
	return &GoParser{
		fset:     fset,
		files:    allFiles,
		packages: pkgs, // Store all packages for import resolution
	}, nil
}

// extractAllTypesFromGo walks the Go AST and extracts all type information in two passes.
// Pass 1: Extract all type names and enum metadata (no field analysis).
// Pass 2: Extract field details for all types.
func (g *OpenAPICollector) extractAllTypesFromGo(goParser *GoParser) error {
	g.l.Debug("Starting type extraction from Go AST")

	// Pass 1: Extract all type names and enum metadata (no field analysis)
	for _, file := range goParser.files {
		// Build import alias map once per file
		imports, err := g.buildImportAliasMap(file)
		if err != nil {
			return fmt.Errorf("failed to build import alias map: %w", err)
		}

		g.currentFileImports = imports

		if err := g.extractTypeNames(file); err != nil {
			return fmt.Errorf("failed to extract type names: %w", err)
		}

		if err := g.extractConstDeclarations(file); err != nil {
			return fmt.Errorf("failed to extract const declarations: %w", err)
		}
	}

	g.l.Debug("Pass 1 complete: extracted type names", slog.Int("typeCount", len(g.types)))

	// Pass 2: Extract field details for all types
	for _, file := range goParser.files {
		// Rebuild import alias map for this file
		imports, err := g.buildImportAliasMap(file)
		if err != nil {
			return fmt.Errorf("failed to build import alias map: %w", err)
		}

		g.currentFileImports = imports

		if err := g.extractTypeFields(file); err != nil {
			return fmt.Errorf("failed to extract type fields: %w", err)
		}
	}

	g.l.Debug("Pass 2 complete: extracted fields", slog.Int("typeCount", len(g.types)))

	return nil
}

// buildImportAliasMap extracts import aliases from an AST file.
// Returns a map from package identifier (as used in code) to full import path.
func (g *OpenAPICollector) buildImportAliasMap(file *ast.File) (map[string]string, error) {
	imports := make(map[string]string)

	for _, imp := range file.Imports {
		// Get the import path without quotes
		importPath := strings.Trim(imp.Path.Value, `"`)

		var pkgIdentifier string
		if imp.Name != nil {
			// Explicit alias (e.g., import customtime "time")
			pkgIdentifier = imp.Name.Name
		} else {
			// No explicit alias - get the actual package name from go/packages
			// This handles cases like "gopkg.in/yaml.v3" where package name is "yaml", not "v3"
			// Search through all loaded packages to find this import
			var importedPkg *packages.Package

			for _, pkg := range g.goParser.packages {
				if imp, exists := pkg.Imports[importPath]; exists {
					importedPkg = imp

					break
				}
			}

			if importedPkg == nil {
				return nil, fmt.Errorf("import %s not found in any loaded package's Imports map", importPath)
			}

			if importedPkg.Name == "" {
				return nil, fmt.Errorf("package name is empty for import %s", importPath)
			}

			pkgIdentifier = importedPkg.Name
		}

		imports[pkgIdentifier] = importPath
	}

	return imports, nil
}

// extractTypeNames extracts type names and basic metadata without analyzing fields.
// This is Pass 1 - creates stub TypeInfo entries in g.types so later lookups succeed.
// Requires g.currentFileImports to be set before calling (by extractAllTypesFromGo).
func (g *OpenAPICollector) extractTypeNames(file *ast.File) error {
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

			// Store the AST node for later analysis
			g.typeASTs[typeName] = genDecl

			// Extract comments - prefer typeSpec.Doc (for grouped declarations) over genDecl.Doc (fallback)
			descDoc := typeSpec.Doc
			if descDoc == nil {
				descDoc = genDecl.Doc
			}

			desc := g.extractCommentsFromDoc(descDoc)

			deprecated, cleanedDesc, err := g.parseDeprecation(desc)
			if err != nil {
				return fmt.Errorf("failed to parse deprecation info for type %s: %w", typeName, err)
			}

			// Create stub TypeInfo (no field analysis yet)
			g.types[typeName] = &TypeInfo{
				Name:        typeName,
				Description: cleanedDesc,
				Deprecated:  deprecated,
				// Kind, Fields, UnderlyingType, etc. will be set in Pass 2
			}
		}
	}

	return nil
}

// extractTypeFields extracts field details for all types in a file.
// This is Pass 2 - fills in Kind, Fields, UnderlyingType, etc. for each type.
// Requires g.currentFileImports to be set before calling (by extractAllTypesFromGo).
func (g *OpenAPICollector) extractTypeFields(file *ast.File) error {
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

			if err := g.processTypeSpecFields(typeSpec); err != nil {
				return err
			}
		}
	}

	return nil
}

// processTypeSpecFields extracts field information for a single type spec.
func (g *OpenAPICollector) processTypeSpecFields(typeSpec *ast.TypeSpec) error {
	typeName := typeSpec.Name.Name
	typeInfo := g.types[typeName]

	// Determine type based on the type expression and populate fields
	switch t := typeSpec.Type.(type) {
	case *ast.StructType:
		if err := g.extractStructTypeFields(typeName, t, typeInfo); err != nil {
			return fmt.Errorf("failed to extract struct fields for %s: %w", typeName, err)
		}

	case *ast.Ident:
		// Type alias to another type (e.g., type MyString string)
		if !isEnumKind(typeInfo.Kind) {
			typeInfo.Kind = TypeKindAlias
		}

		underlyingType, refs, err := g.analyzeGoType(t)
		if err != nil {
			return fmt.Errorf("failed to analyze underlying type for alias %s: %w", typeName, err)
		}

		typeInfo.UnderlyingType = &underlyingType
		typeInfo.References = refs

	case *ast.ArrayType, *ast.MapType:
		// Arrays and maps as top-level types are treated as aliases
		if !isEnumKind(typeInfo.Kind) {
			typeInfo.Kind = TypeKindAlias
		}

		underlyingType, refs, err := g.analyzeGoType(typeSpec.Type)
		if err != nil {
			return fmt.Errorf("failed to analyze underlying type for alias %s: %w", typeName, err)
		}

		typeInfo.UnderlyingType = &underlyingType
		typeInfo.References = refs

	default:
		return fmt.Errorf("unsupported type %s: %T (please use struct, type alias, or basic types)", typeName, typeSpec.Type)
	}

	return nil
}

// extractStructTypeFields extracts struct fields and populates the TypeInfo.
// Wrapper for extractStructType that matches the error-only return signature.
func (g *OpenAPICollector) extractStructTypeFields(name string, structType *ast.StructType, typeInfo *TypeInfo) error {
	_, err := g.extractStructType(name, structType, typeInfo)

	return err
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

	displayType, err := generateDisplayType(fieldType)
	if err != nil {
		return FieldInfo{}, nil, fmt.Errorf("failed to generate display type for field %s.%s: %w", parentName, fieldName, err)
	}

	fieldInfo := FieldInfo{
		Name:        tagInfo.name,
		DisplayType: displayType,
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
	// Extract key type for later validation (OpenAPI/JSON validation happens in openapi.go)
	keyType, keyRefs, err := g.analyzeGoType(t.Key)
	if err != nil {
		return FieldType{}, nil, fmt.Errorf("failed to analyze map key type: %w", err)
	}

	valueType, valueRefs, err := g.analyzeGoType(t.Value)
	if err != nil {
		return FieldType{}, nil, err
	}

	// Combine references
	allRefs := keyRefs
	allRefs = append(allRefs, valueRefs...)

	return FieldType{
		Kind:                 FieldKindObject,
		Type:                 "object",
		MapKeyType:           &keyType,
		AdditionalProperties: &valueType,
	}, allRefs, nil
}

// analyzeSelectorType handles external types (e.g., time.Time, types.URL).
func (g *OpenAPICollector) analyzeSelectorType(t *ast.SelectorExpr) (FieldType, []string, error) {
	pkgIdent, ok := t.X.(*ast.Ident)
	if !ok {
		return FieldType{}, nil, fmt.Errorf("unsupported selector expression with base type %T - expected package.Type format", t.X)
	}

	// Get the package alias as it appears in the code
	pkgAlias := pkgIdent.Name
	typeName := t.Sel.Name

	// Resolve the alias to the full import path
	importPath, exists := g.currentFileImports[pkgAlias]
	if !exists {
		return FieldType{}, nil, fmt.Errorf("package alias %s not found in imports - cannot resolve external type %s.%s", pkgAlias, pkgAlias, typeName)
	}

	// Build the full type key using import path
	fullTypeKey := importPath + "." + typeName

	// Look up the type format using the full import path
	format, exists := g.externalTypeFormats[fullTypeKey]
	if !exists {
		return FieldType{}, nil, fmt.Errorf("unknown external type %s.%s (resolved to %s) - please add it to externalTypeFormats map in NewOpenAPICollector using the full import path as the key", pkgAlias, typeName, fullTypeKey)
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
func generateDisplayType(ft FieldType) (string, error) {
	switch ft.Kind {
	case FieldKindReference, FieldKindEnum:
		return ft.Type, nil
	case FieldKindPrimitive:
		caser := cases.Title(language.English)

		return caser.String(ft.Type), nil

	case FieldKindArray:
		if ft.ItemsType != nil {
			itemDisplay, err := generateDisplayType(*ft.ItemsType)
			if err != nil {
				return "", err
			}

			return itemDisplay + "[]", nil
		}

		return "Array", nil

	case FieldKindObject:
		return "Object", nil

	default:
		return "", fmt.Errorf("unexpected field kind: %s, should have been caught by type analysis", ft.Kind)
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
