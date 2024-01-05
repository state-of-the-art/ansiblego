package core

import (
	"math/rand"
	"time"
)

type AnsibleGo struct {
	cfg *CommonConfig

	running bool
}

func New(cfg *CommonConfig) (*AnsibleGo, error) {
	// Init rand generator
	rand.Seed(time.Now().UnixNano())

	ag := &AnsibleGo{cfg: cfg}
	/*if err := ag.Init(); err != nil {
		return nil, err
	}*/

	return ag, nil
}

func (ag *AnsibleGo) Agent(task_data string) error {
	ag.running = true

	/*task := make(map[string]any)
	task["name"] = "Execute nothing"
	task["command"] = "echo ok"
	vars := make(map[string]any)
	vars["test_variable"] = "test data"
	if err := modules.TaskV1Run("command", &task, &vars); err != nil {
		return log.Error("Error during task execution:", err)
	}*/

	return nil
}

func (ag *AnsibleGo) Close() {
	ag.running = false
}
