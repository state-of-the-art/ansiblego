package sshd

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/mattn/go-shellwords"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/state-of-the-art/ansiblego/pkg/log"
)

// Needs just user, password and addr like "0.0.0.0:2222"
func Run(user, password, addr string) error {
	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == user && string(pass) == password {
				return nil, nil
			}
			return nil, fmt.Errorf("Password rejected for %q", c.User())
		},
	}

	private_key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return log.Error("Failed to generate new private key: ", err)
	}

	private, err := ssh.NewSignerFromKey(private_key)
	if err != nil {
		return log.Error("Failed to create signer for private key: ", err)
	}

	config.AddHostKey(private)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return log.Error("Failed to listen for connection: ", err)
	}
	for {
		nconn, err := listener.Accept()
		if err != nil {
			log.Error("Failed to accept incoming connection: ", err)
		}

		conn, chans, reqs, err := ssh.NewServerConn(nconn, config)
		if err != nil {
			log.Error("Failed to handshake: ", err)
			nconn.Close()
			continue
		}
		log.Infof("User '%s' just logged in\n", conn.User())

		go ssh.DiscardRequests(reqs)
		go handleChannels(chans)
	}
}

func handleChannels(chans <-chan ssh.NewChannel) {
	for new_channel := range chans {
		go handleChannel(new_channel)
	}
}

type exitStatusMsg struct {
	Status uint32
}

func handleChannel(new_channel ssh.NewChannel) {
	if new_channel.ChannelType() != "session" {
		new_channel.Reject(ssh.UnknownChannelType, "unknown channel type")
		return
	}
	channel, requests, err := new_channel.Accept()
	if err != nil {
		log.Errorf("Could not accept channel: %v", err)
	}

	env := os.Environ()

	go func(in <-chan *ssh.Request) {
		for req := range in {
			switch req.Type {
			case "pty-req":
				req.Reply(true, nil)
				log.Debug("Pty request")
			case "env":
				e := struct{ Name, Value string }{}
				ssh.Unmarshal(req.Payload, &e)
				kv := e.Name + "=" + e.Value
				env = appendEnv(env, kv)
				req.Reply(true, nil)
			case "exec":
				cmd := string(req.Payload)
				err := processCmd(channel, cmd[4:], env)
				ex := exitStatusMsg{
					Status: 0,
				}
				if err != nil {
					ex.Status = 1
					log.Error("Executing error:", err)
				}
				if _, err := channel.SendRequest("exit-status", false, ssh.Marshal(&ex)); err != nil {
					log.Errorf("Unable to send status: %v", err)
				}
				channel.Close()
			case "shell":
				//cmd := string(req.Payload)
				// Running shell terminal
				go func() {
					defer channel.Close()
					for {
						term := terminal.NewTerminal(channel, "> ")
						line, err := term.ReadLine()
						if err != nil {
							break
						}
						err = processCmd(term, line, env)
						if err != nil {
							log.Error("Executing error:", err)
						}
					}
				}()
			default:
				log.Error("Unknown type:", req.Type)
				req.Reply(true, nil)
			}
		}
	}(requests)
}

func processCmd(connection io.Writer, command string, envs []string) error {
	// Special case for uname to show proper system on windows
	if command == "uname -s -m" {
		connection.Write([]byte(fmt.Sprintf("%s %s\n", runtime.GOOS, runtime.GOARCH)))
		return nil
	}
	// Parse the shell env and args
	parser := shellwords.NewParser()
	parser.ParseEnv = true
	env, args, err := parser.ParseWithEnvs(command)
	if err != nil {
		return err
	}
	envs = append(envs, env...)

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = envs

	log.Infof("Executing: %s %s", cmd.Path, strings.Join(cmd.Args[1:], " "))
	cmd.Stdout = connection
	cmd.Stderr = connection
	err = cmd.Run()

	return err
}

func appendEnv(env []string, kv string) []string {
	p := strings.SplitN(kv, "=", 2)
	k := p[0] + "="
	for i, e := range env {
		if strings.HasPrefix(e, k) {
			env[i] = kv
			return env
		}
	}
	return append(env, kv)
}
