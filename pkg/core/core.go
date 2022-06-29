package core

import (
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/state-of-the-art/ansiblego/pkg/ansible"
	"github.com/state-of-the-art/ansiblego/pkg/embedbin"
	"github.com/state-of-the-art/ansiblego/pkg/modules"
	"github.com/state-of-the-art/ansiblego/pkg/transport/ssh"
	"github.com/state-of-the-art/ansiblego/pkg/transport/winrm"
)

type AnsibleGo struct {
	cfg *Config

	running bool
}

func New(cfg *Config) (*AnsibleGo, error) {
	// Init rand generator
	rand.Seed(time.Now().UnixNano())

	ag := &AnsibleGo{cfg: cfg}
	if err := ag.Init(); err != nil {
		return nil, err
	}

	return ag, nil
}

func (ag *AnsibleGo) Init() error {
	ag.running = true

	modules.InitEmbedded()

	// Playbook test
	if ag.cfg.Mode == "playbook" {
		log.Println("Loading playbook:", ag.cfg.Args[0])
		pf := ansible.PlaybookFile{}
		err := pf.Load(ag.cfg.Args[0])
		if err != nil {
			return err
		}
		yaml, err := pf.Yaml()
		if err != nil {
			return err
		}
		log.Println("\n" + yaml)
	}

	// SSH test
	ssh_client, err := ssh.New("user", "user", "192.168.56.102", 22)
	if err != nil {
		log.Println("Unable to connect to SSH:", err)
	}
	ssh_kern, ssh_arch, err := ssh_client.Check()
	if err != nil {
		log.Println("Failed to execute SSH check:", err)
	}
	log.Println("SSH Remote system is:", ssh_kern, ssh_arch)
	if err := ssh_client.Execute("ip a", os.Stdout, os.Stderr); err != nil {
		log.Println("Failed to execute command over SSH:", err)
	}

	// SSH Embed binary test
	ssh_embed_fd, err := embedbin.GetEmbeddedBinary(ssh_kern, ssh_arch)
	if err != nil {
		log.Println("Unable to find required binary for remote system:", err)
	}
	defer ssh_embed_fd.Close()

	// Copy embed file with WinRM
	if err := ssh_client.Copy(ssh_embed_fd, "/tmp/ansiblego", 0750); err != nil {
		log.Println("Failed to copy over SSH:", err)
	}

	// WinRM test
	winrm_client, err := winrm.New("user", "user", "192.168.56.101", 5986)
	if err != nil {
		log.Println("Unable to connect to WinRM:", err)
	}
	winrm_kern, winrm_arch, err := winrm_client.Check()
	if err != nil {
		log.Println("Failed to execute WinRM check:", err)
	}
	log.Println("WinRM Remote system is:", winrm_kern, winrm_arch)

	if err := winrm_client.Execute("ipconfig /all", os.Stdout, os.Stderr); err != nil {
		log.Println("Failed to execute command over WinRM:", err)
	}

	// WinRM Embed binary test
	embed_fd, err := embedbin.GetEmbeddedBinary(winrm_kern, winrm_arch)
	if err != nil {
		log.Println("Unable to find required binary for remote system:", err)
	}
	defer embed_fd.Close()

	// Copy embed file with WinRM
	if err := winrm_client.Copy(embed_fd, "C:\\ansiblego.exe", 0750); err != nil {
		log.Println("Failed to execute command over WinRM:", err)
	}

	/*task := make(map[string]any)
	task["name"] = "Execute nothing"
	task["command"] = "echo ok"
	vars := make(map[string]any)
	vars["test_variable"] = "test data"
	if err := modules.TaskV1Run("command", &task, &vars); err != nil {
		log.Println("Error during task execution:", err)
	}*/

	return nil
}

func (ag *AnsibleGo) Close() {
	ag.running = false
}
