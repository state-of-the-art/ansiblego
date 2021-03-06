package ansible

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"

	m "github.com/state-of-the-art/ansiblego/pkg/modules"
)

type Task struct {
	Name        string        `yaml:",omitempty"`
	Environment *m.OrderedMap `yaml:",omitempty"`

	Become     bool          `yaml:",omitempty"`
	Vars       *m.OrderedMap `yaml:",omitempty"`
	With_items []string      `yaml:",omitempty"`
	With_dict  *m.OrderedMap `yaml:",omitempty"`

	ModuleName string            `yaml:"-"`
	ModuleData m.TaskV1Interface `yaml:"-"`
}

type tmpTask Task // Used for quick yml unmarshal

func (c *Task) Load(yml_path string) error {
	// Open and parse
	data, err := ioutil.ReadFile(yml_path)
	if err != nil {
		return err
	}

	return c.Parse(data)
}

func (c *Task) Parse(data []byte) error {
	if err := yaml.Unmarshal(data, c); err != nil {
		return err
	}

	return nil
}

func (c *Task) Yaml() (string, error) {
	buf := bytes.Buffer{}
	enc := yaml.NewEncoder(&buf)
	defer enc.Close()
	enc.SetIndent(2)
	if err := enc.Encode(c); err != nil {
		return "", fmt.Errorf("YAML encode error: %v", err)
	}
	return "---\n" + buf.String(), nil
}

// Determine what kind of the additional fields is here
// The function checks every key in yaml map and compare it to the available in structure. The
// unknown fields are checking on available task and if task found - it's good, if not - then
// keep search until it found.
func (c *Task) UnmarshalYAML(value *yaml.Node) (err error) {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("Task must be YAML Map, but is %v", value.Kind)
	}

	// Simple/dirty way to get the values because dynamic
	// with reflection looks ugly and not working as expected
	var tmp_task tmpTask
	if err := value.Decode(&tmp_task); err != nil {
		return err
	}
	c.Name = tmp_task.Name
	c.Environment = tmp_task.Environment
	c.Become = tmp_task.Become
	c.Vars = tmp_task.Vars
	c.With_items = tmp_task.With_items
	c.With_dict = tmp_task.With_dict

	// Collecting the structure fields to fill
	struct_v := reflect.ValueOf(c)
	struct_t := struct_v.Elem().Type()
	struct_fields := make(map[string]int, struct_t.NumField())

	for i := 0; i < struct_t.NumField(); i++ {
		f := struct_t.Field(i)

		// For now just check for tag yaml:"-" to skip it
		if f.Tag.Get("yaml") == "-" {
			continue
		}

		name := strings.ToLower(f.Name)
		struct_fields[name] = i
	}

	var tmp_fields map[string]yaml.Node
	if err := value.Decode(&tmp_fields); err != nil {
		return err
	}

	var task_fields m.OrderedMap
	// Searching the unknown fields in yaml map
	for k, node := range tmp_fields {
		if _, ok := struct_fields[k]; !ok {
			switch node.Kind {
			case yaml.MappingNode:
				var val m.OrderedMap
				if err := node.Decode(&val); err != nil {
					return err
				}
				task_fields.Set(k, val)
			default:
				var val any
				if err := node.Decode(&val); err != nil {
					return err
				}
				task_fields.Set(k, val)
			}
			if m.IsTask(k) && len(c.ModuleName) < 1 {
				c.ModuleName = k
			}
		}
	}
	if len(c.ModuleName) < 1 {
		y, _ := yaml.Marshal(value)
		return fmt.Errorf("Task module for task `%s` is not implemented:\n", c.Name, string(y))
	}

	// Filling the task with data
	c.ModuleData, err = m.GetTaskV1(c.ModuleName)
	if err != nil {
		return fmt.Errorf("Unable to get definition variable from task module `%s`: %s", c.ModuleName, err)
	}
	err = c.ModuleData.SetData(task_fields)
	if err != nil {
		return fmt.Errorf("Unable to set data for task module `%s`: %s", c.ModuleName, err)
	}

	return nil
}

func (c *Task) MarshalYAML() (interface{}, error) {
	// General task data
	node := &yaml.Node{}
	if err := node.Encode(*c); err != nil {
		return nil, err
	}

	// Adding data from module
	data := c.ModuleData.GetData()
	module_node := &yaml.Node{}
	if err := module_node.Encode(&data); err != nil {
		return nil, err
	}
	node.Content = append(node.Content, module_node.Content...)

	return node, nil
}
