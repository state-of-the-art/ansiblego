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

	// Embed binary test
	embed_bin_data, err := embedbin.GetEmbeddedBinary("linux", "amd64")
	if err != nil {
		log.Println("Unable to find required binary for remote system:", err)
	}
	log.Println("Found binary size:", len(embed_bin_data))

	// SSH test
	ssh_client, err := ssh.New("user", "user", "127.0.0.1", 22)
	if err != nil {
		log.Println("Unable to connect to SSH:", err)
	}
	if err := ssh_client.Execute("ip a", os.Stdout, os.Stderr); err != nil {
		log.Println("Failed to execute command over SSH:", err)
	}

	// WinRM test
	winrm_client, err := winrm.New("user", "user", "192.168.56.101", 5985)
	if err != nil {
		log.Println("Unable to connect to WinRM:", err)
	}

	if err := winrm_client.Execute("ipconfig /all", os.Stdout, os.Stderr); err != nil {
		log.Println("Failed to execute command over WinRM:", err)
	}

	exec_path, _ := os.Executable()
	fd, _ := os.Open(exec_path)
	if err := winrm_client.Copy(fd, "C:\\test\\ansiblego.exe", 0750); err != nil {
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
