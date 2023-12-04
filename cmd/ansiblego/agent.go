package main

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/state-of-the-art/ansiblego/pkg/ansible"
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
		if err := core.ReadConfigFile(cfg, p_cfg_path); err != nil {
			return log.Error("Unable to apply config file:", p_cfg_path, err)
		}
		cfg.Verbosity = log.Verbosity

		log.Debug("AgentConfig:", cfg)

		log.Debug("AnsibleGo initialized")

		// TODO: Run the facts collector for now
		for _, mod := range ansible.ModulesList("fact") {
			if collected_facts, err := ansible.CollectV1(mod); err == nil {
				y, err := ansible.ToYaml(collected_facts)
				if err == nil {
					log.Infof("Facts from '%s': %s", mod, y)
				} else {
					log.Infof("Facts from '%s': %v", mod, err)
				}
			} else {
				log.Infof("Error while collecting facts from '%s': %s", mod, err)
			}
		}

		_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		log.Info("AnsibleGo exiting...")

		return nil
	},
}

func init() {
	agent_cmd.Flags().SortFlags = false
	root_cmd.AddCommand(agent_cmd)
}
