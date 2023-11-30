package ansible

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/state-of-the-art/ansiblego/pkg/log"
)

type Task struct {
	Name        string      `yaml:",omitempty"`
	Environment *OrderedMap `yaml:",omitempty"`

	When     string      `yaml:",omitempty"` // Only string right now, array looks confusing
	Become   bool        `yaml:",omitempty"`
	Vars     *OrderedMap `yaml:",omitempty"`
	Register string      `yaml:",omitempty"`

	With_items []string    `yaml:",omitempty"`
	With_dict  *OrderedMap `yaml:",omitempty"`

	Failed_when string `yaml:",omitempty"`

	// TODO: Actually block could be potentally a task module, but for now in v1 it's just a
	// special case of task. Maybe in v2 it will be possible to pass yaml nodes to the tasks to
	// properly process subtasks of block, who knows...
	Block []*Task `yaml:",omitempty"` // Special case, contains list of tasks to execute

	ModuleName string          `yaml:"-"`
	ModuleData TaskV1Interface `yaml:"-"`
}

type tmpTask Task // Used for quick yml unmarshal

func (t *Task) Load(yml_path string) error {
	// Open and parse
	data, err := ioutil.ReadFile(yml_path)
	if err != nil {
		return err
	}

	return t.Parse(data)
}

func (t *Task) Parse(data []byte) error {
	if err := yaml.Unmarshal(data, t); err != nil {
		return err
	}

	return nil
}

func (t *Task) Yaml() (string, error) {
	return ToYaml(t)
}

// Determine what kind of the additional fields is here
// The function checks every key in yaml map and compare it to the available in structure. The
// unknown fields are checking on available task and if task found - it's good, if not - then
// keep search until it found.
func (t *Task) UnmarshalYAML(value *yaml.Node) (err error) {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("Task must be YAML Map, but is %v", value.Kind)
	}

	// Simple/dirty way to get the values because dynamic
	// with reflection looks ugly and not working as expected
	var tmp_task tmpTask
	if err := value.Decode(&tmp_task); err != nil {
		return err
	}
	t.Name = tmp_task.Name
	t.Environment = tmp_task.Environment
	t.Block = tmp_task.Block
	t.When = tmp_task.When
	t.Become = tmp_task.Become
	t.Vars = tmp_task.Vars
	t.With_items = tmp_task.With_items
	t.With_dict = tmp_task.With_dict
	t.Failed_when = tmp_task.Failed_when
	t.Register = tmp_task.Register

	// Collecting the structure fields to fill
	struct_v := reflect.ValueOf(t)
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

	var task_fields OrderedMap
	// Searching the unknown fields in yaml map
	for k, node := range tmp_fields {
		// Removing prefix for the `win_` field since we have universal ones
		if strings.HasPrefix(k, "win_") {
			log.Warnf("Found win_ prefixed task '%s' - using it without prefix\n", k)
			k = k[4:]
		}
		if _, ok := struct_fields[k]; !ok {
			switch node.Kind {
			case yaml.MappingNode:
				var val OrderedMap
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
			if ModuleIsTask(k) && len(t.ModuleName) < 1 {
				t.ModuleName = k
			}
		}
	}

	// If task is not a block - then processing as task module
	if len(t.Block) < 1 {
		// Processing task module
		if len(t.ModuleName) < 1 {
			y, err := ToYaml(value)
			if err != nil {
				return fmt.Errorf("Task module for task `%s` is not implemented, but unable to show:\n%v", t.Name, err)
			}
			return fmt.Errorf("Task module for task `%s` is not implemented:\n%s", t.Name, y)
		}

		// Filling the task with data
		t.ModuleData, err = GetTaskV1(t.ModuleName)
		if err != nil {
			return fmt.Errorf("Unable to get definition variable from task module `%s`: %s", t.ModuleName, err)
		}
		err = t.ModuleData.SetData(task_fields)
		if err != nil {
			return fmt.Errorf("Unable to set data for task module `%s`: %s", t.ModuleName, err)
		}

		// Remove processed task module, to check later if there is something else left...
		task_fields.Pop(t.ModuleName)
	}

	// In case there are something else - let's tell user about that, because skipping could do more harm
	if task_fields.Size() > 0 {
		y, err := task_fields.Yaml()
		if err != nil {
			return fmt.Errorf("Task `%s` contains unknown fields, but unable to show:\n%v", t.Name, err)
		}
		return fmt.Errorf("Task `%s` contains unknown fields:\n%s", t.Name, y)
	}

	return nil
}

func (t *Task) MarshalYAML() (interface{}, error) {
	// General task data
	node := &yaml.Node{}
	if err := node.Encode(*t); err != nil {
		return nil, err
	}

	// Adding data from module
	var data OrderedMap
	if len(t.Block) < 1 {
		data = t.ModuleData.GetData()
	}
	module_node := &yaml.Node{}
	if err := module_node.Encode(&data); err != nil {
		return nil, err
	}
	node.Content = append(node.Content, module_node.Content...)
	node.Style = 0 // Preventing encoder from switching to FlowStyle which mess the yaml style

	return node, nil
}

func (t *Task) Run(vars map[string]any) error {
	if len(t.Block) > 0 {
		if t.Name != "" {
			log.Warnf("Executing task block '%s'", t.Name)
		}
		for _, task := range t.Block {
			if err := task.Run(vars); err != nil {
				return err
			}
		}
	} else {
		log.Infof("Executing task '%s'", t.Name)
		t.ModuleData.Run(vars)
	}
	return nil
}
