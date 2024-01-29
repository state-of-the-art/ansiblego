package setup

// Doc: https://docs.ansible.com/ansible/2.9/modules/setup_module.html

import (
	"fmt"

	"github.com/state-of-the-art/ansiblego/pkg/ansible"
	"github.com/state-of-the-art/ansiblego/pkg/log"
)

type TaskV1 struct {
	// Path used for local ansible facts (*.fact) - files in this dir will be run (if executable) and their results be added to ansible_local facts if a file is not executable it is read.
	//Fact_path string `task:"def:/etc/ansible/facts.d"`
	// If supplied, only return facts that match this shell-style (fnmatch) wildcard.
	//Filter string `task:"def:*"`
	// If supplied, restrict the additional facts collected to the given subset. Possible values: all, min, hardware, network, virtual, ohai, and facter.
	//Father_subset string `task:"def:all"`
	// Set the default timeout in seconds for individual fact gathering.
	//Gather_timeout int `task:"def:10"`
}

// Here the fields comes as complete values never as jinja2 templates
func (t *TaskV1) SetData(data *ansible.OrderedMap) error {
	_, ok := data.Pop("setup")
	if !ok {
		return fmt.Errorf("Unable to find 'setup' key in task data")
	}
	/*fmap, ok := d.(ansible.OrderedMap)
	if !ok {
		return fmt.Errorf("The 'setup' is not the OrderedMap")
	}*/

	return nil
}

func (t *TaskV1) GetData() (data ansible.OrderedMap) {
	return data
}

func (t *TaskV1) Run(vars map[string]any) (out ansible.OrderedMap, err error) {
	for _, mod := range ansible.ModulesList("fact") {
		log.Tracef("Running fact collector %q...", mod)
		if collected_facts, err := ansible.CollectV1(mod); err == nil {
			for _, key := range collected_facts.Keys() {
				log.Tracef("Received data key from fact collector %q: %q", mod, key)
				if _, exists := out.Get(key); exists {
					log.Warnf("Fact module '%s' overrides existing fact '%s'", mod, key)
				}
				val, _ := collected_facts.Get(key)
				out.Set(key, val)
			}
		} else {
			// Fact collectors can fail, but it should not be a disaster
			err = log.Errorf("Error while collecting facts from %q: %v", mod, err)
		}
	}

	return out, err
}
