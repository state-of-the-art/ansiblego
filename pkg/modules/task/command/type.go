package command

// Doc: https://docs.ansible.com/ansible/2.9/modules/command_module.html

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/state-of-the-art/ansiblego/pkg/modules"
)

type TaskV1 struct {
	Cmd   string   // Executable to run.
	Argv  []string // Arguments of the executable.
	Chdir string   // Change into this directory before running the command.
	// Set the stdin of the command directly to the specified value.
	Stdin string

	// A filename or glob pattern. If it already exists, this step won't be run.
	//Creates string
	// A filename or glob pattern. If it already exists, this step will be run.
	//Removes string

	// If set to true, append a newline to stdin data.
	Stdin_add_newline bool `,def:true`
	// Strip empty lines from the end of stdout/stderr in result.
	Strip_empty_ends bool `,def:true`

	// Enable or disable task warnings.
	//Warn bool `,def:true`

	// Internal vars
	command_is_string bool // In case the command originally is string
}

func (t *TaskV1) SetData(data modules.OrderedMap) error {
	command_data, ok := data.Get("command")
	if !ok {
		return fmt.Errorf("Unable to find the 'command' string or map in task data")
	}
	fmap, ok := command_data.(modules.OrderedMap)
	if !ok {
		// "command" not a map
		var cmdline string
		cmdline, ok = command_data.(string)
		if ok {
			// "command" is a string
			t.command_is_string = true
			cmdsplit := strings.Split(cmdline, " ")
			t.Cmd = cmdsplit[0]
			if len(cmdsplit) > 1 {
				t.Argv = cmdsplit[1:]
			}
			if args_data, ok := data.Get("args"); ok {
				fmap, _ = args_data.(modules.OrderedMap)
			}
		}
	}
	if !ok {
		return fmt.Errorf("The 'command' is not string or OrderedMap")
	}
	return modules.TaskV1SetData(t, fmap)
}

func (t *TaskV1) GetData() (data modules.OrderedMap) {
	fmap := modules.TaskV1GetData(t)
	var out modules.OrderedMap
	if t.command_is_string {
		out.Set("command", strings.Join([]string{t.Cmd, strings.Join(t.Argv, " ")}, " "))
		// Filter out the cmd and vars from the fmap
		fmap.Pop("argv")
		fmap.Pop("cmd")
		if fmap.Size() > 0 {
			out.Set("args", fmap)
		}
	} else {
		out.Set("command", fmap)
	}
	return out
}

func runAndLog(cmd *exec.Cmd) (string, string, error) {
	var stdout, stderr bytes.Buffer

	log.Printf("Executing: %s %s", cmd.Path, strings.Join(cmd.Args[1:], " "))
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	stdout_string := strings.TrimSpace(stdout.String())
	stderr_string := strings.TrimSpace(stderr.String())

	if _, ok := err.(*exec.ExitError); ok {
		message := stderr_string
		if message == "" {
			message = stdout_string
		}

		err = fmt.Errorf("Exit error: %s", message)
	}

	if len(stdout_string) > 0 {
		log.Printf("Stdout: %s", stdout_string)
	}
	if len(stderr_string) > 0 {
		log.Printf("Stderr: %s", stderr_string)
	}

	// Replace these for Windows, we only want to deal with Unix style line endings.
	return_stdout := strings.Replace(stdout.String(), "\r\n", "\n", -1)
	return_stderr := strings.Replace(stderr.String(), "\r\n", "\n", -1)

	return return_stdout, return_stderr, err
}

func (t *TaskV1) Run(vars map[string]any) error {
	var cmd *exec.Cmd
	if t.Cmd != "" {
		cmd = exec.Command(t.Cmd, t.Argv...)
	} else {
		// If it's just argv then use it's first item as cmd
		cmd = exec.Command(t.Argv[0], t.Argv[1:]...)
	}
	runAndLog(cmd)

	return nil
}
