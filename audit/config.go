package audit

import (
	"github.com/go-faster/yaml"
	autoconfigure "github.com/larsartmann/linter-autoconfigure-sdk"
)

// golangciConfig models the parts of a golangci-lint v2 config the drift
// checks read: the linters and formatters sections' enable, disable,
// settings, and exclusions keys. Everything else is ignored on purpose; the
// audit never needs to interpret it.
type golangciConfig struct {
	Linters    lintersSection    `yaml:"linters"`
	Formatters formattersSection `yaml:"formatters"`
}

type lintersSection struct {
	Enable     []string         `yaml:"enable"`
	Disable    []string         `yaml:"disable"`
	Settings   map[string]any   `yaml:"settings"`
	Exclusions exclusionsStruct `yaml:"exclusions"`
}

type exclusionsStruct struct {
	Rules []exclusionRule `yaml:"rules"`
}

type exclusionRule struct {
	Path    string   `yaml:"path"`
	Linters []string `yaml:"linters"`
}

type formattersSection struct {
	Enable   []string       `yaml:"enable"`
	Disable  []string       `yaml:"disable"`
	Settings map[string]any `yaml:"settings"`
}

// loadConfig reads and parses the golangci-lint config at path. Failures
// surface as *autoconfigure.ConfigError (Op read or unmarshal), so callers
// can branch with errors.Is / errors.AsType like on every other SDK config
// helper.
func loadConfig(path string) (golangciConfig, *autoconfigure.ConfigError) {
	data, cfgErr := autoconfigure.ReadConfig(path)
	if cfgErr != nil {
		return golangciConfig{}, cfgErr
	}

	var cfg golangciConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return golangciConfig{}, &autoconfigure.ConfigError{
			Op:   autoconfigure.OpUnmarshal,
			Path: path,
			Err:  err,
		}
	}

	return cfg, nil
}

// validLinters returns the set of names that may legitimately be
// referenced anywhere in the config or in nolint directives: linters and
// formatters enabled or explicitly disabled in the config, plus the given
// default-enabled linters.
func (c golangciConfig) validLinters(defaults []string) map[string]bool {
	valid := make(map[string]bool)

	for _, name := range c.Linters.Enable {
		valid[name] = true
	}

	for _, name := range c.Linters.Disable {
		valid[name] = true
	}

	for _, name := range defaults {
		valid[name] = true
	}

	for _, name := range c.Formatters.Enable {
		valid[name] = true
	}

	for _, name := range c.Formatters.Disable {
		valid[name] = true
	}

	return valid
}
