// Package mandate exposes the canonical adversarial-user mandate as an
// embedded asset so the teststop binary remains a single, zero-config artifact.
//
// The canonical source of truth is mandate/base.md. Edit that file; never edit
// generated copies. Changes propagate to the binary at build time via go:embed.
package mandate

import _ "embed"

// Base is the canonical mandate text (mandate/base.md), embedded at build time.
//
//go:embed base.md
var Base string
