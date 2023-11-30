package ansible

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"

	"github.com/state-of-the-art/ansiblego/pkg/log"
)

type Role struct {
	Name        string      `yaml:"role"`
	Environment *OrderedMap `yaml:",omitempty"`
	Vars        *OrderedMap `yaml:",omitempty"`
}

func (r *Role) Load(yml_path string) error {
	// Open and parse
	data, err := ioutil.ReadFile(yml_path)
	if err != nil {
		return err
	}

	return r.Parse(data)
}

func (r *Role) Parse(data []byte) error {
	if err := yaml.Unmarshal(data, r); err != nil {
		return err
	}

	return nil
}

func (r *Role) Yaml() (string, error) {
	return ToYaml(r)
}

func (r *Role) Run(vars map[string]any) error {
	log.Infof("Executing role '%s'", r.Name)
	return nil
}
