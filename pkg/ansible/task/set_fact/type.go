package set_fact

// Doc: https://docs.ansible.com/ansible/2.9/modules/set_fact_module.html

import (
	"fmt"
	"log"

	"github.com/state-of-the-art/ansiblego/pkg/ansible"
)

type TaskV1 struct {
	// Storage for key:value to process
	keyval ansible.OrderedMap
}

func (t *TaskV1) SetData(data ansible.OrderedMap) error {
	role_data, ok := data.Get("set_fact")
	if !ok {
		return fmt.Errorf("Unable to find the 'set_fact' map in task data")
	}
	fmap, ok := role_data.(ansible.OrderedMap)
	if !ok {
		return fmt.Errorf("The 'set_fact' is not the OrderedMap")
	}

	if fmap.Size() < 1 {
		return fmt.Errorf("The 'set_fact' data is empty")
	}

	// Store data for further processing during Run
	t.keyval = fmap

	return nil
}

func (t *TaskV1) GetData() (data ansible.OrderedMap) {
	data.Set("set_fact", t.keyval)
	return data
}

func (t *TaskV1) Run(vars map[string]any) error {
	log.Println("TODO: Implement set_fact.Run")

	return nil
}