package ansible

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"

	"github.com/state-of-the-art/ansiblego/pkg/ansible/inventory"
	"github.com/state-of-the-art/ansiblego/pkg/log"
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

func (p *Playbook) Yaml() (string, error) {
	return ToYaml(p)
}

func (p *Playbook) Run(host *inventory.Host) error {
	log.Infof("Running playbook '%s' on host '%s'...", p.Name, host.Name)

	// TODO: Connect to the host
	facts := make(map[string]any)

	// TODO: Collect facts & prepare variables
	vars := make(map[string]any)
	vars["facts"] = facts

	// TODO: Collect tasks & roles to run and run one-by-one
	for _, task := range p.Pre_tasks {
		if err := task.Run(vars); err != nil {
			return log.Errorf("Error during playbook execution: %v", err)
		}
	}
	for _, task := range p.Tasks {
		if err := task.Run(vars); err != nil {
			return log.Errorf("Error during playbook execution: %v", err)
		}
	}
	for _, role := range p.Roles {
		if err := role.Run(vars); err != nil {
			return log.Errorf("Error during playbook execution: %v", err)
		}
	}
	for _, task := range p.Post_tasks {
		if err := task.Run(vars); err != nil {
			return log.Errorf("Error during playbook execution: %v", err)
		}
	}

	return nil
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
