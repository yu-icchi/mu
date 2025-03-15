package artifact

type Action struct {
	Runs Runs `yaml:"runs"`
}

type Runs struct {
	Using string  `yaml:"using"`
	Steps []Setup `yaml:"steps"`
}

type Setup struct {
	Name string      `yaml:"name"`
	Uses string      `yaml:"uses"`
	With WithOptions `yaml:"with,omitempty"`
}

type WithOptions interface {
	withType() string
}
