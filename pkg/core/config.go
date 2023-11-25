package core

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

// Common configuration for the framework
type CommonConfig struct {
	Verbosity uint8 `json:"verbosity"` // Level of verbosiness
}

// Playbook execution config
type PlaybookConfig struct {
	CommonConfig

	ExtraVars []string `json:"extra_vars"` // Additional variables to use during execution
	SkipTags  []string `json:"skip_tags"`  // Skips tasks with provided tags
	Inventory []string `json:"inventory"`  // Inventory file or hosts list
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
