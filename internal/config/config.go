package config

type Tool string

const (
	ToolClaude Tool = "claude"
	ToolCursor Tool = "cursor"
)

type Rule struct {
	File  string `yaml:"file"`
	Tools []Tool `yaml:"tools"`
}

type Package struct {
	Path    string `yaml:"path"`
	Context string `yaml:"context,omitempty"`
}

type Config struct {
	Version  int       `yaml:"version"`
	Rules    []Rule    `yaml:"rules"`
	Packages []Package `yaml:"packages"`
}
