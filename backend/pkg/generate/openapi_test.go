package generate

import (
	"http-mqtt-boilerplate/backend/pkg/utils"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestIsPrimitiveType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		typeName string
		want     bool
	}{
		{"string is primitive", "string", true},
		{"number is primitive", "number", true},
		{"integer is primitive", "integer", true},
		{"boolean is primitive", "boolean", true},
		{"array is not primitive", "array", false},
		{"object is not primitive", "object", false},
		{"custom type is not primitive", "MyType", false},
		{"empty string is not primitive", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := isPrimitiveType(tt.typeName); got != tt.want {
				t.Errorf("isPrimitiveType(%q) = %v, want %v", tt.typeName, got, tt.want)
			}
		})
	}
}

func TestIsEnumKind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		kind string
		want bool
	}{
		{"string enum is enum", TypeKindStringEnum, true},
		{"number enum is enum", TypeKindNumberEnum, true},
		{"object is not enum", TypeKindObject, false},
		{"alias is not enum", TypeKindAlias, false},
		{"empty is not enum", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := isEnumKind(tt.kind); got != tt.want {
				t.Errorf("isEnumKind(%q) = %v, want %v", tt.kind, got, tt.want)
			}
		})
	}
}

func TestBuildObjectSchema(t *testing.T) {
	t.Parallel()

	typeInfo := &TypeInfo{
		Name:        "TestObject",
		Kind:        TypeKindObject,
		Description: "Test object description",
		Fields: []FieldInfo{
			{
				Name:        "name",
				DisplayType: "String",
				TypeInfo: FieldType{
					Kind:     FieldKindPrimitive,
					Type:     "string",
					Required: true,
				},
				Description: "Name field",
			},
			{
				Name:        "count",
				DisplayType: "Integer",
				TypeInfo: FieldType{
					Kind:     FieldKindPrimitive,
					Type:     "integer",
					Required: false,
				},
				Description: "Count field",
			},
		},
	}

	schema, err := buildObjectSchema(typeInfo)
	if err != nil {
		t.Fatalf("buildObjectSchema() error = %v", err)
	}

	if schema.Type == nil || len(*schema.Type) != 1 || (*schema.Type)[0] != "object" {
		t.Errorf("Expected type 'object', got %v", schema.Type)
	}

	if schema.Description != typeInfo.Description {
		t.Errorf("Description = %q, want %q", schema.Description, typeInfo.Description)
	}

	if len(schema.Properties) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(schema.Properties))
	}

	if len(schema.Required) != 1 || schema.Required[0] != "name" {
		t.Errorf("Expected required = ['name'], got %v", schema.Required)
	}

	// Check additional properties is false
	if schema.AdditionalProperties.Has == nil || *schema.AdditionalProperties.Has {
		t.Error("AdditionalProperties should be false for structured objects")
	}
}

func TestBuildEnumSchema(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		typeInfo     *TypeInfo
		wantType     string
		wantEnumSize int
	}{
		{
			name: "string enum",
			typeInfo: &TypeInfo{
				Name:        "Status",
				Kind:        TypeKindStringEnum,
				Description: "Status enum",
				EnumValues: []EnumValue{
					{Value: "active", Description: "Active status"},
					{Value: "inactive", Description: "Inactive status"},
				},
			},
			wantType:     "string",
			wantEnumSize: 2,
		},
		{
			name: "number enum",
			typeInfo: &TypeInfo{
				Name: "Priority",
				Kind: TypeKindNumberEnum,
				EnumValues: []EnumValue{
					{Value: int64(1), Description: "Low"},
					{Value: int64(2), Description: "Medium"},
					{Value: int64(3), Description: "High"},
				},
			},
			wantType:     "integer",
			wantEnumSize: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			schema, err := buildEnumSchema(tt.typeInfo)
			if err != nil {
				t.Fatalf("buildEnumSchema() error = %v", err)
			}

			if schema.Type == nil || len(*schema.Type) != 1 || (*schema.Type)[0] != tt.wantType {
				t.Errorf("Expected type %q, got %v", tt.wantType, schema.Type)
			}

			if len(schema.Enum) != tt.wantEnumSize {
				t.Errorf("Expected %d enum values, got %d", tt.wantEnumSize, len(schema.Enum))
			}
		})
	}
}

func TestBuildAliasSchema(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		typeInfo *TypeInfo
		wantErr bool
	}{
		{
			name: "alias to string",
			typeInfo: &TypeInfo{
				Name:        "UserID",
				Kind:        TypeKindAlias,
				Description: "User identifier",
				UnderlyingType: &FieldType{
					Kind: FieldKindPrimitive,
					Type: "string",
				},
			},
			wantErr: false,
		},
		{
			name: "alias without underlying type",
			typeInfo: &TypeInfo{
				Name: "BadAlias",
				Kind: TypeKindAlias,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			schema, err := buildAliasSchema(tt.typeInfo)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildAliasSchema() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && schema == nil {
				t.Error("buildAliasSchema() returned nil schema")
			}
		})
	}
}

