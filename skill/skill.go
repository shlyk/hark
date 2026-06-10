// Package skill embeds the hark agent skill that "hark sync" installs into
// Claude Code and Codex skill directories.
package skill

import _ "embed"

//go:embed SKILL.md
var Content []byte
