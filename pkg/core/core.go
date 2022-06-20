package core

import (
	"log"
	"math/rand"
	"time"

	"github.com/state-of-the-art/ansiblego/pkg/ansible"
	"github.com/state-of-the-art/ansiblego/pkg/embedbin"
	"github.com/state-of-the-art/ansiblego/pkg/modules"
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

	embed_bin_data, err := embedbin.GetEmbeddedBinary("darwin", "amd64")
	if err != nil {
		log.Println("Unable to find required binary for remote system:", err)
	}
	log.Println("Found binary size:", len(embed_bin_data))

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
