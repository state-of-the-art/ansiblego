package core

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"

	"github.com/state-of-the-art/ansiblego/pkg/ansible/inventory"
)

// Common configuration for the framework
type CommonConfig struct {
	Verbosity uint8 `json:"verbosity"` // Level of verbosiness
}

// Playbook execution config
type PlaybookConfig struct {
	CommonConfig

	// Additional variables to use during execution
	ExtraVars map[string]any `json:"extra_vars"`
	// Skips tasks with provided tags
	SkipTags []string `json:"skip_tags"`

	// Parsed inventory data
	Inventory *inventory.Inventory `json:"inventory"`
}

// Agent execution config
type AgentConfig struct {
	CommonConfig
}

func ReadConfigFile(obj any, cfg_path string) error {
	if cfg_path != "" {
		// Open and parse
		data, err := ioutil.ReadFile(cfg_path)
		if err != nil {
			return err
		}

		if err := yaml.Unmarshal(data, obj); err != nil {
			return err
		}
	}

	return nil
}
