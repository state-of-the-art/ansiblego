package ssh

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bramvdbogaerde/go-scp"
	"golang.org/x/crypto/ssh"
)

type TransportSSH struct {
	host   string
	port   int
	config *ssh.ClientConfig
}

func New(user, pass, host string, port int) (*TransportSSH, error) {
	tr := &TransportSSH{
		host: host,
		port: port,
		config: &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{ssh.Password(pass)},

			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		},
	}

	return tr, nil
}

func (tr *TransportSSH) connect() (*ssh.Client, *ssh.Session, error) {
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", tr.host, tr.port), tr.config)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to connect: %v", err)
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, fmt.Errorf("Failed to create SSH session: %v", err)
	}

	if _, _, err := tr.Check(); err != nil {
		return nil, nil, err
	}

	return client, session, nil
}

func (tr *TransportSSH) Execute(cmd string, stdout, stderr io.Writer) (err error) {
	client, session, err := tr.connect()
	if err != nil {
		return err
	}
	defer client.Close()
	defer session.Close()

	stdout_pipe, _ := session.StdoutPipe()
	go io.Copy(stdout, stdout_pipe)
	stderr_pipe, _ := session.StderrPipe()
	go io.Copy(stderr, stderr_pipe)
	err = session.Run(cmd)
	if err != nil {
		return fmt.Errorf("Failed to run command: %v", err)
	}
	return nil
}

func (tr *TransportSSH) Check() (kernel, arch string, err error) {
	stderr_buf := bytes.Buffer{}
	// Get remote system kernel
	kern_buf := bytes.Buffer{}
	if err = tr.Execute("uname -s", &kern_buf, &stderr_buf); err != nil {
		return kernel, arch, fmt.Errorf("Unable to get the remote system kernel: %v, %v", err, stderr_buf.String())
	}

	kernel = strings.ToLower(kern_buf.String())

	// Get remote system arch
	arch_buf := bytes.Buffer{}
	if err = tr.Execute("uname -m", &arch_buf, &stderr_buf); err != nil {
		return kernel, arch, fmt.Errorf("Unable to get the remote system arch: %v, %v", err, stderr_buf.String())
	}

	arch = arch_buf.String()
	if arch == "x86_64" || arch == "x64" {
		arch = "amd64"
	}

	return kernel, arch, nil
}

func (tr *TransportSSH) Copy(content io.Reader, dst string, mode os.FileMode) (err error) {
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", tr.host, tr.port), tr.config)
	if err != nil {
		return fmt.Errorf("Failed to connect: %v", err)
	}
	defer client.Close()

	scp_client, err := scp.NewClientBySSH(client)
	if err != nil {
		return fmt.Errorf("Failed to create SCP session: %v", err)
	}
	defer scp_client.Close()

	scp_client.CopyFile(context.Background(), content, dst, fmt.Sprintf("%#o", mode))
	if err != nil {
		return fmt.Errorf("Error while copying file: %v", err)
	}

	return nil
}
