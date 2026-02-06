package sqlite

import "embed"

//go:embed migrations/*.sql
var migrations embed.FS

func GetMigrationsFS() embed.FS {
	return migrations
}
