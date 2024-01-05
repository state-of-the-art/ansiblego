package transport

import (
	"io"
	"os"
)

// Transports are used internally to communicate with remote system and
// copy the AnsibleGo binary into the system to use it for execution of
// the tasks. Transports should not be used to directly execute commands
// of tasks anyhow.
type Transport interface {
	Execute(cmd string, stdout, stderr io.Writer) (err error)
	ExecuteInput(cmd string, stdin io.Reader, stdout, stderr io.Writer) (err error)

	Check() (kernel, arch string, err error)

	Copy(content io.Reader, dst string, mode os.FileMode) (err error)
}
