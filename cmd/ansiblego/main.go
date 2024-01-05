package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/state-of-the-art/ansiblego/pkg/ansible"
	"github.com/state-of-the-art/ansiblego/pkg/core"
	"github.com/state-of-the-art/ansiblego/pkg/log"
	"github.com/state-of-the-art/ansiblego/pkg/util"
)

var (
	p_cfg_path      string
	p_log_verbosity string
	p_log_timestamp bool
	p_detach        bool
	p_extra_vars    *[]string

	p_task_name *string
	p_task_args *[]string

	root_cmd = &cobra.Command{
		Use:     "ansiblego [OPTIONS] [TASK] [ARGS] [ARGS...]",
		Version: "0.1",
		Short:   "Executes the provided task/s",
		Long:    "It executes the provided task locally, used remotely to actually run tasks and useful to test tasks locally",

		SilenceUsage: true,

		// Init the global variables
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			log.UseTimestamp = p_log_timestamp
			err := log.SetVerbosity(p_log_verbosity)
			if err != nil {
				return err
			}

			return log.InitLoggers(os.Stderr)
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info("AnsibleGo Agent running...")

			cfg := &core.AgentConfig{}
			if err := core.ReadConfigFile(cfg, p_cfg_path); err != nil {
				return log.Error("Unable to apply config file:", p_cfg_path, err)
			}
			cfg.Verbosity = log.Verbosity

			// This decoder is used to read stdin stream with
			// multiple yaml/json documents separated by "---\n"
			yaml_decoder := yaml.NewDecoder(os.Stdin)

			// Processing vars first
			vars := make(map[string]any)
			if p_extra_vars != nil {
				// Parsing extra variables and store to cfg
				for _, keyval := range *p_extra_vars {
					extra_vars_data, err := util.ParseArgument("extra var", keyval, yaml_decoder)
					if err != nil {
						return log.Errorf("Unable to process provided extra var %q: %v", keyval, err)
					}
					for key, val := range extra_vars_data {
						vars[key] = val
						log.Tracef("Provided extra var from stdin: %q: %v", key, val)
					}
				}
			}

			if p_task_name == nil || *p_task_name == "" {
				// Use first argument as task name
				if len(args) < 1 {
					return log.Error("No task name provided")
				}

				p_task_name = &args[0]
			}

			if *p_task_name == "-" {
				// Switch to yaml/json streaming mode - it will work until the
				// received tasks are completed and the stdin stream is closed
				log.Debug("AnsibleGo agent initialized")

				// Loading modules to use them for running the tasks
				ansible.InitEmbeddedModules()
				// TODO: Load external modules from user as well
				// TODO: Load streamed in modules

				log.Debug("Reading tasks from stdin yaml/json stream and execute them...")
				for {
					var task_data map[string]ansible.OrderedMap
					if err := yaml_decoder.Decode(&task_data); err != nil {
						if err == io.EOF {
							break
						}
						return log.Errorf("Unable to parse yaml/json from stdin stream: %v", err)
					}
					for name, data := range task_data {
						if err := runTask(name, data, vars); err != nil {
							return log.Errorf("Executing of task %q ended with error: %v", name, err)
						}
					}
				}
			} else {
				// From here agent runs one task and ending execution with return output of this task

				task_args := ansible.OrderedMap{}
				if p_task_args == nil {
					// If no option provided - using agent command args
					if len(args) > 1 {
						*p_task_args = args[1:]
					}
				}
				if p_task_args != nil {
					// Parsing task args
					for _, keyval := range *p_task_args {
						task_args_data, err := util.ParseArgument("task arg", keyval, yaml_decoder)
						if err != nil {
							return log.Errorf("Unable to process provided task arg %q: %v", keyval, err)
						}
						for key, val := range task_args_data {
							task_args.Set(key, val)
							log.Tracef("Provided task arg: %q=%v", key, val)
						}
					}
				}

				log.Debug("AnsibleGo agent initialized")

				// Loading modules to use them for running the task
				ansible.InitEmbeddedModules()
				// TODO: Load external modules from user as well
				// TODO: Load streamed in modules

				if err := runTask(*p_task_name, task_args, vars); err != nil {
					return log.Errorf("Executing of task %q ended with error: %v", *p_task_name, err)
				}
			}

			log.Info("AnsibleGo exiting...")

			return nil
		},
	}
)

