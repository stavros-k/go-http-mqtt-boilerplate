//go:build cgo
// +build cgo

package migrator

import (
	"embed"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//go:embed testdata/*.sql
var testMigrations embed.FS

func TestNew(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	t.Run("valid migrator", func(t *testing.T) {
		t.Parallel()

		tmpFile := filepath.Join(t.TempDir(), "test.db")

		m, err := New(logger, testMigrations, tmpFile)
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		if m == nil {
			t.Fatal("New() returned nil")
		}

		if m.db == nil {
			t.Error("Migrator.db should not be nil")
		}

		if m.l == nil {
			t.Error("Migrator.l should not be nil")
		}

		if m.sqlPath != tmpFile {
			t.Errorf("Migrator.sqlPath = %q, want %q", m.sqlPath, tmpFile)
		}
	})

	t.Run("empty sqlPath", func(t *testing.T) {
		t.Parallel()

		_, err := New(logger, testMigrations, "")
		if err == nil {
			t.Error("New() should return error for empty sqlPath")
		}

		if !strings.Contains(err.Error(), "sqlPath is required") {
			t.Errorf("Expected 'sqlPath is required' error, got: %v", err)
		}
	})

	t.Run("invalid embed fs", func(t *testing.T) {
		t.Parallel()

		var emptyFS embed.FS
		tmpFile := filepath.Join(t.TempDir(), "test.db")

		_, err := New(logger, emptyFS, tmpFile)
		if err == nil {
			t.Error("New() should return error for embed.FS without migrations directory")
		}
	})
}

func TestMigrator_Migrate(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	t.Run("successful migration", func(t *testing.T) {
		t.Parallel()

		tmpFile := filepath.Join(t.TempDir(), "test.db")

		m, err := New(logger, testMigrations, tmpFile)
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		err = m.Migrate()
		if err != nil {
			t.Fatalf("Migrate() error = %v", err)
		}

		// Verify database file was created
		if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
			t.Error("Migrate() did not create database file")
		}
	})

	t.Run("migrate twice idempotent", func(t *testing.T) {
		t.Parallel()

		tmpFile := filepath.Join(t.TempDir(), "test.db")

		m, err := New(logger, testMigrations, tmpFile)
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		// First migration
		err = m.Migrate()
		if err != nil {
			t.Fatalf("First Migrate() error = %v", err)
		}

		// Second migration (should be idempotent)
		err = m.Migrate()
		if err != nil {
			t.Fatalf("Second Migrate() error = %v", err)
		}
	})
}

func TestMigrator_DumpSchema(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	t.Run("successful dump", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		dbFile := filepath.Join(tmpDir, "test.db")
		schemaFile := filepath.Join(tmpDir, "schema.sql")

		m, err := New(logger, testMigrations, dbFile)
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		// First migrate
		err = m.Migrate()
		if err != nil {
			t.Fatalf("Migrate() error = %v", err)
		}

		// Then dump schema
		err = m.DumpSchema(schemaFile)
		if err != nil {
			t.Fatalf("DumpSchema() error = %v", err)
		}

		// Verify schema file was created
		if _, err := os.Stat(schemaFile); os.IsNotExist(err) {
			t.Error("DumpSchema() did not create schema file")
		}

		// Read and verify schema content
		content, err := os.ReadFile(schemaFile)
		if err != nil {
			t.Fatalf("Failed to read schema file: %v", err)
		}

		if len(content) == 0 {
			t.Error("DumpSchema() created empty schema file")
		}
	})

	t.Run("dump without migrate", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		dbFile := filepath.Join(tmpDir, "test.db")
		schemaFile := filepath.Join(tmpDir, "schema.sql")

		m, err := New(logger, testMigrations, dbFile)
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		// Try to dump without migrating first
		err = m.DumpSchema(schemaFile)
		if err != nil {
			t.Fatalf("DumpSchema() error = %v", err)
		}

		// Should still create a file (empty or minimal)
		if _, err := os.Stat(schemaFile); os.IsNotExist(err) {
			t.Error("DumpSchema() did not create schema file")
		}
	})
}

func TestMigratorFields(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tmpFile := filepath.Join(t.TempDir(), "test.db")

	m, err := New(logger, testMigrations, tmpFile)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Verify all expected fields are set
	if m.db == nil {
		t.Error("db field should not be nil")
	}

	if m.fs == (embed.FS{}) {
		t.Error("fs field should not be empty")
	}

	if m.sqlPath == "" {
		t.Error("sqlPath field should not be empty")
	}

	if m.l == nil {
		t.Error("l (logger) field should not be nil")
	}
}

func TestMigratorWithInvalidPath(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Try to create migrator with invalid path characters
	invalidPath := string([]byte{0})

	_, err := New(logger, testMigrations, invalidPath)
	// This might succeed or fail depending on OS - just verify it doesn't panic
	if err != nil {
		// Expected for some systems
		t.Logf("New() with invalid path returned error: %v", err)
	}
}

func TestMigratorConcurrentAccess(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tmpFile := filepath.Join(t.TempDir(), "test.db")

	m, err := New(logger, testMigrations, tmpFile)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// First migration
	if err := m.Migrate(); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	// Try to dump schema multiple times concurrently
	done := make(chan bool, 3)
	errs := make(chan error, 3)

	for i := 0; i < 3; i++ {
		go func(n int) {
			tmpDir := t.TempDir()
			schemaFile := filepath.Join(tmpDir, "schema.sql")
			if err := m.DumpSchema(schemaFile); err != nil {
				errs <- err
			}
			done <- true
		}(i)
	}

	// Wait for all to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	close(errs)
	for err := range errs {
		t.Errorf("Concurrent DumpSchema() error: %v", err)
	}
}

func TestMigratorDumpSchemaOverwrite(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tmpDir := t.TempDir()
	dbFile := filepath.Join(tmpDir, "test.db")
	schemaFile := filepath.Join(tmpDir, "schema.sql")

	m, err := New(logger, testMigrations, dbFile)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Migrate
	if err := m.Migrate(); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	// First dump
	if err := m.DumpSchema(schemaFile); err != nil {
		t.Fatalf("First DumpSchema() error = %v", err)
	}

	firstContent, err := os.ReadFile(schemaFile)
	if err != nil {
		t.Fatalf("Failed to read first schema: %v", err)
	}

	// Second dump (should overwrite)
	if err := m.DumpSchema(schemaFile); err != nil {
		t.Fatalf("Second DumpSchema() error = %v", err)
	}

	secondContent, err := os.ReadFile(schemaFile)
	if err != nil {
		t.Fatalf("Failed to read second schema: %v", err)
	}

	// Content should be the same (deterministic)
	if string(firstContent) != string(secondContent) {
		t.Error("DumpSchema() should produce consistent output")
	}
}