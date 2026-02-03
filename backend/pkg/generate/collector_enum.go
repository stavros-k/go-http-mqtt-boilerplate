package generate

// This file handles enum extraction and processing from Go AST.

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"log/slog"
	"strconv"
)

// Sentinel errors for enum processing.
var (
	ErrEmptyConstBlock = errors.New("empty const block")
	ErrNoEnumType      = errors.New("no enum type found in const block")
)

// isEnumKind checks if the given kind represents an enum type.
func isEnumKind(kind string) bool {
	return kind == TypeKindStringEnum || kind == TypeKindNumberEnum
}

// extractConstDeclarations extracts enum values from const blocks in a single AST file.
func (g *OpenAPICollector) extractConstDeclarations(file *ast.File) []error {
	var errs []error

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}

		// Try to extract enum from this const block
		err := g.extractEnumsFromConstBlock(genDecl)
		if err == nil {
			continue
		}

		// Skip non-enum const blocks silently
		if errors.Is(err, ErrEmptyConstBlock) || errors.Is(err, ErrNoEnumType) {
			continue
		}

		// All other errors are real problems
		errs = append(errs, fmt.Errorf("failed to process const block: %w", err))
	}

	return errs
}

// extractEnumsFromConstBlock extracts enum values from a const block.
func (g *OpenAPICollector) extractEnumsFromConstBlock(constDecl *ast.GenDecl) error {
	if len(constDecl.Specs) == 0 {
		return ErrEmptyConstBlock
	}

	var (
		enumTypeName string
		enumValues   []EnumValue
	)

	for _, spec := range constDecl.Specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}

		// Check if all names in this spec are exported
		allExported := true

		for _, name := range valueSpec.Names {
			if !name.IsExported() {
				allExported = false

				break
			}
		}

		// If we've identified this as an enum block, all constants must be exported
		if enumTypeName != "" && !allExported {
			return fmt.Errorf("enum %s: const blocks must not contain unexported constants", enumTypeName)
		}

		// Skip const specs with no exported names (only before enum type is established)
		if !allExported {
			continue
		}

		// All exported constants must have explicit type declaration
		if valueSpec.Type == nil {
			if enumTypeName == "" {
				// First exported const without type = not an enum block
				return ErrNoEnumType
			}

			return fmt.Errorf("enum %s: all exported constants in enum const block must have explicit type declaration", enumTypeName)
		}

		ident, ok := valueSpec.Type.(*ast.Ident)
		if !ok {
			if enumTypeName != "" {
				return fmt.Errorf("enum %s: const type must be a simple identifier, got %T", enumTypeName, valueSpec.Type)
			}

			return fmt.Errorf("const type must be a simple identifier, got %T", valueSpec.Type)
		}

		// First exported const establishes the enum type
		if enumTypeName == "" {
			enumTypeName = ident.Name
		} else if ident.Name != enumTypeName {
			// Subsequent consts must match the established type
			return fmt.Errorf("mixed enum types in const block: expected %s, got %s", enumTypeName, ident.Name)
		}

		// Process values from this spec
		values, err := g.processValueSpec(valueSpec, enumTypeName)
		if err != nil {
			return err
		}

		enumValues = append(enumValues, values...)
	}

	// storeEnumType validates we have values and stores the enum
	return g.storeEnumType(enumTypeName, enumValues, constDecl)
}

// processValueSpec processes a single const value spec and extracts enum values.
func (g *OpenAPICollector) processValueSpec(valueSpec *ast.ValueSpec, enumTypeName string) ([]EnumValue, error) {
	var values []EnumValue

	for i, name := range valueSpec.Names {
		if !name.IsExported() {
			continue
		}

		enumValue, err := g.processEnumValue(valueSpec, i, name, enumTypeName)
		if err != nil {
			return nil, err
		}

		values = append(values, enumValue)
	}

	return values, nil
}

