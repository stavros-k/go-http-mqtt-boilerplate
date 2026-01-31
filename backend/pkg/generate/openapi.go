package generate

import (
	"encoding/json"
	"errors"
	"fmt"
	"http-mqtt-boilerplate/backend/pkg/utils"
	"net/http"
	"slices"
	"strconv"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"go.yaml.in/yaml/v4"
)

const (
	// OpenAPIVersion is the OpenAPI specification version used for generated specs.
	OpenAPIVersion = "3.1.0"

	dynamicValueA = 0
	dynamicValueB = 1
)

// newDynamicValueA creates a DynamicValue holding a schema (type A).
// Examples:
//   - items: { type: string } for arrays
//   - additionalProperties: { type: string } for maps
func newDynamicValueA[A, B any](a A) *base.DynamicValue[A, B] {
	return &base.DynamicValue[A, B]{
		N: dynamicValueA,
		A: a,
	}
}

// newDynamicValueB creates a DynamicValue holding a boolean (type B).
// Examples:
//   - additionalProperties: false (no extra properties allowed)
//   - additionalProperties: true (any extra properties allowed)
func newDynamicValueB[A, B any](b B) *base.DynamicValue[A, B] {
	return &base.DynamicValue[A, B]{
		N: dynamicValueB,
		B: b,
	}
}

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
func toOpenAPISchema(typeInfo *TypeInfo) (*base.Schema, error) {
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
func buildObjectSchema(typeInfo *TypeInfo) (*base.Schema, error) {
	schema := &base.Schema{
		Type:        []string{"object"},
		Description: typeInfo.Description,
		Deprecated:  utils.Ptr(typeInfo.Deprecated != ""),
		Properties:  orderedmap.New[string, *base.SchemaProxy](),
		Required:    []string{},
	}

	// Structured objects should not allow additional properties
	schema.AdditionalProperties = newDynamicValueB[*base.SchemaProxy](false)

	for _, field := range typeInfo.Fields {
		fieldSchema, err := buildFieldSchema(field)
		if err != nil {
			return nil, fmt.Errorf("failed to build schema for field %s: %w", field.Name, err)
		}

		schema.Properties.Set(field.Name, fieldSchema)

		if field.TypeInfo.Required {
			schema.Required = append(schema.Required, field.Name)
		}
	}

	return schema, nil
}

// buildFieldSchema creates an OpenAPI schema for a field.
func buildFieldSchema(field FieldInfo) (*base.SchemaProxy, error) {
	schema, err := buildSchemaFromFieldType(field.TypeInfo, field.Description)
	if err != nil {
		return nil, err
	}

	// Apply field-level deprecated metadata
	return applyDeprecated(schema, field.Deprecated != "")
}

// applyNullable sets nullable for a schema in OpenAPI 3.1 style.
func applyNullable(schemaProxy *base.SchemaProxy, nullable bool) (*base.SchemaProxy, error) {
	if !nullable {
		return schemaProxy, nil
	}

	// If it's a reference, we need to use oneOf since we can't modify the ref directly
	// oneOf is semantically correct: value is exactly one of [reference type, null]
	if schemaProxy.IsReference() {
		nullSchema := base.CreateSchemaProxy(&base.Schema{
			Type: []string{"null"},
		})

		wrapperSchema := &base.Schema{
			OneOf: []*base.SchemaProxy{schemaProxy, nullSchema},
		}

		return base.CreateSchemaProxy(wrapperSchema), nil
	}

	// Build the schema to access its properties
	schema, err := schemaProxy.BuildSchema()
	if err != nil {
		return nil, fmt.Errorf("failed to build schema for nullable: %w", err)
	}

	// All schemas we generate should have an explicit type
	// If we reach here without a type, it indicates a bug in schema generation
	if schema == nil || len(schema.Type) == 0 {
		return nil, errors.New("nullable inline schema has no type (schema generation bug)")
	}

	// Add "null" to the type array
	if !slices.Contains(schema.Type, "null") {
		schema.Type = append(schema.Type, "null")
	}

	return base.CreateSchemaProxy(schema), nil
}

// applyDeprecated sets the Deprecated field on a schema.
func applyDeprecated(schemaProxy *base.SchemaProxy, deprecated bool) (*base.SchemaProxy, error) {
	if !deprecated {
		return schemaProxy, nil
	}

	// Build the schema to access it
	schema, err := schemaProxy.BuildSchema()
	if err != nil {
		return nil, fmt.Errorf("failed to build schema for deprecation: %w", err)
	}

	switch {
	case schema != nil:
		// Inline schema - set deprecated directly
		schema.Deprecated = utils.Ptr(true)
		return base.CreateSchemaProxy(schema), nil
	case schemaProxy.IsReference():
		// Reference schema - wrap in allOf to add deprecated metadata
		// This is a workaround: libopenapi's high-level model doesn't support
		// $ref and deprecated at the same level, even though OpenAPI 3.1 allows it
		wrapperSchema := &base.Schema{
			AllOf:      []*base.SchemaProxy{schemaProxy},
			Deprecated: utils.Ptr(true),
		}
		return base.CreateSchemaProxy(wrapperSchema), nil
	default:
		return nil, errors.New("invalid schemaProxy: both Schema and Reference are empty")
	}
}

// buildPrimitiveSchemaFromFieldType builds a schema for primitive types.
func buildPrimitiveSchemaFromFieldType(ft FieldType, description string) (*base.SchemaProxy, error) {
	schema := &base.Schema{
		Type:        []string{ft.Type},
		Description: description,
	}
	if ft.Format != "" {
		schema.Format = ft.Format
	}

	schemaProxy := base.CreateSchemaProxy(schema)

	return applyNullable(schemaProxy, ft.Nullable)
}

// buildArraySchemaFromFieldType builds a schema for array types.
func buildArraySchemaFromFieldType(ft FieldType, description string) (*base.SchemaProxy, error) {
	// Arrays must have an items type per OpenAPI spec
	if ft.ItemsType == nil {
		return nil, errors.New("array type requires items schema: ItemsType is nil")
	}

	itemSchema, err := buildSchemaFromFieldType(*ft.ItemsType, "")
	if err != nil {
		return nil, fmt.Errorf("failed to build items schema for array: %w", err)
	}

	if itemSchema == nil {
		return nil, errors.New("array type requires items schema: buildSchemaFromFieldType returned nil")
	}

	schema := &base.Schema{
		Type:        []string{"array"},
		Items:       newDynamicValueA[*base.SchemaProxy, bool](itemSchema),
		Description: description,
	}

	schemaProxy := base.CreateSchemaProxy(schema)

	return applyNullable(schemaProxy, ft.Nullable)
}

// buildReferenceSchemaFromFieldType builds a schema reference for type references and enums.
func buildReferenceSchemaFromFieldType(ft FieldType) (*base.SchemaProxy, error) {
	ref := createSchemaRef(ft.Type)

	return applyNullable(ref, ft.Nullable)
}

// buildObjectSchemaFromFieldType builds a schema for object/map types.
func buildObjectSchemaFromFieldType(ft FieldType, description string) (*base.SchemaProxy, error) {
	schema := &base.Schema{
		Type:        []string{"object"},
		Description: description,
	}

	schema.AdditionalProperties = newDynamicValueB[*base.SchemaProxy](false)

	// Handle additionalProperties for map types
	if ft.AdditionalProperties != nil {
		additionalPropsSchema, err := buildSchemaFromFieldType(*ft.AdditionalProperties, "")
		if err != nil {
			return nil, fmt.Errorf("failed to build additionalProperties schema: %w", err)
		}

		schema.AdditionalProperties = newDynamicValueA[*base.SchemaProxy, bool](additionalPropsSchema)
	}

	schemaProxy := base.CreateSchemaProxy(schema)

	return applyNullable(schemaProxy, ft.Nullable)
}

// buildSchemaFromFieldType converts a FieldType to an OpenAPI schema.
func buildSchemaFromFieldType(ft FieldType, description string) (*base.SchemaProxy, error) {
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

// buildEnumSchema creates an OpenAPI enum schema using oneOf with const values.
// This approach provides better structured documentation where each enum value
// can have its own description and deprecation status.
func buildEnumSchema(typeInfo *TypeInfo) (*base.Schema, error) {
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

	// Build oneOf schemas, one for each enum value
	oneOfSchemas := make([]*base.SchemaProxy, len(typeInfo.EnumValues))

	for i, ev := range typeInfo.EnumValues {
		// Convert value to yaml.Node for const
		constNode := &yaml.Node{}
		if err := constNode.Encode(ev.Value); err != nil {
			return nil, fmt.Errorf("failed to encode enum value: %w", err)
		}

		// Build schema for this enum value
		valueSchema := &base.Schema{
			Type:  []string{schemaType},
			Const: constNode,
		}

		// Add title (the enum value itself as a string)
		valueSchema.Title = fmt.Sprintf("%v", ev.Value)

		// Add description if present
		if ev.Description != "" {
			valueSchema.Description = ev.Description
		}

		// Mark as deprecated if needed
		if ev.Deprecated != "" {
			valueSchema.Deprecated = utils.Ptr(true)
			// Use deprecation reason as description if no description exists
			if valueSchema.Description == "" {
				valueSchema.Description = ev.Deprecated
			}
		}

		oneOfSchemas[i] = base.CreateSchemaProxy(valueSchema)
	}

	schema := &base.Schema{
		OneOf: oneOfSchemas,
	}

	// Add top-level description if present
	if typeInfo.Description != "" {
		schema.Description = typeInfo.Description
	}

	// Mark entire type as deprecated if needed
	if typeInfo.Deprecated != "" {
		schema.Deprecated = utils.Ptr(true)
	}

	return schema, nil
}

// buildAliasSchema creates an OpenAPI schema for alias types by resolving to the underlying type.
func buildAliasSchema(typeInfo *TypeInfo) (*base.Schema, error) {
	if typeInfo.UnderlyingType == nil {
		return nil, fmt.Errorf("alias type %s has no underlying type information", typeInfo.Name)
	}

	// Build schema from the underlying type
	schemaProxy, err := buildSchemaFromFieldType(*typeInfo.UnderlyingType, typeInfo.Description)
	if err != nil {
		return nil, fmt.Errorf("failed to build schema for alias %s: %w", typeInfo.Name, err)
	}

	// Build the actual schema
	schema, err := schemaProxy.BuildSchema()
	if err != nil {
		return nil, fmt.Errorf("failed to build schema for alias %s: %w", typeInfo.Name, err)
	}

	// Handle both inline schemas and reference schemas
	switch {
	case schema != nil:
		// Inline schema - use directly
		// Apply type-level deprecation if set
		if typeInfo.Deprecated != "" {
			schema.Deprecated = utils.Ptr(true)
		}
		return schema, nil
	case schemaProxy.IsReference():
		// Reference schema - wrap in allOf to allow adding metadata
		// This is a workaround: libopenapi's high-level model doesn't support
		// $ref and deprecated/description at the same level
		wrapperSchema := &base.Schema{
			AllOf:       []*base.SchemaProxy{schemaProxy},
			Description: typeInfo.Description,
			Deprecated:  utils.Ptr(typeInfo.Deprecated != ""),
		}
		return wrapperSchema, nil
	default:
		return nil, fmt.Errorf("alias type %s resolved to invalid schema (both Schema and Reference are empty)", typeInfo.Name)
	}
}

// buildComponentSchemas builds OpenAPI component schemas from HTTP-related types only.
// Types are marked as HTTP-related during RegisterRoute.
func buildComponentSchemas(doc *APIDocumentation) (*orderedmap.Map[string, *base.SchemaProxy], error) {
	schemas := orderedmap.New[string, *base.SchemaProxy]()

	// Build schemas only for types marked as used by HTTP
	for name, typeInfo := range doc.Types {
		if !typeInfo.UsedByHTTP {
			continue
		}

		schema, err := toOpenAPISchema(typeInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to build schema for %s: %w", name, err)
		}

		schemas.Set(name, base.CreateSchemaProxy(schema))
	}

	return schemas, nil
}

// generateOpenAPISpec generates a complete OpenAPI specification from documentation.
func generateOpenAPISpec(doc *APIDocumentation) (*v3.Document, error) {
	spec := &v3.Document{
		Version:    OpenAPIVersion,
		Info:       &base.Info{},
		Paths:      &v3.Paths{PathItems: orderedmap.New[string, *v3.PathItem]()},
		Components: &v3.Components{Schemas: orderedmap.New[string, *base.SchemaProxy]()},
	}

	// Build all component schemas
	schemas, err := buildComponentSchemas(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to build component schemas: %w", err)
	}

	spec.Components.Schemas = schemas

	// Build paths from http_operations
	for _, route := range doc.HTTPOperations {
		// Get or create path item for this path
		pathItem, exists := spec.Paths.PathItems.Get(route.Path)
		if !exists {
			pathItem = &v3.PathItem{}
			spec.Paths.PathItems.Set(route.Path, pathItem)
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

	return spec, nil
}

// buildOperation builds an OpenAPI operation from RouteInfo.
func buildOperation(route *RouteInfo, types map[string]*TypeInfo) (*v3.Operation, error) {
	op := &v3.Operation{
		OperationId: route.OperationID,
		Summary:     route.Summary,
		Description: route.Description,
		Tags:        []string{route.Group},
		Deprecated:  utils.Ptr(route.Deprecated != ""),
		Responses:   &v3.Responses{Codes: orderedmap.New[string, *v3.Response]()},
	}

	// Add parameters
	for _, param := range route.Parameters {
		p := &v3.Parameter{
			Name:        param.Name,
			In:          param.In,
			Required:    utils.Ptr(param.Required),
			Description: param.Description,
		}

		// Build schema for parameter type
		if typeInfo, ok := types[param.TypeName]; ok {
			schema, err := toOpenAPISchema(typeInfo)
			if err != nil {
				return nil, fmt.Errorf("failed to build schema for parameter %s: %w", param.Name, err)
			}

			p.Schema = base.CreateSchemaProxy(schema)
		} else {
			// Validate that it's a known primitive type before creating inline schema
			if !isPrimitiveType(param.TypeName) {
				return nil, fmt.Errorf("parameter %s has unregistered type %s (not found in types map and not a valid primitive type)", param.Name, param.TypeName)
			}

			// Primitive type - create inline schema
			p.Schema = base.CreateSchemaProxy(&base.Schema{Type: []string{param.TypeName}})
		}

		op.Parameters = append(op.Parameters, p)
	}

	// Add request body
	if route.Request != nil {
		content, err := buildJSONContent(route.Request.TypeName, route.Request.Examples, types)
		if err != nil {
			return nil, fmt.Errorf("request body: %w", err)
		}

		op.RequestBody = &v3.RequestBody{
			Required:    utils.Ptr(true),
			Description: route.Request.Description,
			Content:     content,
		}
	}

	// Add responses
	for statusCode, resp := range route.Responses {
		statusStr := strconv.Itoa(statusCode)
		response := &v3.Response{Description: resp.Description}

		if resp.TypeName != "" {
			content, err := buildJSONContent(resp.TypeName, resp.Examples, types)
			if err != nil {
				return nil, fmt.Errorf("response for status %d: %w", statusCode, err)
			}

			response.Content = content
		}

		op.Responses.Codes.Set(statusStr, response)
	}

	return op, nil
}

// createJSONContent creates OpenAPI content for application/json with given type and examples.
func createJSONContent(typeName string, examples map[string]any) (*orderedmap.Map[string, *v3.MediaType], error) {
	examplesMap, err := convertExamplesToOpenAPI(examples)
	if err != nil {
		return nil, err
	}

	content := orderedmap.New[string, *v3.MediaType]()
	content.Set("application/json", &v3.MediaType{
		Schema:   createSchemaRef(typeName),
		Examples: examplesMap,
	})
	return content, nil
}

// buildJSONContent creates OpenAPI content for application/json with validation.
// Returns content for registered types (via reference), inline schemas for primitives, or error for unknown types.
func buildJSONContent(typeName string, examples map[string]any, types map[string]*TypeInfo) (*orderedmap.Map[string, *v3.MediaType], error) {
	// Check if type is registered in types map
	if _, ok := types[typeName]; ok {
		// Type exists - create reference via createJSONContent
		return createJSONContent(typeName, examples)
	}

	// Check if it's a primitive type
	if isPrimitiveType(typeName) {
		examplesMap, err := convertExamplesToOpenAPI(examples)
		if err != nil {
			return nil, err
		}

		// Primitive type - create inline schema
		content := orderedmap.New[string, *v3.MediaType]()
		content.Set("application/json", &v3.MediaType{
			Schema:   base.CreateSchemaProxy(&base.Schema{Type: []string{typeName}}),
			Examples: examplesMap,
		})
		return content, nil
	}

	// Unknown type - return error
	return nil, fmt.Errorf("type %s not found in types map and not a valid primitive type", typeName)
}

// createSchemaRef creates a schema reference for the given type name.
func createSchemaRef(typeName string) *base.SchemaProxy {
	return base.CreateSchemaProxyRef("#/components/schemas/" + typeName)
}

// convertExamplesToOpenAPI converts examples map to OpenAPI format.
// Returns an error if any example fails to encode (indicates a bug in the example data).
func convertExamplesToOpenAPI(examples map[string]any) (*orderedmap.Map[string, *base.Example], error) {
	if len(examples) == 0 {
		return nil, nil
	}

	result := orderedmap.New[string, *base.Example]()
	for name, value := range examples {
		// Marshal through JSON first to respect json tags (not yaml field names)
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal example %q to JSON: %w", name, err)
		}

		// Unmarshal into a generic map to get clean structure
		var intermediate any
		if err := json.Unmarshal(jsonBytes, &intermediate); err != nil {
			return nil, fmt.Errorf("failed to unmarshal example %q from JSON: %w", name, err)
		}

		// Now encode to yaml.Node (this will use proper YAML structure)
		node := &yaml.Node{}
		if err := node.Encode(intermediate); err != nil {
			return nil, fmt.Errorf("failed to encode example %q to YAML node: %w", name, err)
		}

		result.Set(name, &base.Example{Value: node})
	}

	return result, nil
}
