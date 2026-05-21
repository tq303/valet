package config

type Tool string

const (
	ToolClaude Tool = "claude"
	ToolCursor Tool = "cursor"
)

type Rule struct {
	File    string   `yaml:"file"`
	Dest    string   `yaml:"dest,omitempty"`
	Exclude []string `yaml:"exclude,omitempty"`
	Tools   []Tool   `yaml:"tools,omitempty"`
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