func TestApplyNullable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		schema   *openapi3.SchemaRef
		nullable bool
		wantErr  bool
	}{
		{
			name: "inline schema nullable true",
			schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{Type: &openapi3.Types{"string"}},
			},
			nullable: true,
			wantErr:  false,
		},
		{
			name: "inline schema nullable false",
			schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{Type: &openapi3.Types{"string"}},
			},
			nullable: false,
			wantErr:  false,
		},
		{
			name: "reference schema nullable true",
			schema: &openapi3.SchemaRef{
				Ref: "#/components/schemas/User",
			},
			nullable: true,
			wantErr:  false,
		},
		{
			name: "reference schema nullable false",
			schema: &openapi3.SchemaRef{
				Ref: "#/components/schemas/User",
			},
			nullable: false,
			wantErr:  false,
		},
		{
			name:     "invalid schema",
			schema:   &openapi3.SchemaRef{},
			nullable: true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := applyNullable(tt.schema, tt.nullable)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyNullable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Fatal("applyNullable() returned nil")
				}

				if tt.nullable {
					// Check that nullable was applied
					if tt.schema.Value != nil && !result.Value.Nullable {
						t.Error("Expected nullable to be true")
					}
					if tt.schema.Ref != "" && result.Value == nil {
						t.Error("Expected wrapper schema for reference")
					}
				}
			}
		})
	}
}

func TestApplyDeprecated(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		schema     *openapi3.SchemaRef
		deprecated bool
		wantErr    bool
	}{
		{
			name: "inline schema deprecated true",
			schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{Type: &openapi3.Types{"string"}},
			},
			deprecated: true,
			wantErr:    false,
		},
		{
			name: "inline schema deprecated false",
			schema: &openapi3.SchemaRef{
				Value: &openapi3.Schema{Type: &openapi3.Types{"string"}},
			},
			deprecated: false,
			wantErr:    false,
		},
		{
			name: "reference schema deprecated true",
			schema: &openapi3.SchemaRef{
				Ref: "#/components/schemas/User",
			},
			deprecated: true,
			wantErr:    false,
		},
		{
			name:       "invalid schema",
			schema:     &openapi3.SchemaRef{},
			deprecated: true,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := applyDeprecated(tt.schema, tt.deprecated)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyDeprecated() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result != nil && tt.deprecated {
				if tt.schema.Value != nil && !result.Value.Deprecated {
					t.Error("Expected deprecated to be true")
				}
			}
		})
	}
}

func TestBuildPrimitiveSchemaFromFieldType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ft       FieldType
		desc     string
		wantType string
		wantFmt  string
	}{
		{
			name:     "string",
			ft:       FieldType{Kind: FieldKindPrimitive, Type: "string"},
			desc:     "A string field",
			wantType: "string",
		},
		{
			name:     "integer with format",
			ft:       FieldType{Kind: FieldKindPrimitive, Type: "integer", Format: "int64"},
			desc:     "An integer field",
			wantType: "integer",
			wantFmt:  "int64",
		},
		{
			name:     "nullable string",
			ft:       FieldType{Kind: FieldKindPrimitive, Type: "string", Nullable: true},
			desc:     "A nullable string",
			wantType: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := buildPrimitiveSchemaFromFieldType(tt.ft, tt.desc)
			if err != nil {
				t.Fatalf("buildPrimitiveSchemaFromFieldType() error = %v", err)
			}

			if result == nil || result.Value == nil {
				t.Fatal("Expected non-nil schema")
			}

			if result.Value.Description != tt.desc {
				t.Errorf("Description = %q, want %q", result.Value.Description, tt.desc)
			}

			if tt.wantFmt != "" && result.Value.Format != tt.wantFmt {
				t.Errorf("Format = %q, want %q", result.Value.Format, tt.wantFmt)
			}
		})
	}
}

func TestBuildArraySchemaFromFieldType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ft      FieldType
		wantErr bool
	}{
		{
			name: "array of strings",
			ft: FieldType{
				Kind: FieldKindArray,
				Type: "array",
				ItemsType: &FieldType{
					Kind: FieldKindPrimitive,
					Type: "string",
				},
			},
			wantErr: false,
		},
		{
			name: "array without items type",
			ft: FieldType{
				Kind: FieldKindArray,
				Type: "array",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := buildArraySchemaFromFieldType(tt.ft, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("buildArraySchemaFromFieldType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && (result == nil || result.Value == nil) {
				t.Fatal("Expected non-nil schema")
			}
		})
	}
}

