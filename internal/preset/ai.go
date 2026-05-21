package preset

// AI tool identifiers used in .rules subdirectory conventions and preset resolution.
const (
	Claude = "claude"
	Cursor = "cursor"
)

// Dest returns the destination path for a given AI tool preset and filename.
func Dest(tool, filename string) string {
	switch tool {
	case Claude:
		return ".claude/CLAUDE.md"
	case Cursor:
		return ".cursor/rules/" + filename
	default:
		return filename
	}
}
