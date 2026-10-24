// Package gitwhy holds the embedded docs and recipes.
package gitwhy

import "embed"

// Content holds docs/commands, docs/shared and recipes. ui.New reads it at start-up, so the
// binary works offline and ships as one file.
//
//go:embed docs/commands docs/shared recipes
var Content embed.FS
