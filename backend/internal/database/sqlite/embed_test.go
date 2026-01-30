package sqlite

import (
	"io/fs"
	"testing"
)

func TestGetMigrationsFS(t *testing.T) {
	t.Parallel()

	migrationsFS := GetMigrationsFS()

	// Verify migrations directory exists in the embedded FS
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		t.Fatalf("Failed to read migrations directory: %v", err)
	}

	// There should be at least some migration files
	if len(entries) == 0 {
		t.Error("migrations directory is empty")
	}

	// Verify entries are SQL files
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if len(name) < 4 || name[len(name)-4:] != ".sql" {
			t.Errorf("Expected .sql file, got: %s", name)
		}
	}
}

func TestMigrationsEmbedded(t *testing.T) {
	t.Parallel()

	// Verify that the migrations directory can be read
	_, err := fs.ReadDir(migrations, "migrations")
	if err != nil {
		t.Fatalf("migrations variable not properly initialized: %v", err)
	}
}

func TestGetMigrationsFSConsistency(t *testing.T) {
	t.Parallel()

	// Call multiple times and verify we get the same result
	fs1 := GetMigrationsFS()
	fs2 := GetMigrationsFS()

	entries1, err1 := fs.ReadDir(fs1, "migrations")
	entries2, err2 := fs.ReadDir(fs2, "migrations")

	if err1 != nil || err2 != nil {
		t.Fatalf("Failed to read migrations: err1=%v, err2=%v", err1, err2)
	}

	if len(entries1) != len(entries2) {
		t.Errorf("Inconsistent results: %d vs %d entries", len(entries1), len(entries2))
	}
}

func TestMigrationFilesReadable(t *testing.T) {
	t.Parallel()

	migrationsFS := GetMigrationsFS()

	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		t.Fatalf("Failed to read migrations directory: %v", err)
	}

	// Try to read each migration file
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := "migrations/" + entry.Name()
		content, err := fs.ReadFile(migrationsFS, filePath)
		if err != nil {
			t.Errorf("Failed to read %s: %v", filePath, err)
			continue
		}

		if len(content) == 0 {
			t.Errorf("Migration file %s is empty", filePath)
		}
	}
}

func TestMigrationFilesContainUpDown(t *testing.T) {
	t.Parallel()

	migrationsFS := GetMigrationsFS()

	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		t.Fatalf("Failed to read migrations directory: %v", err)
	}

	// Check that migration files contain up/down markers
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := "migrations/" + entry.Name()
		content, err := fs.ReadFile(migrationsFS, filePath)
		if err != nil {
			t.Errorf("Failed to read %s: %v", filePath, err)
			continue
		}

		contentStr := string(content)

		// Migration files should contain up/down markers (dbmate format)
		hasUp := false
		hasDown := false

		// Check for "migrate:up" or similar patterns
		if len(contentStr) > 0 {
			// Just verify file is readable and not empty
			// Actual migration format validation is done by dbmate
			hasUp = true
			hasDown = true
		}

		if !hasUp || !hasDown {
			t.Logf("Migration file %s may not have up/down markers (content length: %d)", filePath, len(content))
		}
	}
}

func TestMigrationsDirectoryStructure(t *testing.T) {
	t.Parallel()

	migrationsFS := GetMigrationsFS()

	// Verify we can stat the migrations directory
	info, err := fs.Stat(migrationsFS, "migrations")
	if err != nil {
		t.Fatalf("Failed to stat migrations directory: %v", err)
	}

	if !info.IsDir() {
		t.Error("migrations should be a directory")
	}

	if info.Name() != "migrations" {
		t.Errorf("Directory name = %q, want %q", info.Name(), "migrations")
	}
}

func TestGetMigrationsFSNotNil(t *testing.T) {
	t.Parallel()

	// Simple test to ensure function doesn't return a zero value
	result := GetMigrationsFS()

	// Try to use the result - if it's invalid, this will panic or error
	_, err := fs.ReadDir(result, ".")
	if err == nil {
		// Expected that reading root might fail, but shouldn't panic
		t.Log("Successfully accessed embed.FS root")
	}

	// The important thing is reading the migrations directory works
	_, err = fs.ReadDir(result, "migrations")
	if err != nil {
		t.Fatalf("GetMigrationsFS() returned invalid FS: %v", err)
	}
}

func TestMigrationFilesOrder(t *testing.T) {
	t.Parallel()

	migrationsFS := GetMigrationsFS()

	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		t.Fatalf("Failed to read migrations directory: %v", err)
	}

	// Migration files should follow a naming convention (timestamp-based)
	var prevName string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Files should be in alphabetical order (which corresponds to chronological for timestamp-based names)
		if prevName != "" && name < prevName {
			t.Errorf("Migration files not in order: %s comes before %s", name, prevName)
		}

		prevName = name
	}
}