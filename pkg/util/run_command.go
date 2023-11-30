package util

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/state-of-the-art/ansiblego/pkg/log"
)

// Runs & logs the executable command
func RunCommand(timeout time.Duration, path string, arg ...string) (string, string, error) {
	var stdout, stderr bytes.Buffer

	// Running command with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, arg...)

	log.Debug("Executing:", cmd.Path, strings.Join(cmd.Args[1:], " "))
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	stdoutString := strings.TrimSpace(stdout.String())
	stderrString := strings.TrimSpace(stderr.String())

	// Check the context error to see if the timeout was executed
	if ctx.Err() == context.DeadlineExceeded {
		err = fmt.Errorf("Command timed out")
	} else if _, ok := err.(*exec.ExitError); ok {
		message := stderrString
		if message == "" {
			message = stdoutString
		}

		err = fmt.Errorf("Command exited with error: %v: %s", err, message)
	}

	if len(stdoutString) > 0 {
		log.Debug("stdout:", stdoutString)
	}
	if len(stderrString) > 0 {
		log.Debug("stderr:", stderrString)
	}

	// Replace these for Windows, we only want to deal with Unix style line endings.
	returnStdout := strings.Replace(stdout.String(), "\r\n", "\n", -1)
	returnStderr := strings.Replace(stderr.String(), "\r\n", "\n", -1)

	return returnStdout, returnStderr, err
}

// Will retry on error and store the retry output and errors to return
func RunCommandRetry(retry int, timeout time.Duration, path string, arg ...string) (stdout string, stderr string, err error) {
	counter := 0
	for {
		counter++
		rout, rerr, err := RunCommand(timeout, path, arg...)
		if err != nil {
			stdout += fmt.Sprintf("\n--- Command execution attempt %d ---\n", counter)
			stdout += rout
			stderr += fmt.Sprintf("\n--- Command execution attempt %d ---\n", counter)
			stderr += rerr
			if counter <= retry {
				// Give command 5 seconds to rest
				time.Sleep(5 * time.Second)
				continue
			}
		}
		return stdout, stderr, err
	}
}
