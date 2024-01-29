package set_fact

// Doc: https://docs.ansible.com/ansible/2.9/modules/set_fact_module.html

import (
	"fmt"

	"github.com/state-of-the-art/ansiblego/pkg/ansible"
	"github.com/state-of-the-art/ansiblego/pkg/log"
)

type TaskV1 struct {
	// Storage for key:value to process
	keyval ansible.TAnyMap
}

// Here the fields comes as complete values never as jinja2 templates
func (t *TaskV1) SetData(data *ansible.OrderedMap) error {
	d, ok := data.Pop("set_fact")
	if !ok {
		return fmt.Errorf("Unable to find the 'set_fact' map in task data")
	}
	fmap, ok := d.(ansible.OrderedMap)
	if !ok {
		return fmt.Errorf("The 'set_fact' is not the OrderedMap")
	}

	if fmap.Size() < 1 {
		return fmt.Errorf("The 'set_fact' data is empty")
	}

	// Store data for further processing during Run
	t.keyval = make(ansible.TAnyMap, fmap.Size())
	for key, val := range fmap.Data() {
		k := ansible.TString{}
		k.SetUnknown(key)
		v := ansible.TAny{}
		v.SetUnknown(val)
		t.keyval[k] = v
	}

	return nil
}

func (t *TaskV1) GetData() (data ansible.OrderedMap) {
	data.Set("set_fact", t.keyval)
	return data
}

func (t *TaskV1) Run(vars map[string]any) (out ansible.OrderedMap, err error) {
	log.Error("TODO: Implement set_fact.Run")

	return
}
