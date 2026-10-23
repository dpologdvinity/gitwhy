// Package gitwhy holds the embedded docs and recipes.
package gitwhy

import "embed"

// Content holds docs/commands, docs/shared and recipes.
//
//go:embed docs/commands docs/shared recipes
var Content embed.FS
