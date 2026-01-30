package generate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"http-mqtt-boilerplate/backend/pkg/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

const (
	// OpenAPIVersion is the OpenAPI specification version used for generated specs.
	OpenAPIVersion = "3.0.3"
)

// isPrimitiveType checks if a type name represents a valid OpenAPI primitive type.
// Note: "array" and "object" are excluded as they require additional schema information.
func isPrimitiveType(typeName string) bool {
	switch typeName {
	case "string", "number", "integer", "boolean":
		return true
	default:
		return false
	}
}

// toOpenAPISchema converts extracted type metadata to an OpenAPI schema.
func toOpenAPISchema(typeInfo *TypeInfo) (*openapi3.Schema, error) {
	switch {
	case typeInfo.Kind == TypeKindObject:
		return buildObjectSchema(typeInfo)
	case isEnumKind(typeInfo.Kind):
		return buildEnumSchema(typeInfo)
	case typeInfo.Kind == TypeKindAlias:
		return buildAliasSchema(typeInfo)
	default:
		return nil, fmt.Errorf("unsupported type kind: %s", typeInfo.Kind)
	}
}

// buildObjectSchema creates an OpenAPI object schema.
func buildObjectSchema(typeInfo *TypeInfo) (*openapi3.Schema, error) {
	schema := &openapi3.Schema{
		Type:        &openapi3.Types{"object"},
		Description: typeInfo.Description,
		Deprecated:  typeInfo.Deprecated != "",
		Properties:  make(openapi3.Schemas),
		Required:    []string{},
	}

	// Structured objects should not allow additional properties
	schema.AdditionalProperties = openapi3.AdditionalProperties{
		Has: utils.Ptr(false),
	}

	for _, field := range typeInfo.Fields {
		fieldSchema, err := buildFieldSchema(field)
		if err != nil {
			return nil, fmt.Errorf("failed to build schema for field %s: %w", field.Name, err)
		}

		schema.Properties[field.Name] = fieldSchema

		if field.TypeInfo.Required {
			schema.Required = append(schema.Required, field.Name)
		}
	}

	return schema, nil
}

// buildFieldSchema creates an OpenAPI schema for a field.
func buildFieldSchema(field FieldInfo) (*openapi3.SchemaRef, error) {
	schema, err := buildSchemaFromFieldType(field.TypeInfo, field.Description)
	if err != nil {
		return nil, err
	}

	// Apply field-level deprecated metadata
	return applyDeprecated(schema, field.Deprecated != "")
}

// applyNullable sets the Nullable field on a schema if needed.
// For inline schemas (Value != nil), sets nullable directly.
// For $ref schemas (Value == nil, Ref != ""), wraps with allOf in OpenAPI 3.0.
// FIXME: OpenAPI 3.1 uses JSON Schema which supports type: 'null'.
// When upgrading to OpenAPI 3.1, replace allOf+nullable pattern with anyOf containing type: 'null'.
func applyNullable(schemaRef *openapi3.SchemaRef, nullable bool) (*openapi3.SchemaRef, error) {
	switch {
	case !nullable:
		return schemaRef, nil
	case schemaRef.Value != nil:
		// Inline schema - set nullable directly
		schemaRef.Value.Nullable = true

		return schemaRef, nil
	case schemaRef.Ref != "":
		// OpenAPI 3.0: Reference schema - must wrap with allOf
		return &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf:    []*openapi3.SchemaRef{schemaRef},
				Nullable: true,
			},
		}, nil
		// OpenAPI 3.1: Use anyOf with null type instead
		// nullType := &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"null"}}}
		// return &openapi3.SchemaRef{Value: &openapi3.Schema{AnyOf: []*openapi3.SchemaRef{schemaRef, nullType}}}
	default:
		return nil, errors.New("invalid schemaRef: both Value and Ref are empty")
	}
}

