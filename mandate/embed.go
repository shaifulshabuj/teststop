package mandate

import _ "embed"

//go:embed base.md
var BaseMandateContent string

// Base is the embedded mandate content, kept for backward compatibility with CLI commands.
var Base = BaseMandateContent
