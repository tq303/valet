package config

type Tool string

const (
	ToolClaude Tool = "claude"
	ToolCursor Tool = "cursor"
)

type Package struct {
	Path    string   `yaml:"path"`
	Context string   `yaml:"context"`
	Tools   []Tool   `yaml:"tools"`
}

type Config struct {
	Version  string    `yaml:"version"`
	Packages []Package `yaml:"packages"`
}
