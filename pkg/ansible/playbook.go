package ansible

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Playbook struct {
	Name        string      `yaml:",omitempty"`
	Environment *OrderedMap `yaml:",omitempty"`

	Pre_tasks []*Task `yaml:",omitempty"`

	Tasks []*Task `yaml:",omitempty"`

	Roles []*Role `yaml:",omitempty"`

	Post_tasks []*Task `yaml:",omitempty"`
}

type PlaybookFile []Playbook

func (c *Playbook) Yaml() (string, error) {
	return ToYaml(c)
}

func (c *PlaybookFile) Load(yml_path string) error {
	// Open and parse
	data, err := ioutil.ReadFile(yml_path)
	if err != nil {
		return err
	}

	return c.Parse(data)
}

func (c *PlaybookFile) Parse(data []byte) error {
	if err := yaml.Unmarshal(data, c); err != nil {
		return err
	}

	return nil
}

func (c *PlaybookFile) Yaml() (string, error) {
	return ToYaml(c)
}
