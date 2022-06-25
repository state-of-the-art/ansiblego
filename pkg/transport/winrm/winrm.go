package winrm

import (
	"fmt"
	"io"
	"os"

	"github.com/masterzen/winrm"
)

type TransportWinRM struct {
	client *winrm.Client
}

func New(user, pass, host string, port int) (*TransportWinRM, error) {
	tr := &TransportWinRM{}

	if err := tr.connectPassword(user, pass, host, port); err != nil {
		return nil, fmt.Errorf("Failed to connect with password: %v", err)
	}

	return tr, nil
}

func (tr *TransportWinRM) connectPassword(user, pass, host string, port int) (err error) {
	config := winrm.NewEndpoint(
		host,  // Host to connect to
		port,  // WinRM port
		false, // Use TLS
		false, // Allow insecure connection
		nil,   // CA certificate
		nil,   // Client Certificate
		nil,   // Client Key
		0,     // Timeout
	)

	tr.client, err = winrm.NewClient(config, user, pass)
	if err != nil {
		return fmt.Errorf("Failed to create WinRM client: %v", err)
	}

	return nil
}

func (tr *TransportWinRM) Execute(cmd string, stdout, stderr io.Writer) (err error) {
	_, err = tr.client.Run(cmd, stdout, stderr)
	if err != nil {
		return fmt.Errorf("Failed to run command: %v", err)
	}
	return nil
}

// A bit adjusted function from https://github.com/jbrekelmans/go-winrm/blob/master/copier.go
func (tr *TransportWinRM) Copy(content io.Reader, dst string, mode os.FileMode) error {
	// Uses the usual chunk-based approach, while could be optimized
	// TODO: through PSRP or https://github.com/jbrekelmans/go-winrm
	return doCopy(tr.client, 10, 512, content, dst)
}