func main() {
	cobra.EnableCommandSorting = false

	// Check if the first argument is a registered command with prefix "--"
	// This is needed for backward compatibility with ansible cmd interfaces
	if len(os.Args) > 1 {
		for _, cmd := range root_cmd.Commands() {
			tmp := []string{"--" + cmd.Name()}
			cmd.Aliases = append(tmp, cmd.Aliases...)
			if os.Args[1][:2] == "--" && os.Args[1][2:] == cmd.Name() {
				os.Args[1] = cmd.Name()
				break
			}
		}
	}

	/*if version {
		log.Println("ansiblego-playbook 2.9.27")
		log.Println("  config file = /home/user/Work/adobe/aquarium-bait/ansible.cfg")
		log.Println("  configured module search path = ['/home/user/.ansible/plugins/modules', '/usr/share/ansible/plugins/modules']")
		log.Println("  ansible python module location = /home/user/Work/adobe/aquarium-bait/.venv/lib/python3.10/site-packages/ansible")
		log.Println("  executable location = /home/user/Work/adobe/aquarium-bait/.venv/bin/ansible-playbook")
		log.Println("  python version = 3.10.12 (main, Jun 11 2023, 05:26:28) [GCC 11.4.0]")
	}*/
	root_cmd.SetVersionTemplate(`{{if .HasParent}}{{.Parent.Name}}-{{end}}{{.Name}} {{.Version}}
	  config file = 
	  configured module search path = 
	  executable location = 
	`)

	// Global flags
	flags := root_cmd.PersistentFlags()
	flags.SortFlags = false

	flags.StringVarP(&p_cfg_path, "cfg", "c", "", "yaml configuration file")
	flags.StringVarP(&p_log_verbosity, "verbosity", "v", "info", "log level (error,warn,info,debug,trace). WARN: trace could expose credentials")
	flags.BoolVar(&p_log_timestamp, "timestamp", true, "prepend timestamps for each log line")
	flags.BoolVar(&p_detach, "detach", false, "detach from shell to background")

	p_extra_vars = flags.StringArrayP("extra-vars", "e", nil, "set additional variables as key=value, - to read YAML from stdin, inline JSON or if filename prepend with @ (can occur multiple times)")

	// Local flags
	root_cmd.Flags().SortFlags = false
	p_task_name = root_cmd.Flags().StringP("module-name", "m", "", "task name to execute or '-' to read yaml/json stream from stdin. Will use first argument if option not provided")
	p_task_args = root_cmd.Flags().StringArrayP("args", "a", nil, "task arguments in key=value or YAML/JSON, if filename prepend with @ (can occur multiple times). Will use consequent arguments if no option is provided")

	if err := root_cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runTask(name string, args ansible.OrderedMap, vars map[string]any) error {
	log.Debugf("Loading task %q", name)
	module_data, err := ansible.GetTaskV1(name)
	if err != nil {
		return log.Errorf("Unable to find %q task: %v", name, err)
	}
	if err = module_data.SetData(args); err != nil {
		return log.Errorf("Unable to set data for task module `%s`: %s", name, err)
	}

	log.Debugf("Running task %q", name)
	out, err := module_data.Run(vars)
	if err != nil {
		return log.Errorf("Error during running task %q: %v", name, err)
	}

	// Print task data output to stdin as yaml
	y, err := out.Yaml()
	if err != nil {
		return log.Errorf("Unable to encode task output to YAML: %v", err)
	}
	fmt.Println(y)

	return nil
}

// Rerun the same application in background
func rerunDetached() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	var args []string
	for _, item := range os.Args {
		if item != "--detach" {
			args = append(args, item)
		}
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = cwd
	err = cmd.Start()
	if err != nil {
		return err
	}
	cmd.Process.Release()

	return nil
}
