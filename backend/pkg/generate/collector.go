package generate

// This file (collector.go) handles Go AST parsing and metadata extraction
// using native Go AST parser to extract type information and generate
// Go source representations with full metadata.

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"http-mqtt-boilerplate/backend/pkg/utils"
	"log/slog"
	"maps"
	"os"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/coder/guts"
	"github.com/coder/guts/bindings"
	"github.com/coder/guts/config"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/yaml"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/tools/go/packages"
)

// Sentinel errors for specific cases.
var (
	ErrEmptyConstBlock = errors.New("empty const block")
	ErrNoEnumType      = errors.New("no enum type found in const block")
	ErrFieldSkipped    = errors.New("field skipped")
)

// isEnumKind checks if the given kind represents an enum type.
func isEnumKind(kind string) bool {
	return kind == TypeKindStringEnum || kind == TypeKindNumberEnum
}

// isNilOrNilPointer checks if a value is nil or a nil pointer.
// Returns true if the value should be rejected (is nil).
func isNilOrNilPointer(value any) bool {
	if value == nil {
		return true
	}

	val := reflect.ValueOf(value)

	return val.Kind() == reflect.Pointer && val.IsNil()
}

// isZeroValueStruct checks if a value is a zero-value struct (all fields are zero values).
// Returns true if it's a struct with all zero-valued fields.
// This is different from nil - a zero-value struct is an explicitly created struct
// like MyStruct{} or MyStruct{Field: false} where all fields happen to be zero.
func isZeroValueStruct(value any) bool {
	if value == nil {
		return false
	}

	val := reflect.ValueOf(value)
	// Handle pointer to struct
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return false
		}

		val = val.Elem()
	}

	// Only check structs
	if val.Kind() != reflect.Struct {
		return false
	}

	return val.IsZero()
}

// External type format constants for OpenAPI schema generation.
const (
	FormatDateTime = "date-time"
	FormatURI      = "uri"
)

// GoParser holds the parsed Go AST and type information.
type GoParser struct {
	fset  *token.FileSet
	files []*ast.File
	pkg   *packages.Package
}

type TSParser struct {
	ts *guts.Typescript
	vm *bindings.Bindings
}

// OpenAPICollector handles Go AST parsing and metadata extraction from Go types.
// It walks the Go AST to extract comprehensive type information in a single pass.
type OpenAPICollector struct {
	goParser            *GoParser
	tsParser            *TSParser
	externalTypeFormats map[string]string
	l                   *slog.Logger

	types             map[string]*TypeInfo             // Extracted type information, keyed by type name
	httpOps           map[string]*RouteInfo            // Registered HTTP operations, keyed by operationID
	mqttPublications  map[string]*MQTTPublicationInfo  // Registered MQTT publications, keyed by operationID
	mqttSubscriptions map[string]*MQTTSubscriptionInfo // Registered MQTT subscriptions, keyed by operationID
	database          Database                         // Database schema and stats

	// AST nodes for generating Go source representations
	typeASTs  map[string]*ast.GenDecl // Type declaration AST nodes, keyed by type name
	constASTs map[string]*ast.GenDecl // Const block AST nodes for enums, keyed by type name

	docsFilePath        string // Path to write documentation JSON file
	openAPISpecFilePath string // Path to write OpenAPI YAML file

	apiInfo     APIInfo
	openapiSpec string

	primitiveTypeMapping map[string]FieldType
}

