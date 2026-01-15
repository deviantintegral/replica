// Package migrations provides embedded SQL migration files for the replica database.
package migrations

import "embed"

// FS contains all embedded SQL migration files.
// The migrations follow the golang-migrate naming convention:
// {version}_{description}.up.sql and {version}_{description}.down.sql
//
//go:embed *.sql
var FS embed.FS
