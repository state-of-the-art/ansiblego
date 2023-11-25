package task

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/state-of-the-art/ansiblego/pkg/log"
)

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

func TaskV1(data, vars *map[string]any) {
	cmd := exec.Command("/usr/bin/sh", "-c", fmt.Sprintf("%v", (*data)["shell"]))
	runAndLog(cmd)
}

func main() {
	// TODO: command line interface
}
