//go:build cgo
// +build cgo

package generate

import (
	"strings"
	"testing"
)

func TestGetDatabaseStats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		schema         string
		wantTableCount int
	}{
		{
			name: "single table",
			schema: `CREATE TABLE users (
				id INTEGER PRIMARY KEY,
				name TEXT NOT NULL
			);`,
			wantTableCount: 1,
		},
		{
			name: "multiple tables",
			schema: `CREATE TABLE users (id INTEGER);
			CREATE TABLE posts (id INTEGER);
			CREATE TABLE comments (id INTEGER);`,
			wantTableCount: 3,
		},
		{
			name:           "no tables",
			schema:         "-- Just a comment",
			wantTableCount: 0,
		},
		{
			name:           "empty schema",
			schema:         "",
			wantTableCount: 0,
		},
		{
			name: "case sensitivity",
			schema: `create table lowercase (id int);
			CREATE TABLE uppercase (id int);`,
			wantTableCount: 1, // Only counts "CREATE TABLE" (uppercase)
		},
		{
			name: "mixed with other statements",
			schema: `CREATE INDEX idx_user ON users(name);
			CREATE TABLE users (id INTEGER);
			CREATE VIEW user_view AS SELECT * FROM users;
			CREATE TABLE posts (id INTEGER);`,
			wantTableCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a minimal collector (we only need the method)
			g := &OpenAPICollector{}

			stats, err := g.GetDatabaseStats(tt.schema)
			if err != nil {
				t.Fatalf("GetDatabaseStats() error = %v", err)
			}

			if stats.TableCount != tt.wantTableCount {
				t.Errorf("GetDatabaseStats() TableCount = %d, want %d", stats.TableCount, tt.wantTableCount)
			}
		})
	}
}

func TestGetDatabaseStatsReturnValue(t *testing.T) {
	t.Parallel()

	g := &OpenAPICollector{}
	schema := "CREATE TABLE test (id INTEGER);"

	stats, err := g.GetDatabaseStats(schema)
	if err != nil {
		t.Fatalf("GetDatabaseStats() error = %v", err)
	}

	if stats == nil {
		t.Fatal("GetDatabaseStats() returned nil stats")
	}

	// Verify the struct is properly populated
	if stats.TableCount != 1 {
		t.Errorf("Expected TableCount = 1, got %d", stats.TableCount)
	}
}

func TestGetDatabaseStatsLargeSchema(t *testing.T) {
	t.Parallel()

	// Build a large schema with many tables
	var schemaBuilder strings.Builder
	tableCount := 100

	for i := 0; i < tableCount; i++ {
		schemaBuilder.WriteString("CREATE TABLE table")
		schemaBuilder.WriteString(string(rune('0' + i%10)))
		schemaBuilder.WriteString(" (id INTEGER);\n")
	}

	g := &OpenAPICollector{}
	stats, err := g.GetDatabaseStats(schemaBuilder.String())
	if err != nil {
		t.Fatalf("GetDatabaseStats() error = %v", err)
	}

	if stats.TableCount != tableCount {
		t.Errorf("GetDatabaseStats() TableCount = %d, want %d", stats.TableCount, tableCount)
	}
}

func TestDatabaseStatsStruct(t *testing.T) {
	t.Parallel()

	// Test that DatabaseStats struct has correct fields
	stats := DatabaseStats{
		TableCount: 42,
	}

	if stats.TableCount != 42 {
		t.Errorf("DatabaseStats.TableCount = %d, want 42", stats.TableCount)
	}
}

func TestGetDatabaseStatsWithComplexSQL(t *testing.T) {
	t.Parallel()

	schema := `
		-- User table
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		-- Posts table with foreign key
		CREATE TABLE posts (
			id INTEGER PRIMARY KEY,
			user_id INTEGER,
			title TEXT NOT NULL,
			content TEXT,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

		-- Index on posts
		CREATE INDEX idx_posts_user ON posts(user_id);

		-- Comments table
		CREATE TABLE comments (
			id INTEGER PRIMARY KEY,
			post_id INTEGER,
			user_id INTEGER,
			comment TEXT NOT NULL,
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);

		-- View
		CREATE VIEW user_posts AS
		SELECT u.username, p.title
		FROM users u
		JOIN posts p ON u.id = p.user_id;
	`

	g := &OpenAPICollector{}
	stats, err := g.GetDatabaseStats(schema)
	if err != nil {
		t.Fatalf("GetDatabaseStats() error = %v", err)
	}

	expectedCount := 3 // users, posts, comments (not the view or index)
	if stats.TableCount != expectedCount {
		t.Errorf("GetDatabaseStats() TableCount = %d, want %d", stats.TableCount, expectedCount)
	}
}