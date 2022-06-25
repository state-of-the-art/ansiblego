package ssh

import (
	"context"
	"fmt"
	"io"
	"os"

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
