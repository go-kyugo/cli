//go:build tools
// +build tools

package tools

// This file pins CLI tool dependencies so they are recorded in go.mod.
// Import tools here with a blank identifier.

import (
	_ "github.com/swaggo/swag"
)
