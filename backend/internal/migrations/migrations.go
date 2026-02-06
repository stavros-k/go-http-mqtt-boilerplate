package migrations

import (
	"embed"
)

// migrationsFS embeds all SQL migration files from the specified directories.
// Structure:
//
//	.
//	|-- shared
//	|   |-- migrations
//	|       |-- *.sql
//	|-- local
//	|   |-- migrations
//	|       |-- *.sql
//	|-- cloud
//	|   |-- migrations
//	|       |-- *.sql
//
//go:embed shared/migrations/*.sql local/migrations/*.sql cloud/migrations/*.sql
var migrationsFS embed.FS

func GetFS() embed.FS {
	return migrationsFS
}