// applyDeprecated sets the Deprecated field on a schema if needed.
// For inline schemas (Value != nil), sets deprecated directly.
// For $ref schemas (Value == nil, Ref != ""), wraps with allOf in OpenAPI 3.0.
// FIXME: OpenAPI 3.1 allows deprecated directly on $ref without wrapping.
// When upgrading to OpenAPI 3.1, set deprecated as a sibling to $ref instead of wrapping.
func applyDeprecated(schemaRef *openapi3.SchemaRef, deprecated bool) (*openapi3.SchemaRef, error) {
	switch {
	case !deprecated:
		return schemaRef, nil
	case schemaRef.Value != nil:
		// Inline schema - set deprecated directly
		schemaRef.Value.Deprecated = true

		return schemaRef, nil
	case schemaRef.Ref != "":
		// OpenAPI 3.0: Reference schema - must wrap with allOf
		return &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				AllOf:      []*openapi3.SchemaRef{schemaRef},
				Deprecated: true,
			},
		}, nil
		// OpenAPI 3.1: Can set deprecated as sibling to $ref
		// Example approach:
		// schemaRef.Value = &openapi3.Schema{Deprecated: true}
		// // Keep schemaRef.Ref as is
		// return schemaRef
	default:
		return nil, errors.New("invalid schemaRef: both Value and Ref are empty")
	}
}

// buildPrimitiveSchemaFromFieldType builds a schema for primitive types.
func buildPrimitiveSchemaFromFieldType(ft FieldType, description string) (*openapi3.SchemaRef, error) {
	schema := &openapi3.Schema{
		Type:        &openapi3.Types{ft.Type},
		Description: description,
	}
	if ft.Format != "" {
		schema.Format = ft.Format
	}

	schemaRef := &openapi3.SchemaRef{Value: schema}

	return applyNullable(schemaRef, ft.Nullable)
}

// buildArraySchemaFromFieldType builds a schema for array types.
func buildArraySchemaFromFieldType(ft FieldType, description string) (*openapi3.SchemaRef, error) {
	var itemSchema *openapi3.SchemaRef

	if ft.ItemsType != nil {
		var err error

		itemSchema, err = buildSchemaFromFieldType(*ft.ItemsType, "")
		if err != nil {
			return nil, err
		}
	}

	schema := &openapi3.Schema{
		Type:        &openapi3.Types{"array"},
		Items:       itemSchema,
		Description: description,
	}

	schemaRef := &openapi3.SchemaRef{Value: schema}

	return applyNullable(schemaRef, ft.Nullable)
}

// buildReferenceSchemaFromFieldType builds a schema reference for type references and enums.
func buildReferenceSchemaFromFieldType(ft FieldType) (*openapi3.SchemaRef, error) {
	ref := createSchemaRef(ft.Type)

	return applyNullable(ref, ft.Nullable)
}

// buildObjectSchemaFromFieldType builds a schema for object/map types.
func buildObjectSchemaFromFieldType(ft FieldType, description string) (*openapi3.SchemaRef, error) {
	schema := &openapi3.Schema{
		Type:        &openapi3.Types{"object"},
		Description: description,
	}

	schema.AdditionalProperties = openapi3.AdditionalProperties{Has: utils.Ptr(false)}

	// Handle additionalProperties for map types
	if ft.AdditionalProperties != nil {
		additionalPropsSchema, err := buildSchemaFromFieldType(*ft.AdditionalProperties, "")
		if err != nil {
			return nil, fmt.Errorf("failed to build additionalProperties schema: %w", err)
		}

		schema.AdditionalProperties = openapi3.AdditionalProperties{
			Schema: additionalPropsSchema,
		}
	}

	schemaRef := &openapi3.SchemaRef{Value: schema}

	return applyNullable(schemaRef, ft.Nullable)
}

// buildSchemaFromFieldType converts a FieldType to an OpenAPI schema.
func buildSchemaFromFieldType(ft FieldType, description string) (*openapi3.SchemaRef, error) {
	switch ft.Kind {
	case FieldKindPrimitive:
		return buildPrimitiveSchemaFromFieldType(ft, description)

	case FieldKindArray:
		return buildArraySchemaFromFieldType(ft, description)

	case FieldKindReference, FieldKindEnum:
		return buildReferenceSchemaFromFieldType(ft)

	case FieldKindObject:
		return buildObjectSchemaFromFieldType(ft, description)

	default:
		// Unhandled type kind - fail with error
		return nil, fmt.Errorf("unhandled field kind: %s", ft.Kind)
	}
}

// formatEnumValueDescription formats a single enum value for documentation.
func formatEnumValueDescription(ev EnumValue) string {
	switch {
	case ev.Deprecated != "":
		result := fmt.Sprintf("- `%v`: **[DEPRECATED]** ", ev.Value)
		result += ev.Deprecated

		if ev.Description != "" {
			result += " - " + ev.Description
		}

		return result + "\n"
	case ev.Description != "":
		return fmt.Sprintf("- `%v`: %s\n", ev.Value, ev.Description)
	default:
		return fmt.Sprintf("- `%v`\n", ev.Value)
	}
}

