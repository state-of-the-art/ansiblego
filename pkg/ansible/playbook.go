package ansible

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"

	m "github.com/state-of-the-art/ansiblego/pkg/modules"
)

type Playbook struct {
	Name        string        `yaml:",omitempty"`
	Environment *m.OrderedMap `yaml:",omitempty"`

	Pre_tasks []*Task `yaml:",omitempty"`

	Tasks []*Task `yaml:",omitempty"`

	Roles []*Role `yaml:",omitempty"`

	Post_tasks []*Task `yaml:",omitempty"`
}

type PlaybookFile []Playbook

func (c *Playbook) Yaml() (string, error) {
	buf := bytes.Buffer{}
	enc := yaml.NewEncoder(&buf)
	defer enc.Close()
	enc.SetIndent(2)
	if err := enc.Encode(c); err != nil {
		return "", fmt.Errorf("YAML encode error: %v", err)
	}
	return "---\n" + buf.String(), nil
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
	buf := bytes.Buffer{}
	enc := yaml.NewEncoder(&buf)
	defer enc.Close()
	enc.SetIndent(2)
	if err := enc.Encode(c); err != nil {
		return "", fmt.Errorf("YAML encode error: %v", err)
	}
	return "---\n" + buf.String(), nil
}
