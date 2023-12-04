package ansible

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"

	"github.com/state-of-the-art/ansiblego/pkg/ansible/inventory"
	"github.com/state-of-the-art/ansiblego/pkg/core"
	"github.com/state-of-the-art/ansiblego/pkg/log"
)

type Playbook struct {
	Name         string      `yaml:",omitempty"`
	Gather_facts bool        `yaml:",omitempty"`
	Environment  *OrderedMap `yaml:",omitempty"`

	Pre_tasks []*Task `yaml:",omitempty"`

	Tasks []*Task `yaml:",omitempty"`

	Roles []*Role `yaml:",omitempty"`

	Post_tasks []*Task `yaml:",omitempty"`
}

type PlaybookFile []Playbook

func (p *Playbook) Yaml() (string, error) {
	return ToYaml(p)
}

func (p *Playbook) Run(cfg *core.PlaybookConfig, host *inventory.Host) (err error) {
	log.Infof("Running playbook '%s' on host '%s'...", p.Name, host.Name)

	// TODO: Connect to the host and get the facts
	facts := make(map[string]any)
	if p.Gather_facts {
		log.Error("TODO: connect to host and get facts")
	}

	// TODO: Prepare variables
	vars := make(map[string]any)
	vars["ansible_facts"] = facts
	p.fillVariables(cfg, host, vars)

	// TODO: Collect tasks & roles to run and run one-by-one
	for _, task := range p.Pre_tasks {
		if _, err = task.Run(vars); err != nil {
			return log.Errorf("Error during playbook execution: %v", err)
		}
	}
	for _, task := range p.Tasks {
		if _, err = task.Run(vars); err != nil {
			return log.Errorf("Error during playbook execution: %v", err)
		}
	}
	for _, role := range p.Roles {
		if _, err = role.Run(vars); err != nil {
			return log.Errorf("Error during playbook execution: %v", err)
		}
	}
	for _, task := range p.Post_tasks {
		if _, err = task.Run(vars); err != nil {
			return log.Errorf("Error during playbook execution: %v", err)
		}
	}

	return nil
}

// Will collect all the variables except for the facts in the right order to create vars
// https://docs.ansible.com/ansible/2.9/user_guide/playbooks_variables.html#variable-precedence-where-should-i-put-a-variable
// 01. command line values (eg “-u user”)
// 02. role defaults
// 03. inventory file or script group vars
// 04. inventory group_vars/all
// 05. playbook group_vars/all
// 06. inventory group_vars/*
// 07. playbook group_vars/*
// 08. inventory file or script host vars
// 09. inventory host_vars/*
// 10. playbook host_vars/*
// 11. host facts / cached set_facts
// 12. play vars
// 13. play vars_prompt
// 14. play vars_files
// 15. role vars (defined in role/vars/main.yml)
// 16. block vars (only for tasks in block)
// 17. task vars (only for the task)
// 18. include_vars
// 19. set_facts / registered vars
// 20. role (and include_role) params
// 21. include params
// 22. extra vars (always win precedence)
func (p *Playbook) fillVariables(cfg *core.PlaybookConfig, host *inventory.Host, vars map[string]any) {
	// 03. Adding host variables from inventory
	for key, val := range host.Vars {
		vars[key] = val
	}
	// 22. Setting extra vars
	for key, val := range cfg.ExtraVars {
		vars[key] = val
	}
}

func (pf *PlaybookFile) Load(yml_path string) error {
	// Open and parse
	data, err := ioutil.ReadFile(yml_path)
	if err != nil {
		return err
	}

	return pf.Parse(data)
}

func (pf *PlaybookFile) Parse(data []byte) error {
	if err := yaml.Unmarshal(data, pf); err != nil {
		return err
	}

	return nil
}

func (pf *PlaybookFile) Yaml() (string, error) {
	return ToYaml(pf)
}