// buildEnumSchema creates an OpenAPI enum schema.
// FIXME: Once upstream supports OpenAPI 3.1, switch to using oneOf with const.
func buildEnumSchema(typeInfo *TypeInfo) (*openapi3.Schema, error) {
	values := make([]any, len(typeInfo.EnumValues))

	var enumDesc strings.Builder

	if typeInfo.Description != "" {
		enumDesc.WriteString(typeInfo.Description)
		enumDesc.WriteString("\n\n")
	}

	enumDesc.WriteString("Possible values:\n")

	for i, ev := range typeInfo.EnumValues {
		values[i] = ev.Value
		enumDesc.WriteString(formatEnumValueDescription(ev))
	}

	// Determine OpenAPI type based on enum kind
	var schemaType string

	switch typeInfo.Kind {
	case TypeKindStringEnum:
		schemaType = "string"
	case TypeKindNumberEnum:
		schemaType = "integer"
	default:
		return nil, fmt.Errorf("unsupported enum kind: %s", typeInfo.Kind)
	}

	return &openapi3.Schema{
		Type:        &openapi3.Types{schemaType},
		Enum:        values,
		Description: enumDesc.String(),
		Deprecated:  typeInfo.Deprecated != "",
	}, nil
}

// buildAliasSchema creates an OpenAPI schema for alias types by resolving to the underlying type.
func buildAliasSchema(typeInfo *TypeInfo) (*openapi3.Schema, error) {
	if typeInfo.UnderlyingType == nil {
		return nil, fmt.Errorf("alias type %s has no underlying type information", typeInfo.Name)
	}

	// Build schema from the underlying type
	schemaRef, err := buildSchemaFromFieldType(*typeInfo.UnderlyingType, typeInfo.Description)
	if err != nil {
		return nil, fmt.Errorf("failed to build schema for alias %s: %w", typeInfo.Name, err)
	}

	var schema *openapi3.Schema

	// Handle both inline schemas and reference schemas
	switch {
	case schemaRef.Value != nil:
		// Inline schema - use directly
		schema = schemaRef.Value

		// Apply type-level deprecation if set
		if typeInfo.Deprecated != "" {
			schema.Deprecated = true
		}
	case schemaRef.Ref != "":
		// Reference schema - wrap in allOf to allow adding metadata
		schema = &openapi3.Schema{
			AllOf:       []*openapi3.SchemaRef{schemaRef},
			Description: typeInfo.Description,
			Deprecated:  typeInfo.Deprecated != "",
		}
	default:
		return nil, fmt.Errorf("alias type %s resolved to invalid schema (both Value and Ref are empty)", typeInfo.Name)
	}

	return schema, nil
}

// buildComponentSchemas builds OpenAPI component schemas from HTTP-related types only.
// Types are marked as HTTP-related during RegisterRoute.
func buildComponentSchemas(doc *APIDocumentation) (openapi3.Schemas, error) {
	schemas := make(openapi3.Schemas)

	// Build schemas only for types marked as used by HTTP
	for name, typeInfo := range doc.Types {
		if !typeInfo.UsedByHTTP {
			continue
		}

		schema, err := toOpenAPISchema(typeInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to build schema for %s: %w", name, err)
		}

		schemas[name] = &openapi3.SchemaRef{Value: schema}
	}

	return schemas, nil
}

// generateOpenAPISpec generates a complete OpenAPI specification from documentation.
func generateOpenAPISpec(doc *APIDocumentation) (*openapi3.T, error) {
	spec := &openapi3.T{
		OpenAPI:    OpenAPIVersion,
		Info:       &openapi3.Info{},
		Paths:      openapi3.NewPaths(),
		Components: &openapi3.Components{Schemas: make(openapi3.Schemas)},
	}

	// Build all component schemas
	schemas, err := buildComponentSchemas(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to build component schemas: %w", err)
	}

	spec.Components.Schemas = schemas

	// Build paths from http_operations
	pathItems := make(map[string]*openapi3.PathItem)

	for _, route := range doc.HTTPOperations {
		// Get or create path item for this path
		pathItem, exists := pathItems[route.Path]
		if !exists {
			pathItem = &openapi3.PathItem{}
			pathItems[route.Path] = pathItem
		}

		op, err := buildOperation(route, doc.Types)
		if err != nil {
			return nil, fmt.Errorf("failed to build operation %s %s: %w", route.Method, route.Path, err)
		}

		// Add operation to path item
		switch route.Method {
		case http.MethodGet:
			pathItem.Get = op
		case http.MethodPost:
			pathItem.Post = op
		case http.MethodPut:
			pathItem.Put = op
		case http.MethodPatch:
			pathItem.Patch = op
		case http.MethodDelete:
			pathItem.Delete = op
		default:
			return nil, fmt.Errorf("unsupported HTTP method: %s", route.Method)
		}
	}

	// Add all path items to spec
	for path, pathItem := range pathItems {
		spec.Paths.Set(path, pathItem)
	}

	return spec, nil
}