func TestBuildReferenceSchemaFromFieldType(t *testing.T) {
	t.Parallel()

	ft := FieldType{
		Kind: FieldKindReference,
		Type: "User",
	}

	result, err := buildReferenceSchemaFromFieldType(ft)
	if err != nil {
		t.Fatalf("buildReferenceSchemaFromFieldType() error = %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil schema")
	}

	if result.Ref != "#/components/schemas/User" {
		t.Errorf("Ref = %q, want %q", result.Ref, "#/components/schemas/User")
	}
}

func TestBuildObjectSchemaFromFieldType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ft      FieldType
		wantErr bool
	}{
		{
			name: "map type",
			ft: FieldType{
				Kind: FieldKindObject,
				Type: "object",
				AdditionalProperties: &FieldType{
					Kind: FieldKindPrimitive,
					Type: "string",
				},
			},
			wantErr: false,
		},
		{
			name: "plain object",
			ft: FieldType{
				Kind: FieldKindObject,
				Type: "object",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := buildObjectSchemaFromFieldType(tt.ft, "Test description")
			if (err != nil) != tt.wantErr {
				t.Errorf("buildObjectSchemaFromFieldType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && (result == nil || result.Value == nil) {
				t.Fatal("Expected non-nil schema")
			}
		})
	}
}

func TestFormatEnumValueDescription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		enumValue EnumValue
		want      string
	}{
		{
			name: "with description",
			enumValue: EnumValue{
				Value:       "active",
				Description: "User is active",
			},
			want: "- `active`: User is active\n",
		},
		{
			name: "without description",
			enumValue: EnumValue{
				Value: "inactive",
			},
			want: "- `inactive`\n",
		},
		{
			name: "deprecated with description",
			enumValue: EnumValue{
				Value:       "legacy",
				Description: "Old status",
				Deprecated:  "Use 'active' instead",
			},
			want: "- `legacy`: **[DEPRECATED]** Use 'active' instead - Old status\n",
		},
		{
			name: "deprecated without description",
			enumValue: EnumValue{
				Value:      "old",
				Deprecated: "No longer used",
			},
			want: "- `old`: **[DEPRECATED]** No longer used\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := formatEnumValueDescription(tt.enumValue)
			if got != tt.want {
				t.Errorf("formatEnumValueDescription() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCreateSchemaRef(t *testing.T) {
	t.Parallel()

	typeName := "User"
	ref := createSchemaRef(typeName)

	if ref == nil {
		t.Fatal("createSchemaRef() returned nil")
	}

	expectedRef := "#/components/schemas/User"
	if ref.Ref != expectedRef {
		t.Errorf("Ref = %q, want %q", ref.Ref, expectedRef)
	}
}

func TestConvertExamplesToOpenAPI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		examples map[string]any
		wantSize int
	}{
		{
			name: "multiple examples",
			examples: map[string]any{
				"example1": map[string]string{"name": "test1"},
				"example2": map[string]string{"name": "test2"},
			},
			wantSize: 2,
		},
		{
			name:     "nil examples",
			examples: nil,
			wantSize: 0,
		},
		{
			name:     "empty examples",
			examples: map[string]any{},
			wantSize: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := convertExamplesToOpenAPI(tt.examples)

			if result == nil && tt.wantSize > 0 {
				t.Fatal("convertExamplesToOpenAPI() returned nil")
			}

			if len(result) != tt.wantSize {
				t.Errorf("Result size = %d, want %d", len(result), tt.wantSize)
			}
		})
	}
}

func TestSchemaToJSONString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		schema  *openapi3.Schema
		wantErr bool
	}{
		{
			name: "simple schema",
			schema: &openapi3.Schema{
				Type:        &openapi3.Types{"string"},
				Description: "A string field",
			},
			wantErr: false,
		},
		{
			name:    "nil schema",
			schema:  nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := schemaToJSONString(tt.schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("schemaToJSONString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.schema != nil && result == "" {
				t.Error("schemaToJSONString() returned empty string for non-nil schema")
			}
		})
	}
}

func TestBuildJSONContent(t *testing.T) {
	t.Parallel()

	types := map[string]*TypeInfo{
		"User": {
			Name: "User",
			Kind: TypeKindObject,
		},
	}

	tests := []struct {
		name     string
		typeName string
		examples map[string]any
		wantErr  bool
	}{
		{
			name:     "registered type",
			typeName: "User",
			examples: map[string]any{"example1": map[string]string{"name": "test"}},
			wantErr:  false,
		},
		{
			name:     "primitive type",
			typeName: "string",
			examples: nil,
			wantErr:  false,
		},
		{
			name:     "unknown type",
			typeName: "UnknownType",
			examples: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			content, err := buildJSONContent(tt.typeName, tt.examples, types)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildJSONContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && content == nil {
				t.Error("buildJSONContent() returned nil content")
			}
		})
	}
}