// processEnumValue processes a single enum constant value and returns the EnumValue.
// The index parameter maps the const name to its corresponding value in valueSpec.Values.
// For example, in `const (Foo = "foo"; Bar = "bar")`, index 0 maps Foo to "foo".
func (g *OpenAPICollector) processEnumValue(valueSpec *ast.ValueSpec, index int, name *ast.Ident, enumTypeName string) (EnumValue, error) {
	if index >= len(valueSpec.Values) {
		return EnumValue{}, fmt.Errorf("enum constant %s.%s is missing a value", enumTypeName, name.Name)
	}

	basicLit, ok := valueSpec.Values[index].(*ast.BasicLit)
	if !ok {
		return EnumValue{}, fmt.Errorf("enum constant %s.%s must have a literal value, got %T", enumTypeName, name.Name, valueSpec.Values[index])
	}

	var value any

	//nolint:exhaustive // Only STRING and INT literals are valid for enum constants
	switch basicLit.Kind {
	case token.STRING:
		strVal, err := strconv.Unquote(basicLit.Value)
		if err != nil {
			return EnumValue{}, fmt.Errorf("enum constant %s.%s has invalid string value %s: %w", enumTypeName, name.Name, basicLit.Value, err)
		}

		value = strVal

	case token.INT:
		intVal, err := strconv.ParseInt(basicLit.Value, 10, 64)
		if err != nil {
			return EnumValue{}, fmt.Errorf("enum constant %s.%s has invalid integer value %s: %w", enumTypeName, name.Name, basicLit.Value, err)
		}

		value = intVal

	default:
		return EnumValue{}, fmt.Errorf("enum constant %s.%s must be a string or integer, got %v", enumTypeName, name.Name, basicLit.Kind)
	}

	// Extract documentation
	desc := ""
	if valueSpec.Doc != nil {
		desc = g.extractCommentsFromDoc(valueSpec.Doc)
	} else if valueSpec.Comment != nil {
		desc = g.extractCommentsFromDoc(valueSpec.Comment)
	}

	deprecated, cleanedDesc, err := g.parseDeprecation(desc)
	if err != nil {
		return EnumValue{}, fmt.Errorf("failed to parse deprecation for enum value %s.%s: %w", enumTypeName, name.Name, err)
	}

	return EnumValue{
		Value:       value,
		Description: cleanedDesc,
		Deprecated:  deprecated,
	}, nil
}

// storeEnumType stores or updates an enum type in the types map.
func (g *OpenAPICollector) storeEnumType(enumTypeName string, enumValues []EnumValue, constDecl *ast.GenDecl) error {
	// Validate we have at least one enum value
	if len(enumValues) == 0 {
		return fmt.Errorf("enum %s: cannot store enum type with no values", enumTypeName)
	}

	// Store the const block AST node for later Go source generation
	g.constASTs[enumTypeName] = constDecl

	// Determine enum kind from first value
	// Note: All values are guaranteed to have the same type because:
	// 1. extractEnumsFromConstBlock ensures all consts have the same type name
	// 2. Go's type system enforces type compatibility
	var enumKind string

	switch enumValues[0].Value.(type) {
	case int64:
		enumKind = TypeKindNumberEnum
	case string:
		enumKind = TypeKindStringEnum
	default:
		return fmt.Errorf("enum %s: unsupported enum value type %T", enumTypeName, enumValues[0].Value)
	}

	existingType, exists := g.types[enumTypeName]
	if exists {
		// Update existing type to be an enum
		existingType.Kind = enumKind
		existingType.EnumValues = enumValues
		g.l.Debug("Updated type to enum", slog.String("name", enumTypeName), slog.String("kind", enumKind), slog.Int("valueCount", len(enumValues)))
	} else {
		// Create new enum type
		g.types[enumTypeName] = &TypeInfo{
			Name:       enumTypeName,
			Kind:       enumKind,
			EnumValues: enumValues,
		}
		g.l.Debug("Created new enum type", slog.String("name", enumTypeName), slog.String("kind", enumKind), slog.Int("valueCount", len(enumValues)))
	}

	return nil
}