// buildOperation builds an OpenAPI operation from RouteInfo.
func buildOperation(route *RouteInfo, types map[string]*TypeInfo) (*openapi3.Operation, error) {
	op := &openapi3.Operation{
		OperationID: route.OperationID,
		Summary:     route.Summary,
		Description: route.Description,
		Tags:        []string{route.Group},
		Deprecated:  route.Deprecated != "",
		Responses:   &openapi3.Responses{},
	}

	// Add parameters
	for _, param := range route.Parameters {
		p := &openapi3.Parameter{
			Name:        param.Name,
			In:          param.In,
			Required:    param.Required,
			Description: param.Description,
		}

		// Build schema for parameter type
		if typeInfo, ok := types[param.TypeName]; ok {
			schema, err := toOpenAPISchema(typeInfo)
			if err != nil {
				return nil, fmt.Errorf("failed to build schema for parameter %s: %w", param.Name, err)
			}

			p.Schema = &openapi3.SchemaRef{Value: schema}
		} else {
			// Validate that it's a known primitive type before creating inline schema
			if !isPrimitiveType(param.TypeName) {
				return nil, fmt.Errorf("parameter %s has unregistered type %s (not found in types map and not a valid primitive type)", param.Name, param.TypeName)
			}

			// Primitive type - create inline schema
			p.Schema = &openapi3.SchemaRef{
				Value: &openapi3.Schema{Type: &openapi3.Types{param.TypeName}},
			}
		}

		op.Parameters = append(op.Parameters, &openapi3.ParameterRef{Value: p})
	}

	// Add request body
	if route.Request != nil {
		op.RequestBody = &openapi3.RequestBodyRef{
			Value: &openapi3.RequestBody{
				Required:    true,
				Description: route.Request.Description,
				Content:     createJSONContent(route.Request.TypeName, route.Request.Examples),
			},
		}
	}

	// Add responses
	for statusCode, resp := range route.Responses {
		statusStr := strconv.Itoa(statusCode)
		response := &openapi3.Response{Description: &resp.Description}

		if resp.TypeName != "" {
			response.Content = createJSONContent(resp.TypeName, resp.Examples)
		}

		op.Responses.Set(statusStr, &openapi3.ResponseRef{Value: response})
	}

	return op, nil
}

// createJSONContent creates OpenAPI content for application/json with given type and examples.
func createJSONContent(typeName string, examples map[string]any) openapi3.Content {
	return openapi3.Content{
		"application/json": &openapi3.MediaType{
			Schema:   createSchemaRef(typeName),
			Examples: convertExamplesToOpenAPI(examples),
		},
	}
}

// createSchemaRef creates a schema reference for the given type name.
func createSchemaRef(typeName string) *openapi3.SchemaRef {
	return &openapi3.SchemaRef{
		Ref: "#/components/schemas/" + typeName,
	}
}

// convertExamplesToOpenAPI converts examples map to OpenAPI format.
func convertExamplesToOpenAPI(examples map[string]any) openapi3.Examples {
	if len(examples) == 0 {
		return nil
	}

	result := make(openapi3.Examples)
	for name, value := range examples {
		result[name] = &openapi3.ExampleRef{Value: &openapi3.Example{Value: value}}
	}

	return result
}

// schemaToJSONString converts an OpenAPI schema to a stringified JSON representation.
func schemaToJSONString(schema *openapi3.Schema) (string, error) {
	if schema == nil {
		return "", nil
	}

	jsonBytes, err := schema.MarshalJSON()
	if err != nil {
		return "", fmt.Errorf("failed to marshal schema to JSON: %w", err)
	}

	var dest bytes.Buffer
	if err := json.Indent(&dest, jsonBytes, "", "  "); err != nil {
		return "", fmt.Errorf("failed to indent JSON: %w", err)
	}

	return dest.String(), nil
}
