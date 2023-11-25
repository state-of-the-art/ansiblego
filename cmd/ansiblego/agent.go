package main

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/state-of-the-art/ansiblego/pkg/core"
	"github.com/state-of-the-art/ansiblego/pkg/log"
)

var agent_cmd = &cobra.Command{
	Use:   "agent",
	Short: "Executes the provided task/s",
	Long:  "It runs remotely to execute the requests and is usable to test commands",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Info("AnsibleGo Agent running...")

		cfg := &core.AgentConfig{}
		if err := core.ReadConfigFile(cfg, cfg_path); err != nil {
			return log.Error("Unable to apply config file:", cfg_path, err)
		}
		cfg.Verbosity = log.Verbosity

		ango, err := core.New(&cfg.CommonConfig)
		if err != nil {
			return err
		}

		log.Debug("AgentConfig:", cfg)

		log.Debug("AnsibleGo initialized")

		// For now use it as test runner
		/*for _, task_data := range args {
			ango.Agent(task_data)
		}*/

		_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		ango.Close()

		log.Info("AnsibleGo exiting...")

		return nil
	},
}

func init() {
	agent_cmd.Flags().SortFlags = false
	root_cmd.AddCommand(agent_cmd)
}
