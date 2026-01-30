package generate

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateAPIDocs(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name    string
		doc     *APIDocumentation
		wantErr bool
	}{
		{
			name: "valid documentation",
			doc: &APIDocumentation{
				Info: APIInfo{
					Title:       "Test API",
					Version:     "1.0.0",
					Description: "Test Description",
				},
				Types: map[string]*TypeInfo{
					"TestType": {
						Name: "TestType",
						Kind: TypeKindObject,
					},
				},
				HTTPOperations:    map[string]*RouteInfo{},
				MQTTPublications:  map[string]*MQTTPublicationInfo{},
				MQTTSubscriptions: map[string]*MQTTSubscriptionInfo{},
				Database: Database{
					Schema:     "CREATE TABLE test (id INTEGER);",
					TableCount: 1,
				},
				OpenAPISpec: "openapi: 3.0.0",
			},
			wantErr: false,
		},
		{
			name: "empty documentation",
			doc: &APIDocumentation{
				Types:             map[string]*TypeInfo{},
				HTTPOperations:    map[string]*RouteInfo{},
				MQTTPublications:  map[string]*MQTTPublicationInfo{},
				MQTTSubscriptions: map[string]*MQTTSubscriptionInfo{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			outputPath := filepath.Join(tmpDir, "docs.json")

			err := GenerateAPIDocs(logger, tt.doc, outputPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateAPIDocs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify file was created
				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					t.Errorf("GenerateAPIDocs() did not create file at %s", outputPath)
				}

				// Verify file is not empty
				info, err := os.Stat(outputPath)
				if err != nil {
					t.Fatalf("Failed to stat output file: %v", err)
				}
				if info.Size() == 0 {
					t.Error("GenerateAPIDocs() created empty file")
				}

				// Verify file contains valid JSON (basic check)
				data, err := os.ReadFile(outputPath)
				if err != nil {
					t.Fatalf("Failed to read output file: %v", err)
				}
				if len(data) == 0 {
					t.Error("GenerateAPIDocs() wrote no data")
				}
			}
		})
	}
}

func TestGenerateAPIDocsInvalidPath(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	doc := &APIDocumentation{
		Types:             map[string]*TypeInfo{},
		HTTPOperations:    map[string]*RouteInfo{},
		MQTTPublications:  map[string]*MQTTPublicationInfo{},
		MQTTSubscriptions: map[string]*MQTTSubscriptionInfo{},
	}

	// Try to write to invalid path
	invalidPath := "/invalid/path/that/does/not/exist/docs.json"
	err := GenerateAPIDocs(logger, doc, invalidPath)
	if err == nil {
		t.Error("GenerateAPIDocs() should fail with invalid path")
	}
}

func TestSortDocumentation(t *testing.T) {
	t.Parallel()

	doc := &APIDocumentation{
		Types: map[string]*TypeInfo{
			"Type1": {
				Name:         "Type1",
				References:   []string{"Type3", "Type2"},
				ReferencedBy: []string{"Type4", "Type2"},
				UsedBy: []UsageInfo{
					{OperationID: "op2", Role: "request"},
					{OperationID: "op1", Role: "response"},
					{OperationID: "op1", Role: "request"},
				},
			},
		},
	}

	sortDocumentation(doc)

	type1 := doc.Types["Type1"]

	// Check References are sorted
	if len(type1.References) != 2 {
		t.Fatalf("Expected 2 references, got %d", len(type1.References))
	}
	if type1.References[0] != "Type2" || type1.References[1] != "Type3" {
		t.Errorf("References not sorted correctly: %v", type1.References)
	}

	// Check ReferencedBy are sorted
	if len(type1.ReferencedBy) != 2 {
		t.Fatalf("Expected 2 referencedBy, got %d", len(type1.ReferencedBy))
	}
	if type1.ReferencedBy[0] != "Type2" || type1.ReferencedBy[1] != "Type4" {
		t.Errorf("ReferencedBy not sorted correctly: %v", type1.ReferencedBy)
	}

	// Check UsedBy are sorted
	if len(type1.UsedBy) != 3 {
		t.Fatalf("Expected 3 usedBy, got %d", len(type1.UsedBy))
	}
	// Should be sorted by OperationID, then Role
	if type1.UsedBy[0].OperationID != "op1" || type1.UsedBy[0].Role != "request" {
		t.Errorf("UsedBy[0] not sorted correctly: %v", type1.UsedBy[0])
	}
	if type1.UsedBy[1].OperationID != "op1" || type1.UsedBy[1].Role != "response" {
		t.Errorf("UsedBy[1] not sorted correctly: %v", type1.UsedBy[1])
	}
	if type1.UsedBy[2].OperationID != "op2" || type1.UsedBy[2].Role != "request" {
		t.Errorf("UsedBy[2] not sorted correctly: %v", type1.UsedBy[2])
	}
}

func TestSortUsageInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []UsageInfo
		want  []UsageInfo
	}{
		{
			name: "sort by operationID",
			input: []UsageInfo{
				{OperationID: "op3", Role: "request"},
				{OperationID: "op1", Role: "request"},
				{OperationID: "op2", Role: "request"},
			},
			want: []UsageInfo{
				{OperationID: "op1", Role: "request"},
				{OperationID: "op2", Role: "request"},
				{OperationID: "op3", Role: "request"},
			},
		},
		{
			name: "sort by role when operationID same",
			input: []UsageInfo{
				{OperationID: "op1", Role: "response"},
				{OperationID: "op1", Role: "parameter"},
				{OperationID: "op1", Role: "request"},
			},
			want: []UsageInfo{
				{OperationID: "op1", Role: "parameter"},
				{OperationID: "op1", Role: "request"},
				{OperationID: "op1", Role: "response"},
			},
		},
		{
			name: "mixed sort",
			input: []UsageInfo{
				{OperationID: "op2", Role: "response"},
				{OperationID: "op1", Role: "response"},
				{OperationID: "op1", Role: "request"},
			},
			want: []UsageInfo{
				{OperationID: "op1", Role: "request"},
				{OperationID: "op1", Role: "response"},
				{OperationID: "op2", Role: "response"},
			},
		},
		{
			name:  "empty slice",
			input: []UsageInfo{},
			want:  []UsageInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Make a copy to avoid modifying test data
			input := make([]UsageInfo, len(tt.input))
			copy(input, tt.input)

			sortUsageInfo(input)

			if len(input) != len(tt.want) {
				t.Fatalf("Length mismatch: got %d, want %d", len(input), len(tt.want))
			}

			for i := range input {
				if input[i] != tt.want[i] {
					t.Errorf("At index %d: got %v, want %v", i, input[i], tt.want[i])
				}
			}
		})
	}
}

func TestGenerateAPIDocsDeterministic(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

	doc := &APIDocumentation{
		Info: APIInfo{Title: "Test", Version: "1.0.0"},
		Types: map[string]*TypeInfo{
			"Type1": {
				Name:       "Type1",
				References: []string{"Type3", "Type2"},
			},
			"Type2": {
				Name: "Type2",
			},
		},
		HTTPOperations:    map[string]*RouteInfo{},
		MQTTPublications:  map[string]*MQTTPublicationInfo{},
		MQTTSubscriptions: map[string]*MQTTSubscriptionInfo{},
	}

	tmpDir := t.TempDir()
	output1 := filepath.Join(tmpDir, "docs1.json")
	output2 := filepath.Join(tmpDir, "docs2.json")

	// Generate twice
	if err := GenerateAPIDocs(logger, doc, output1); err != nil {
		t.Fatalf("First GenerateAPIDocs() failed: %v", err)
	}

	if err := GenerateAPIDocs(logger, doc, output2); err != nil {
		t.Fatalf("Second GenerateAPIDocs() failed: %v", err)
	}

	// Read both files
	data1, err := os.ReadFile(output1)
	if err != nil {
		t.Fatalf("Failed to read first output: %v", err)
	}

	data2, err := os.ReadFile(output2)
	if err != nil {
		t.Fatalf("Failed to read second output: %v", err)
	}

	// Compare
	if string(data1) != string(data2) {
		t.Error("GenerateAPIDocs() should produce deterministic output")
	}
}