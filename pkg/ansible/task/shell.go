package task

// Doc: https://docs.ansible.com/ansible/2.9/modules/shell_module.html

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/state-of-the-art/ansiblego/pkg/ansible"
	"github.com/state-of-the-art/ansiblego/pkg/log"
)

type TaskV1 struct {
	// Executable to run.
	Cmd string
	// Change into this directory before running the command.
	Chdir string
	// Set the stdin of the command directly to the specified value.
	Stdin string

	// Change the shell used to execute the command.
	//Executable string
	// The shell module takes a free form command to run, as a string.
	//Free_form string
	// A filename or glob pattern. If it already exists, this step won't be run.
	//Creates string
	// A filename or glob pattern. If it already exists, this step will be run.
	//Removes string

	// If set to true, append a newline to stdin data.
	Stdin_add_newline bool `task:",def:true"`

	// Enable or disable task warnings.
	//Warn bool `task:",def:true"`

	// Internal vars
	shell_is_string bool // In case the shell originally is string
}

func (t *TaskV1) SetData(data ansible.OrderedMap) error {
	shell_data, ok := data.Get("shell")
	if !ok {
		return fmt.Errorf("Unable to find the 'shell' string or map in task data")
	}
	fmap, ok := shell_data.(ansible.OrderedMap)
	if !ok {
		// "shell" not a map
		var cmdline string
		cmdline, ok = shell_data.(string)
		if ok {
			// "shell" is a string
			t.shell_is_string = true
			t.Cmd = cmdline
			// Args are confusing and instead module need to be used, so skip processing
			/*if args_data, ok := data.Get("args"); ok {
				fmap, _ = args_data.(ansible.OrderedMap)
			}*/
		}
	}
	if !ok {
		return fmt.Errorf("The 'shell' is not string or OrderedMap")
	}
	return ansible.TaskV1SetData(t, fmap)
}

func (t *TaskV1) GetData() (data ansible.OrderedMap) {
	fmap := ansible.TaskV1GetData(t)
	if t.shell_is_string {
		data.Set("shell", t.Cmd)
		// Args are confusing and instead module need to be used, so skip processing
		/*// Filter out the cmd and vars from the fmap
		fmap.Pop("argv")
		fmap.Pop("cmd")
		if fmap.Size() > 0 {
			data.Set("args", fmap)
		}*/
	} else {
		data.Set("shell", fmap)
	}
	return data
}

func runAndLog(cmd *exec.Cmd) (string, string, error) {
	var stdout, stderr bytes.Buffer

	log.Debugf("Executing: %s %s", cmd.Path, strings.Join(cmd.Args[1:], " "))
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
		log.Debugf("Stdout: %s", stdout_string)
	}
	if len(stderr_string) > 0 {
		log.Debugf("Stderr: %s", stderr_string)
	}

	// Replace these for Windows, we only want to deal with Unix style line endings.
	return_stdout := strings.Replace(stdout.String(), "\r\n", "\n", -1)
	return_stderr := strings.Replace(stderr.String(), "\r\n", "\n", -1)

	return return_stdout, return_stderr, err
}

func (t *TaskV1) Run(vars map[string]any) (out ansible.OrderedMap, err error) {
	log.Error("TODO: Implement shell.Run")
	cmd := exec.Command("/usr/bin/sh", "-c", fmt.Sprintf("%v", vars["shell"]))
	runAndLog(cmd)

	return
}

func main() {
	// TODO: shell line interface
}
