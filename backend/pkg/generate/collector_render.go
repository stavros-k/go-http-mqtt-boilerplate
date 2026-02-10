package generate

// This file handles generation of type representations (Go, TypeScript, JSON Schema).

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"log/slog"
	"strings"
)

// generateTypesRepresentations generates Go, TypeScript, and JSON Schema representations for all types as a post-processing step.
func (g *OpenAPICollector) generateTypesRepresentations() error {
	g.l.Debug("Generating all representations for all types", slog.Int("typeCount", len(g.types)))

	for name, typeInfo := range g.types {
		// Go Representation
		goSource, err := g.generateGoSource(typeInfo)
		if err != nil {
			return fmt.Errorf("failed to generate Go representation for %s: %w", name, err)
		}

		typeInfo.Representations.Go = goSource

		// TypeScript Representation
		tsSource, err := g.serializeTSNode(name)
		if err != nil {
			return fmt.Errorf("failed to serialize TS representation for %s: %w", name, err)
		}

		typeInfo.Representations.TS = tsSource

		// JSON Schema Representation
		schema, err := toOpenAPISchema(typeInfo)
		if err != nil {
			return fmt.Errorf("failed to generate JSON schema for type %s: %w", name, err)
		}

		jsonSchema, err := schemaToJSONString(schema)
		if err != nil {
			return fmt.Errorf("failed to serialize JSON schema for type %s: %w", name, err)
		}

		typeInfo.Representations.JSONSchema = jsonSchema

		// YAML Schema Representation (same schema, different format)
		yamlSchema, err := schemaToYAMLString(schema)
		if err != nil {
			return fmt.Errorf("failed to serialize YAML schema for type %s: %w", name, err)
		}

		typeInfo.Representations.YAMLSchema = yamlSchema
	}

	g.l.Debug("All representations generated successfully")

	return nil
}

// serializeTSNode serializes a TypeScript AST node to a string.
func (g *OpenAPICollector) serializeTSNode(name string) (string, error) {
	node, exists := g.tsParser.ts.Node(name)
	if !exists {
		return "", fmt.Errorf("type %s not found in TypeScript AST", name)
	}

	tsNode, err := g.tsParser.vm.ToTypescriptNode(node)
	if err != nil {
		return "", fmt.Errorf("failed to convert node to TypeScript node for type %s: %w", name, err)
	}

	serialized, err := g.tsParser.vm.SerializeToTypescript(tsNode)
	if err != nil {
		return "", fmt.Errorf("failed to serialize TypeScript node for type %s: %w", name, err)
	}

	// Filter unwanted lines while preserving spacing
	skipPrefixes := []string{"// From", "*nolint:"}

	return cleanupSourceLines(serialized, skipPrefixes), nil
}

// generateGoSource generates Go source code for a type using the parsed AST.
func (g *OpenAPICollector) generateGoSource(typeInfo *TypeInfo) (string, error) {
	var buf bytes.Buffer

	// Look up the AST nodes by type name
	typeDecl, exists := g.typeASTs[typeInfo.Name]
	if !exists {
		return "", fmt.Errorf("no type declaration AST found for type %s", typeInfo.Name)
	}

	if err := g.printASTNode(&buf, typeDecl, typeInfo.Name, "type declaration"); err != nil {
		return "", err
	}

	buf.WriteString("\n")

	if isEnumKind(typeInfo.Kind) {
		constDecl, exists := g.constASTs[typeInfo.Name]
		if !exists {
			return "", fmt.Errorf("no const declaration AST found for enum type %s", typeInfo.Name)
		}

		if err := g.printASTNode(&buf, constDecl, typeInfo.Name, "const declaration"); err != nil {
			return "", err
		}

		buf.WriteString("\n")
	}

	// Filter unwanted lines
	skipPrefixes := []string{"//nolint:"}
	cleaned := cleanupSourceLines(buf.String(), skipPrefixes)

	formatted, err := format.Source([]byte(cleaned))
	if err != nil {
		return "", fmt.Errorf("failed to format Go source for type %s: %w", typeInfo.Name, err)
	}

	return string(formatted), nil
}

// printASTNode prints an AST node to a buffer using go/printer.
func (g *OpenAPICollector) printASTNode(buf *bytes.Buffer, node *ast.GenDecl, typeName, nodeType string) error {
	if node == nil {
		return nil
	}

	if err := printer.Fprint(buf, g.goParser.fset, node); err != nil {
		return fmt.Errorf("failed to print %s for %s: %w", nodeType, typeName, err)
	}

	return nil
}

// cleanupSourceLines filters out unwanted lines from generated source code.
// Assumes input spacing is already correct and preserves it.
func cleanupSourceLines(input string, skipPrefixes []string) string {
	var result strings.Builder

	for line := range strings.SplitSeq(input, "\n") {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and empty comment lines
		if trimmed == "" || trimmed == "//" || trimmed == "*" {
			continue
		}

		// Skip lines matching any of the specified prefixes (check on trimmed)
		shouldSkip := false

		for _, prefix := range skipPrefixes {
			if strings.HasPrefix(trimmed, prefix) {
				shouldSkip = true

				break
			}
		}

		if shouldSkip {
			continue
		}

		// Output original line with spacing intact
		result.WriteString(line + "\n")
	}

	return strings.TrimSpace(result.String()) + "\n"
}
