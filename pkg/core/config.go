package core

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ExtraVars []string `json:"extra_vars"` // Additional variables to use during execution
	SkipTags  []string `json:"skip_tags"`  // Skips tasks with provided tags
	Inventory []string `json:"inventory"`  // Inventory file or hosts list

	Verbose int `json:"verbose"` // Level of verbosiness, 0 is normal, >0 more verbose, <0 less
}

func (c *Config) ReadConfigFile(cfg_path string) error {
	c.initDefaults()

	if cfg_path != "" {
		// Open and parse
		data, err := ioutil.ReadFile(cfg_path)
		if err != nil {
			return err
		}

		if err := yaml.Unmarshal(data, c); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) initDefaults() {
}
