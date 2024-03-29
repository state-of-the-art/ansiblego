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

func NewPass(user, pass, host string, port int) (*TransportSSH, error) {
	tr := &TransportSSH{
		host: host,
		port: port,
		config: &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{ssh.Password(pass)},

			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		},
	}

	if _, _, err := tr.Check(); err != nil {
		return nil, err
	}

	return tr, nil
}

func NewKey(user, key_path, host string, port int) (*TransportSSH, error) {
	key, err := os.ReadFile(key_path)
	if err != nil {
		return nil, fmt.Errorf("Unable to read ssh key file: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse private key: %v", err)
	}

	tr := &TransportSSH{
		host: host,
		port: port,
		config: &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},

			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		},
	}

	if _, _, err := tr.Check(); err != nil {
		return nil, fmt.Errorf("Unable to verify ssh connection: %v", err)
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

func (tr *TransportSSH) ExecuteInput(cmd string, stdin io.Reader, stdout, stderr io.Writer) (err error) {
	client, session, err := tr.connect()
	if err != nil {
		return err
	}
	defer client.Close()
	defer session.Close()

	stdin_pipe, _ := session.StdinPipe()
	go io.Copy(stdin_pipe, stdin)
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
	stdout_buf := bytes.Buffer{}
	if err = tr.Execute("uname -s -m", &stdout_buf, &stderr_buf); err != nil {
		return kernel, arch, fmt.Errorf("Unable to get the remote system kernel: %v, %v", err, stderr_buf.String())
	}

	kernel_arch_list := strings.Split(stdout_buf.String(), " ")
	if len(kernel_arch_list) != 2 {
		return kernel, arch, fmt.Errorf("Bad output from uname: %v, %v", stdout_buf.String(), stderr_buf.String())
	}

	kernel = strings.TrimSpace(strings.ToLower(kernel_arch_list[0]))
	arch = strings.TrimSpace(strings.ToLower(kernel_arch_list[1]))

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
