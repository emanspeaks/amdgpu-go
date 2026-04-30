package amdgpu

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var versionFile string

// Version is the semantic version of this package, read from the VERSION file.
var Version = strings.TrimSpace(versionFile)