// normalizeLocalPackagePath normalizes a path to be recognized as a local package.
// It ensures the path starts with "./" so the Go package parser treats it as local.
func normalizeLocalPackagePath(path string) string {
	path = strings.TrimPrefix(path, "./")
	path = strings.TrimPrefix(path, "/")

	return "./" + path
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

type OpenAPICollectorOptions struct {
	GoTypesDirPath               string // Path to Go types file for parsing
	DocsFileOutputPath           string // Path for generated API docs JSON file
	DatabaseSchemaFileOutputPath string // Path for generated DB schema SQL file
	OpenAPISpecOutputPath        string // Path for generated OpenAPI YAML file
	APIInfo                      APIInfo
}

// NewOpenAPICollector parses the Go types directory and generates a TypeScript AST for metadata extraction.
func NewOpenAPICollector(l *slog.Logger, opts OpenAPICollectorOptions) (*OpenAPICollector, error) {
	var err error

	l = l.With(slog.String("component", "openapi-collector"))

	if opts.GoTypesDirPath == "" {
		return nil, errors.New("go types dir path is required")
	}

	if opts.DatabaseSchemaFileOutputPath == "" {
		return nil, errors.New("database schema file path is required")
	}

	if opts.DocsFileOutputPath == "" {
		return nil, errors.New("docs file path is required")
	}

	if opts.OpenAPISpecOutputPath == "" {
		return nil, errors.New("OpenAPI spec file path is required")
	}

	// Normalize path to be recognized as a local package
	goTypesDirPath := normalizeLocalPackagePath(opts.GoTypesDirPath)

	l.Debug("Creating doc collector", slog.String("goTypesDirPath", goTypesDirPath))

	docCollector := &OpenAPICollector{
		l:                 l,
		types:             make(map[string]*TypeInfo),
		httpOps:           make(map[string]*RouteInfo),
		mqttPublications:  make(map[string]*MQTTPublicationInfo),
		mqttSubscriptions: make(map[string]*MQTTSubscriptionInfo),
		typeASTs:          make(map[string]*ast.GenDecl),
		constASTs:         make(map[string]*ast.GenDecl),
		externalTypeFormats: map[string]string{
			"time.Time": FormatDateTime,
			"http-mqtt-boilerplate/backend/pkg/types.URL": FormatURI,
		},
		docsFilePath:        opts.DocsFileOutputPath,
		openAPISpecFilePath: opts.OpenAPISpecOutputPath,
		apiInfo:             opts.APIInfo,

		primitiveTypeMapping: getPrimitiveTypeMappings(),
	}

	dbSchema, err := docCollector.GenerateDatabaseSchema(opts.DatabaseSchemaFileOutputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get database schema: %w", err)
	}

	docCollector.database.Schema = dbSchema

	dbStats, err := docCollector.GetDatabaseStats(dbSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to get database stats: %w", err)
	}

	docCollector.database.TableCount = dbStats.TableCount

	goParser, err := docCollector.parseGoTypesDir(goTypesDirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go types directory: %w", err)
	}

	docCollector.goParser = goParser

	tsParser, err := newTSParser(l, goTypesDirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create TypeScript parser: %w", err)
	}

	docCollector.tsParser = tsParser

	// Walk the AST and extract all type information in one pass
	if err := docCollector.extractAllTypesFromGo(goParser); err != nil {
		return nil, fmt.Errorf("failed to extract types: %w", err)
	}

	l.Info("OpenAPI collector created successfully", slog.Int("types", len(docCollector.types)))

	return docCollector, nil
}

// newTSParser creates a TypeScript parser using guts for the specified Go types directory.
func newTSParser(l *slog.Logger, goTypesDirPath string) (*TSParser, error) {
	l.Debug("Parsing Go types directory", slog.String("path", goTypesDirPath))

	goParser, err := guts.NewGolangParser()
	if err != nil {
		return nil, fmt.Errorf("failed to create guts parser: %w", err)
	}

	goParser.PreserveComments()
	goParser.IncludeCustomDeclaration(map[string]guts.TypeOverride{
		"time.Time": func() bindings.ExpressionType {
			return utils.Ptr(bindings.KeywordString)
		},
		"http-mqtt-boilerplate/backend/pkg/types.URL": func() bindings.ExpressionType {
			return utils.Ptr(bindings.KeywordString)
		},
	})

	if _, err := os.Stat(goTypesDirPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to validate TypeScript parser: go types dir path %s does not exist", goTypesDirPath)
	}

	if err := goParser.IncludeGenerate(goTypesDirPath); err != nil {
		return nil, fmt.Errorf("failed to include go types dir for parsing: %w", err)
	}

	var errs []error

	for _, pkg := range goParser.Pkgs {
		for _, e := range pkg.Errors {
			errs = append(errs, fmt.Errorf("failed to parse go types in %s: %w", pkg.PkgPath, e))
		}
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	l.Debug("Generating TypeScript AST from Go types")

	ts, err := goParser.ToTypescript()
	if err != nil {
		return nil, fmt.Errorf("failed to generate TypeScript AST: %w", err)
	}

	ts.ApplyMutations(
		config.InterfaceToType,
		config.SimplifyOptional,
		config.NotNullMaps,
	)

	l.Debug("TypeScript AST generated successfully")

	vm, err := bindings.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create bindings: %w", err)
	}

	tsParser := &TSParser{
		ts: ts,
		vm: vm,
	}

	return tsParser, nil
}

func (g *OpenAPICollector) SerializeTSNode(name string) (string, error) {
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

// Generate generates both the OpenAPI spec YAML and the docs JSON file.
func (g *OpenAPICollector) Generate() error {
	// Compute type relationships
	g.computeTypeRelationships()

	// Generate type representations
	if err := g.generateTypesRepresentations(); err != nil {
		return fmt.Errorf("failed to generate types representations: %w", err)
	}

	// Write OpenAPI spec
	if err := g.writeSpecYAML(g.openAPISpecFilePath); err != nil {
		return fmt.Errorf("failed to write OpenAPI spec: %w", err)
	}

	// read the written OpenAPI spec file
	yamlBytes, err := os.ReadFile(g.openAPISpecFilePath)
	if err != nil {
		return fmt.Errorf("failed to read OpenAPI spec file: %w", err)
	}

	g.openapiSpec = string(yamlBytes)

	g.l.Info("OpenAPI spec written", slog.String("file", g.openAPISpecFilePath))

	// Write docs JSON
	if err := g.writeDocsJSON(); err != nil {
		return fmt.Errorf("failed to write docs JSON: %w", err)
	}

	g.l.Info("API documentation generated")

	return nil
}

// stringifyExamples converts examples to stringified JSON.
func stringifyExamples(examples map[string]any) map[string]string {
	stringified := make(map[string]string)
	for name, example := range examples {
		stringified[name] = string(utils.MustToJSONIndent(example))
	}

	return stringified
}

// RegisterJSONRepresentation registers the JSON representation of a type value.
// It makes sure to only store the largest representation for the type.
func (g *OpenAPICollector) RegisterJSONRepresentation(value any) error {
	typeName, err := extractTypeNameFromValue(value)
	if err != nil {
		return fmt.Errorf("failed to extract type name: %w", err)
	}

	// Skip primitive types - they don't need JSON representations
	if _, isPrimitive := g.primitiveTypeMapping[typeName]; isPrimitive {
		return nil
	}

	typeInfo, ok := g.types[typeName]
	if !ok {
		return fmt.Errorf("failed to register JSON representation: type %s not found in types map", typeName)
	}

	representation := string(utils.MustToJSONIndent(value))

	// If stored representation is empty or shorter, update it
	if typeInfo.Representations.JSON == "" || len(representation) > len(typeInfo.Representations.JSON) {
		typeInfo.Representations.JSON = representation
	}

	return nil
}

func (g *OpenAPICollector) RegisterRoute(route *RouteInfo) error {
	// Validate operationID format
	if err := validateOperationIDFormat(route.OperationID); err != nil {
		return err
	}

	// Validate operationID is unique
	if _, exists := g.httpOps[route.OperationID]; exists {
		return fmt.Errorf("duplicate operationID: %s", route.OperationID)
	}

	// Process request body if provided
	// route.Request can be nil (operation has no request body)
	// If route.Request is provided, its TypeValue must be a valid (non-nil) type
	if route.Request != nil {
		if isNilOrNilPointer(route.Request.TypeValue) {
			return fmt.Errorf("request TypeValue must not be nil when Request is provided in route [%s]", route.OperationID)
		}

		typeName, stringifiedExamples, err := g.processHTTPType(route.Request.TypeValue, route.Request.Examples, "request")
		if err != nil {
			return fmt.Errorf("failed to process request type in route [%s]: %w", route.OperationID, err)
		}

		route.Request.TypeName = typeName
		route.Request.ExamplesStringified = stringifiedExamples
	}

	// Process responses (required - every route must have at least one response)
	// Response TypeValue must be a zero-value struct (e.g., MyResponse{})
	// This indicates the type without providing actual data (examples provide the data)
	for statusCode, response := range route.Responses {
		if isNilOrNilPointer(response.TypeValue) {
			return fmt.Errorf("response TypeValue must not be nil in route [%s] for status %d", route.OperationID, statusCode)
		}

		if !isZeroValueStruct(response.TypeValue) {
			return fmt.Errorf("response TypeValue must be zero value struct (e.g., MyResponse{}) in route [%s] for status %d - use Examples for actual data", route.OperationID, statusCode)
		}

		resp := response

		typeName, stringifiedExamples, err := g.processHTTPType(resp.TypeValue, resp.Examples, "response")
		if err != nil {
			return fmt.Errorf("failed to process response type [%s] for status code [%d] in route [%s]: %w", typeName, statusCode, route.OperationID, err)
		}

		resp.TypeName = typeName
		resp.ExamplesStringified = stringifiedExamples
		route.Responses[statusCode] = resp
	}

	for i := range route.Parameters {
		typeName, _, err := g.processHTTPType(route.Parameters[i].TypeValue, nil, "parameter")
		if err != nil {
			return fmt.Errorf("failed to process parameter type in route [%s]: %w", route.OperationID, err)
		}

		route.Parameters[i].TypeName = typeName
	}

	// Add operation keyed by operationID
	g.httpOps[route.OperationID] = route

	return nil
}

func (g *OpenAPICollector) RegisterMQTTPublication(pub *MQTTPublicationInfo) error {
	// Validate operationID format
	if err := validateOperationIDFormat(pub.OperationID); err != nil {
		return err
	}

	// Validate operationID is unique
	if err := g.validateUniqueOperationID(pub.OperationID); err != nil {
		return err
	}

	// Process message type and examples
	typeName, stringifiedExamples, err := g.processMQTTMessageType(pub.OperationID, pub.TypeValue, pub.Examples, "publication")
	if err != nil {
		return err
	}

	pub.TypeName = typeName
	pub.ExamplesStringified = stringifiedExamples

	for i := range pub.TopicParameters {
		typeName, err := g.processMQTTTopicParameter(pub.OperationID, pub.TopicParameters[i].TypeValue, "publication topic parameter")
		if err != nil {
			return err
		}

		pub.TopicParameters[i].TypeName = typeName
	}

	// Store publication
	g.mqttPublications[pub.OperationID] = pub

	return nil
}

func (g *OpenAPICollector) RegisterMQTTSubscription(sub *MQTTSubscriptionInfo) error {
	// Validate operationID format
	if err := validateOperationIDFormat(sub.OperationID); err != nil {
		return err
	}

	// Validate operationID is unique
	if err := g.validateUniqueOperationID(sub.OperationID); err != nil {
		return err
	}

	// Process message type and examples
	typeName, stringifiedExamples, err := g.processMQTTMessageType(sub.OperationID, sub.TypeValue, sub.Examples, "subscription")
	if err != nil {
		return err
	}

	sub.TypeName = typeName
	sub.ExamplesStringified = stringifiedExamples

	for i := range sub.TopicParameters {
		typeName, err := g.processMQTTTopicParameter(sub.OperationID, sub.TopicParameters[i].TypeValue, "subscription topic parameter")
		if err != nil {
			return err
		}

		sub.TopicParameters[i].TypeName = typeName
	}

	// Store subscription
	g.mqttSubscriptions[sub.OperationID] = sub

	return nil
}

// processHTTPType extracts type name, marks it as HTTP, and registers JSON representations.
// Returns the extracted type name.
func (g *OpenAPICollector) processHTTPType(typeValue any, examples map[string]any, contextMsg string) (string, map[string]string, error) {
	typeName, err := extractTypeNameFromValue(typeValue)
	if err != nil {
		return "", nil, fmt.Errorf("failed to extract %s type name: %w", contextMsg, err)
	}

	// Mark as used by HTTP (for OpenAPI spec filtering)
	g.markTypeAsHTTP(typeName)

	if err := g.RegisterJSONRepresentation(typeValue); err != nil {
		return "", nil, fmt.Errorf("failed to register JSON representation for %s type [%s]: %w", contextMsg, typeName, err)
	}

	// Register and stringify examples if provided
	var stringifiedExamples map[string]string

	if examples != nil {
		if err := g.registerExamples(examples); err != nil {
			return "", nil, fmt.Errorf("failed to register JSON representation for %s example: %w", contextMsg, err)
		}

		stringifiedExamples = stringifyExamples(examples)
	}

	return typeName, stringifiedExamples, nil
}

// processMQTTMessageType extracts type information and registers representations for an MQTT message.
// Returns the type name and stringified examples.
func (g *OpenAPICollector) processMQTTMessageType(operationID string, typeValue any, examples map[string]any, messageKind string) (typeName string, stringifiedExamples map[string]string, err error) {
	// Validate type value is not nil
	if isNilOrNilPointer(typeValue) {
		return "", nil, fmt.Errorf("MessageType must not be nil in %s [%s]", messageKind, operationID)
	}

	// Extract type name from zero value using reflection
	typeName, err = extractTypeNameFromValue(typeValue)
	if err != nil {
		return "", nil, fmt.Errorf("failed to extract message type name: %w", err)
	}

	// Mark as used by MQTT
	g.markTypeAsMQTT(typeName)

	// Register JSON representation
	if err := g.RegisterJSONRepresentation(typeValue); err != nil {
		return "", nil, fmt.Errorf("failed to register JSON representation for message type: %w", err)
	}

	// Register examples
	if err := g.registerExamples(examples); err != nil {
		return "", nil, fmt.Errorf("failed to register JSON representation for example: %w", err)
	}

	// Stringify examples
	stringifiedExamples = stringifyExamples(examples)

	return typeName, stringifiedExamples, nil
}

// processMQTTTopicParameter extracts type information and registers representations for an MQTT topic parameter.
// Returns the type name.
func (g *OpenAPICollector) processMQTTTopicParameter(operationID string, typeValue any, contextMsg string) (string, error) {
	// Validate type value is not nil
	if isNilOrNilPointer(typeValue) {
		return "", fmt.Errorf("topic parameter TypeValue must not be nil in %s [%s]", contextMsg, operationID)
	}

	// Extract type name from zero value using reflection
	typeName, err := extractTypeNameFromValue(typeValue)
	if err != nil {
		return "", fmt.Errorf("failed to extract topic parameter type name: %w", err)
	}

	// Mark as used by MQTT
	g.markTypeAsMQTT(typeName)

	// Register JSON representation
	if err := g.RegisterJSONRepresentation(typeValue); err != nil {
		return "", fmt.Errorf("failed to register JSON representation for topic parameter: %w", err)
	}

	return typeName, nil
}

// registerExamples registers JSON representations for a slice of examples.
func (g *OpenAPICollector) registerExamples(examples map[string]any) error {
	for name, ex := range examples {
		if isNilOrNilPointer(ex) {
			return fmt.Errorf("value for example [%s] should not be nil", name)
		}

		if err := g.RegisterJSONRepresentation(ex); err != nil {
			return err
		}
	}

	return nil
}

// validateOperationIDFormat checks that an operationID contains only characters a-z, A-Z.
func validateOperationIDFormat(operationID string) error {
	if operationID == "" {
		return errors.New("operationID cannot be empty")
	}

	if !IsASCIILetterString(operationID) {
		return fmt.Errorf("operationID %q contains invalid characters (only characters a-z, A-Z are allowed)", operationID)
	}

	return nil
}

// validateUniqueOperationID checks that an operationID is not already used.
func (g *OpenAPICollector) validateUniqueOperationID(operationID string) error {
	if _, exists := g.mqttPublications[operationID]; exists {
		return fmt.Errorf("duplicate operationID (MQTT publication exists): %s", operationID)
	}

	if _, exists := g.mqttSubscriptions[operationID]; exists {
		return fmt.Errorf("duplicate operationID (MQTT subscription exists): %s", operationID)
	}

	if _, exists := g.httpOps[operationID]; exists {
		return fmt.Errorf("duplicate operationID (HTTP operation exists): %s", operationID)
	}

	return nil
}

// ProtocolType represents the type of protocol using a type.
type ProtocolType int

const (
	ProtocolHTTP ProtocolType = iota
	ProtocolMQTT
)

// markTypeAsUsedBy recursively marks a type and all its referenced types as used by a protocol.
func (g *OpenAPICollector) markTypeAsUsedBy(typeName string, protocol ProtocolType) {
	if typeName == "" {
		return
	}

	typeInfo, exists := g.types[typeName]
	if !exists {
		return // Primitive or external type
	}

	// Check if already marked and mark this type
	switch protocol {
	case ProtocolHTTP:
		if typeInfo.UsedByHTTP {
			return
		}

		typeInfo.UsedByHTTP = true
	case ProtocolMQTT:
		if typeInfo.UsedByMQTT {
			return
		}

		typeInfo.UsedByMQTT = true
	}

	// Recursively mark referenced types
	for _, ref := range typeInfo.References {
		g.markTypeAsUsedBy(ref, protocol)
	}
}

// markTypeAsHTTP marks a type and all its referenced types as used by HTTP.
func (g *OpenAPICollector) markTypeAsHTTP(typeName string) {
	g.markTypeAsUsedBy(typeName, ProtocolHTTP)
}

// markTypeAsMQTT marks a type and all its referenced types as used by MQTT.
func (g *OpenAPICollector) markTypeAsMQTT(typeName string) {
	g.markTypeAsUsedBy(typeName, ProtocolMQTT)
}

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
			return nil, fmt.Errorf("field of struct type %s has no name", name)
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

func (g *OpenAPICollector) getDocumentation() *APIDocumentation {
	return &APIDocumentation{
		Types:             g.types,
		HTTPOperations:    g.httpOps,
		MQTTPublications:  g.mqttPublications,
		MQTTSubscriptions: g.mqttSubscriptions,
		Database:          g.database,
		Info:              g.apiInfo,
		OpenAPISpec:       g.openapiSpec,
	}
}

// computeTypeRelationships computes ReferencedBy and UsedBy for all types
// (References are already computed during type extraction).
func (g *OpenAPICollector) computeTypeRelationships() {
	g.l.Debug("Computing type relationships", slog.Int("typeCount", len(g.types)), slog.Int("operationCount", len(g.httpOps)))

	// Build ReferencedBy from References
	g.buildReferencedBy()
	g.l.Debug("Built ReferencedBy relationships")

	// Build UsedBy from routes
	g.buildUsedBy()
	g.l.Debug("Computed UsedBy relationships")
}

// generateGoRepresentations generates Go source representations for all types as a post-processing step.
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
		tsSource, err := g.SerializeTSNode(name)
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
	}

	g.l.Debug("All representations generated successfully")

	return nil
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

// buildReferencedBy builds the inverse of References for all types.
func (g *OpenAPICollector) buildReferencedBy() {
	// Track which types were modified so we only sort those
	modifiedTypes := make(map[string]struct{})

	for typeName, typeInfo := range g.types {
		for _, ref := range typeInfo.References {
			if refType, exists := g.types[ref]; exists {
				refType.ReferencedBy = append(refType.ReferencedBy, typeName)
				modifiedTypes[ref] = struct{}{}
			}
		}
	}

	// Sort only the types that received new references
	for typeName := range modifiedTypes {
		sort.Strings(g.types[typeName].ReferencedBy)
	}
}

// buildUsedBy tracks which operations use each type.
func (g *OpenAPICollector) buildUsedBy() {
	// Track HTTP operations
	for _, route := range g.httpOps {
		// Track request type
		if route.Request != nil {
			g.addUsage(route.Request.TypeName, route.OperationID, "request")
		}

		// Track response types
		for _, resp := range route.Responses {
			g.addUsage(resp.TypeName, route.OperationID, "response")
		}

		// Track parameter types
		for _, param := range route.Parameters {
			g.addUsage(param.TypeName, route.OperationID, "parameter")
		}
	}

	// Track MQTT publications
	for _, pub := range g.mqttPublications {
		g.addUsage(pub.TypeName, pub.OperationID, "mqtt_publication")
	}

	// Track MQTT subscriptions
	for _, sub := range g.mqttSubscriptions {
		g.addUsage(sub.TypeName, sub.OperationID, "mqtt_subscription")
	}

	for _, typ := range g.types {
		usages := make(map[string]struct{})
		dedupedUsages := []UsageInfo{}

		for _, usage := range typ.UsedBy {
			key := fmt.Sprintf("%s:%s", usage.OperationID, usage.Role)
			if _, used := usages[key]; used {
				continue
			}

			dedupedUsages = append(dedupedUsages, usage)
			usages[key] = struct{}{}
		}

		if len(dedupedUsages) == 0 {
			dedupedUsages = nil
		}

		typ.UsedBy = dedupedUsages
	}
}

// addUsage adds a UsageInfo entry to a type if it exists.
func (g *OpenAPICollector) addUsage(typeName, operationID, role string) {
	if typeName == "" {
		return
	}

	if typeInfo, exists := g.types[typeName]; exists {
		typeInfo.UsedBy = append(typeInfo.UsedBy, UsageInfo{
			OperationID: operationID,
			Role:        role,
		})
	}
}

// generateOpenAPISpec generates a complete OpenAPI specification from all collected metadata.
func (g *OpenAPICollector) generateOpenAPISpec() (*openapi3.T, error) {
	doc := g.getDocumentation()

	spec, err := generateOpenAPISpec(doc)
	if err != nil {
		return nil, err
	}

	// Set API metadata
	spec.Info.Title = g.apiInfo.Title
	spec.Info.Version = g.apiInfo.Version
	spec.Info.Description = g.apiInfo.Description

	for _, server := range g.apiInfo.Servers {
		spec.Servers = append(spec.Servers, &openapi3.Server{
			URL:         server.URL,
			Description: server.Description,
		})
	}

	return spec, nil
}

const DEPRECATED_PREFIX = "deprecated:"

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

// writeSpecYAML writes the OpenAPI specification to a YAML file.
func (g *OpenAPICollector) writeSpecYAML(filename string) error {
	spec, err := g.generateOpenAPISpec()
	if err != nil {
		return fmt.Errorf("failed to generate spec: %w", err)
	}

	yamlData, err := yaml.Marshal(spec)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, yamlData, 0600)
}

// writeDocsJSON writes the complete API documentation to a JSON file.
func (g *OpenAPICollector) writeDocsJSON() error {
	if g.docsFilePath == "" {
		return nil // Skip if no path configured
	}

	doc := g.getDocumentation()

	// Use GenerateAPIDocs for sorted, deterministic output
	if err := GenerateAPIDocs(g.l, doc, g.docsFilePath); err != nil {
		return fmt.Errorf("failed to write docs JSON: %w", err)
	}

	g.l.Info("API documentation written", slog.String("file", g.docsFilePath))

	return nil
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
