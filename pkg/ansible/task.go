package ansible

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/state-of-the-art/ansiblego/pkg/embedbin"
	"github.com/state-of-the-art/ansiblego/pkg/log"
	"github.com/state-of-the-art/ansiblego/pkg/transport"
	"github.com/state-of-the-art/ansiblego/pkg/transport/ssh"
	"github.com/state-of-the-art/ansiblego/pkg/transport/winrm"
)

type Task struct {
	// Identifier. Can be used for documentation, in or tasks/handlers.
	Name string `yaml:",omitempty"`
	// A dictionary that gets converted into environment vars to be provided for the task upon execution. This cannot affect Ansible itself nor its configuration, it just sets the variables for the code responsible for executing the task.
	Environment *OrderedMap `yaml:",omitempty"`

	// Conditional expression, determines if an iteration of a task is run or not.
	When string `yaml:",omitempty"` // Only string right now, array looks confusing
	// Boolean that controls if privilege escalation is used or not on Task execution.
	Become bool `yaml:",omitempty"`
	// Dictionary/map of variables specified in task
	Vars *OrderedMap `yaml:",omitempty"`
	// Name of variable that will contain task status and module return data.
	Register string `yaml:",omitempty"`

	// Host to execute task instead of the target (inventory_hostname). Connection vars from the delegated host will also be used for the task.
	Delegate_to string `yaml:",omitempty"`
	// Conditional expression that overrides the task’s normal ‘failed’ status.
	Failed_when string `yaml:",omitempty"`

	// Loop through list of items
	With_items []string `yaml:",omitempty"`
	// Loop through dict key value
	With_dict *OrderedMap `yaml:",omitempty"`

	// TODO: Actually block could be potentally a task module, but for now in v1 it's just a
	// special case of task. Maybe in v2 it will be possible to pass yaml nodes to the tasks to
	// properly process subtasks of block, who knows...
	Block []*Task `yaml:",omitempty"` // Special case, contains list of tasks to execute

	// Found task module name
	ModuleName string `yaml:"-"`
	// Loaded implementation of the task module
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

func (t *Task) Run(vars map[string]any) (data OrderedMap, err error) {
	if len(t.Block) > 0 {
		if t.Name != "" {
			log.Warnf("Executing task block '%s'", t.Name)
		}
		for _, task := range t.Block {
			if _, err = task.Run(vars); err != nil {
				return
			}
		}
	} else {
		// In case need to be executed remotely - route task to the proper transport
		if t.IsRemote(vars) {
			log.Infof("Executing task '%s' remotely", t.Name)
			var client transport.Transport

			if vars["ansible_connection"] == "ssh" {
				user, ok1 := vars["ansible_ssh_user"].(string)
				host, ok2 := vars["ansible_ssh_host"].(string)
				port_str, ok3 := vars["ansible_ssh_port"].(string)
				if !(ok1 && ok2 && ok3) {
					return data, log.Error("Unable to get the necessary vars to connect via SSH")
				}

				port, err := strconv.Atoi(port_str)
				if err != nil {
					return data, log.Errorf("Unable to use non-int port: %q", port_str)
				}

				log.Debugf("Connecting via SSH to '%s@%s:%d'", user, host, port)

				// Check if password or key provided and create needed ssh connection
				if password, ok4 := vars["ansible_password"].(string); ok4 {
					if client, err = ssh.NewPass(user, password, host, port); err != nil {
						return data, log.Error("Unable to connect to SSH by password:", err)
					}
				} else if password, ok4 := vars["ansible_ssh_private_key_file"].(string); ok4 {
					if client, err = ssh.NewKey(user, password, host, port); err != nil {
						return data, log.Error("Unable to connect to SSH by key:", err)
					}
				} else {
					return data, log.Error("Unable to get password or key to connect via SSH")
				}
			} else if vars["ansible_connection"] == "winrm" {
				user, ok1 := vars["ansible_winrm_user"].(string)
				password, ok2 := vars["ansible_winrm_password"].(string)
				host, ok3 := vars["ansible_winrm_host"].(string)
				port_str, ok4 := vars["ansible_winrm_port"].(string)
				if !(ok1 && ok2 && ok3 && ok4) {
					return data, log.Error("Unable to get the necessary vars to connect via WinRM")
				}

				port, err := strconv.Atoi(port_str)
				if err != nil {
					return data, log.Errorf("Unable to use non-int port: %q", port_str)
				}

				log.Debugf("Connecting via WinRM to '%s@%s:%d'", user, host, port)

				if client, err = winrm.New(user, password, host, port); err != nil {
					return data, log.Error("Unable to connect to WinRM:", err)
				}
				/*kern, arch, err := client.Check()
				if err != nil {
					return data, log.Error("Failed to execute WinRM check:", err)
				}
				log.Info("WinRM Remote system is:", kern, arch)

				if err := client.Execute("ipconfig /all", os.Stdout, os.Stderr); err != nil {
					return data, log.Error("Failed to execute command over WinRM:", err)
				}

				// WinRM Embed binary test
				embed_fd, err := embedbin.GetEmbeddedBinary(kern, arch)
				if err != nil {
					return data, log.Error("Unable to find required binary for remote system:", err)
				}
				defer embed_fd.Close()

				// Copy embed file with WinRM
				if err := client.Copy(embed_fd, "C:\\ansiblego.exe", 0750); err != nil {
					return data, log.Error("Failed to execute command over WinRM:", err)
				}*/
			} else {
				return data, log.Errorf("Unable to find connection plugin for %q", vars["ansible_connection"])
			}

			// Getting the system info
			kern, arch, err := client.Check()
			if err != nil {
				return data, log.Error("Failed to execute remote system check:", err)
			}
			log.Debug("Remote system is:", kern, arch)

			// Getting fitting embed ansiblego exec data
			embed_fd, err := embedbin.GetEmbeddedBinary(kern, arch)
			if err != nil {
				return data, log.Error("Unable to find ansiblego binary for target system:", err)
			}
			defer embed_fd.Close()

			// Copy embed file to the target system
			if err := client.Copy(embed_fd, "/tmp/ansiblego", 0750); err != nil {
				return data, log.Error("Failed to copy ansiblego agent to target system:", err)
			}

			// TODO: Execute the module via ansiblego agent
			// Check ansiblego can be running
			if err := client.Execute("/tmp/ansiblego", os.Stdout, os.Stderr); err != nil {
				return data, log.Error("Failed to execute ansiblego agent:", err)
			}
		} else {
			log.Infof("Executing task '%s' locally", t.Name)
			return t.ModuleData.Run(vars)
		}
	}

	return
}

func (t *Task) IsRemote(vars map[string]any) bool {
	if t.Delegate_to == "localhost" {
		return false
	}
	if t.Delegate_to == "" {
		// Refer to the variables
		if val, ok := vars["ansible_connection"]; ok && val.(string) == "local" || !ok {
			return false
		}
	}

	return true
}
