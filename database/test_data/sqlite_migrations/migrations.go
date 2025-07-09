package sqlitetestmigrations

import "embed"

//go:embed *.sql
var Migrations embed.FS
