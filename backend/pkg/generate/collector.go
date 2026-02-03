package generate

// This file (collector.go) handles Go AST parsing and metadata extraction
// using native Go AST parser to extract type information and generate
// Go source representations with full metadata.

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"http-mqtt-boilerplate/backend/pkg/utils"
	"log/slog"
	"os"
	"reflect"
	"strings"

	"github.com/coder/guts"
	"github.com/coder/guts/bindings"
	"github.com/coder/guts/config"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/yaml"
	"golang.org/x/tools/go/packages"
)

// Sentinel errors for specific cases.
var (
	ErrFieldSkipped = errors.New("field skipped")
)

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

	// Import resolution for current file being processed
	currentFileImports map[string]string // Maps package alias to full import path

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
		l:                  l,
		types:              make(map[string]*TypeInfo),
		httpOps:            make(map[string]*RouteInfo),
		mqttPublications:   make(map[string]*MQTTPublicationInfo),
		mqttSubscriptions:  make(map[string]*MQTTSubscriptionInfo),
		typeASTs:           make(map[string]*ast.GenDecl),
		constASTs:          make(map[string]*ast.GenDecl),
		currentFileImports: make(map[string]string),
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