func TestOpenAPIVersion(t *testing.T) {
	t.Parallel()

	if OpenAPIVersion != "3.0.3" {
		t.Errorf("OpenAPIVersion = %q, want %q", OpenAPIVersion, "3.0.3")
	}
}

func TestBuildComponentSchemas(t *testing.T) {
	t.Parallel()

	doc := &APIDocumentation{
		Types: map[string]*TypeInfo{
			"User": {
				Name:       "User",
				Kind:       TypeKindObject,
				UsedByHTTP: true,
				Fields:     []FieldInfo{},
			},
			"MQTTOnly": {
				Name:       "MQTTOnly",
				Kind:       TypeKindObject,
				UsedByMQTT: true,
				UsedByHTTP: false,
				Fields:     []FieldInfo{},
			},
		},
	}

	schemas, err := buildComponentSchemas(doc)
	if err != nil {
		t.Fatalf("buildComponentSchemas() error = %v", err)
	}

	if len(schemas) != 1 {
		t.Errorf("Expected 1 schema (only HTTP types), got %d", len(schemas))
	}

	if _, ok := schemas["User"]; !ok {
		t.Error("Expected User schema to be present")
	}

	if _, ok := schemas["MQTTOnly"]; ok {
		t.Error("MQTTOnly schema should not be present (not used by HTTP)")
	}
}

func TestGenerateOpenAPISpec(t *testing.T) {
	t.Parallel()

	doc := &APIDocumentation{
		Info: APIInfo{
			Title:       "Test API",
			Version:     "1.0.0",
			Description: "Test description",
		},
		Types: map[string]*TypeInfo{
			"TestType": {
				Name:       "TestType",
				Kind:       TypeKindObject,
				UsedByHTTP: true,
				Fields:     []FieldInfo{},
			},
		},
		HTTPOperations: map[string]*RouteInfo{
			"testOp": {
				OperationID: "testOp",
				Method:      "GET",
				Path:        "/test",
				Summary:     "Test endpoint",
			},
		},
	}

	spec, err := generateOpenAPISpec(doc)
	if err != nil {
		t.Fatalf("generateOpenAPISpec() error = %v", err)
	}

	if spec == nil {
		t.Fatal("generateOpenAPISpec() returned nil")
	}

	if spec.OpenAPI != OpenAPIVersion {
		t.Errorf("OpenAPI version = %q, want %q", spec.OpenAPI, OpenAPIVersion)
	}

	if spec.Info.Title != doc.Info.Title {
		t.Errorf("Title = %q, want %q", spec.Info.Title, doc.Info.Title)
	}
}

func TestToOpenAPISchema(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		typeInfo *TypeInfo
		wantErr  bool
	}{
		{
			name: "object type",
			typeInfo: &TypeInfo{
				Name:   "User",
				Kind:   TypeKindObject,
				Fields: []FieldInfo{},
			},
			wantErr: false,
		},
		{
			name: "string enum",
			typeInfo: &TypeInfo{
				Name: "Status",
				Kind: TypeKindStringEnum,
				EnumValues: []EnumValue{
					{Value: "active"},
				},
			},
			wantErr: false,
		},
		{
			name: "alias type",
			typeInfo: &TypeInfo{
				Name: "UserID",
				Kind: TypeKindAlias,
				UnderlyingType: &FieldType{
					Kind: FieldKindPrimitive,
					Type: "string",
				},
			},
			wantErr: false,
		},
		{
			name: "unsupported kind",
			typeInfo: &TypeInfo{
				Name: "Bad",
				Kind: "unsupported",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			schema, err := toOpenAPISchema(tt.typeInfo)
			if (err != nil) != tt.wantErr {
				t.Errorf("toOpenAPISchema() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && schema == nil {
				t.Error("toOpenAPISchema() returned nil schema")
			}
		})
	}
}

func TestBuildSchemaFromFieldType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ft      FieldType
		wantErr bool
	}{
		{
			name:    "primitive",
			ft:      FieldType{Kind: FieldKindPrimitive, Type: "string"},
			wantErr: false,
		},
		{
			name: "array",
			ft: FieldType{
				Kind: FieldKindArray,
				ItemsType: utils.Ptr(FieldType{
					Kind: FieldKindPrimitive,
					Type: "string",
				}),
			},
			wantErr: false,
		},
		{
			name:    "reference",
			ft:      FieldType{Kind: FieldKindReference, Type: "User"},
			wantErr: false,
		},
		{
			name:    "object",
			ft:      FieldType{Kind: FieldKindObject, Type: "object"},
			wantErr: false,
		},
		{
			name:    "unhandled kind",
			ft:      FieldType{Kind: "unknown"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := buildSchemaFromFieldType(tt.ft, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("buildSchemaFromFieldType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result == nil {
				t.Error("buildSchemaFromFieldType() returned nil")
			}
		})
	}
}