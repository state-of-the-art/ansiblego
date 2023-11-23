package main

import (
	"context"
	"log"
	"time"

	"github.com/spf13/cobra"

	"github.com/state-of-the-art/ansiblego/pkg/core"
)

var agent_cmd = &cobra.Command{
	Use:   "agent",
	Short: "Executes the provided task/s",
	Long:  "It runs remotely to execute the requests and is usable to test commands",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Println("AnsibleGo Agent running...")

		cfg := &core.Config{}
		if err := cfg.ReadConfigFile(cfg_path); err != nil {
			log.Println("Unable to apply config file:", cfg_path, err)
			return err
		}
		if verbose != -1000 {
			cfg.Verbose = verbose
		}

		ango, err := core.New(cfg)
		if err != nil {
			return err
		}

		log.Println("DEBUG:", cfg)

		log.Println("AnsibleGo initialized")

		// For now use it as test runner
		/*for _, task_data := range args {
			ango.Agent(task_data)
		}*/

		_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		ango.Close()

		log.Println("AnsibleGo exiting...")

		return nil
	},
}

func init() {
	agent_cmd.Flags().SortFlags = false
	root_cmd.AddCommand(agent_cmd)
}
