package include_role

// Doc: https://docs.ansible.com/ansible/2.9/modules/include_role_module.html

import (
	"fmt"

	"github.com/state-of-the-art/ansiblego/pkg/ansible"
	"github.com/state-of-the-art/ansiblego/pkg/log"
)

type TaskV1 struct {
	// The name of the role to be executed.
	Name string `task:",req"`

	// Overrides the role's metadata setting to allow using a role more than once with the same parameters.
	//Allow_duplicates bool `task:",def:true"`

	// Accepts a hash of task keywords (e.g. tags, become) that will be applied to all tasks within the included role.
	//Apply map[string]any

	// File to load from a role's defaults/ directory.
	//Defaults_from string `task:",def:main"`
	// File to load from a role's handlers/ directory.
	//Handlers_from string `task:",def:main"`
	// File to load from a role's tasks/ directory.
	//Tasks_from    string `task:",def:main"`
	// File to load from a role's vars/ directory.
	//Vars_from     string `task:",def:main"`

	// This option dictates whether the role's vars and defaults are exposed to the playbook.
	//Public bool
}

func (t *TaskV1) SetData(data ansible.OrderedMap) error {
	role_data, ok := data.Get("include_role")
	if !ok {
		return fmt.Errorf("Unable to find the 'include_role' map in task data")
	}
	fmap, ok := role_data.(ansible.OrderedMap)
	if !ok {
		return fmt.Errorf("The 'include_role' is not the OrderedMap")
	}
	return ansible.TaskV1SetData(t, fmap)
}

func (t *TaskV1) GetData() (data ansible.OrderedMap) {
	fmap := ansible.TaskV1GetData(t)
	data.Set("include_role", fmap)
	return data
}

func (t *TaskV1) Run(vars map[string]any) (out ansible.OrderedMap, err error) {
	log.Error("TODO: Implement include_role.Run")

	return
}
