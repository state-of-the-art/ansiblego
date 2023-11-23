package ansible

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Role struct {
	Name        string      `yaml:"role"`
	Environment *OrderedMap `yaml:",omitempty"`
	Vars        *OrderedMap `yaml:",omitempty"`
}

func (c *Role) Load(yml_path string) error {
	// Open and parse
	data, err := ioutil.ReadFile(yml_path)
	if err != nil {
		return err
	}

	return c.Parse(data)
}

func (c *Role) Parse(data []byte) error {
	if err := yaml.Unmarshal(data, c); err != nil {
		return err
	}

	return nil
}

func (c *Role) Yaml() (string, error) {
	return ToYaml(c)
}
