package main

import (
	"context"
	"log"
	"time"

	"github.com/spf13/cobra"

	"github.com/state-of-the-art/ansiblego/pkg/core"
)

var extra_vars *[]string

var playbook_cmd = &cobra.Command{
	Use:   "playbook",
	Short: "Runs the playbooks you have",
	Long:  "Makes it possible to apply the playbook configuration to the required system",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Println("AnsibleGo Playbook running...")
		cfg := &core.Config{}
		if err := cfg.ReadConfigFile(cfg_path); err != nil {
			log.Println("Unable to apply config file:", cfg_path, err)
			return err
		}
		if len(*extra_vars) > 0 {
			cfg.ExtraVars = *extra_vars
		}
		if len(args) > 0 {
			cfg.Args = args
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

		_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		ango.Close()

		log.Println("AnsibleGo exiting...")

		return nil
	},
}

func init() {
	playbook_cmd.Flags().SortFlags = false
	extra_vars = playbook_cmd.Flags().StringSliceP("extra-vars", "e", nil, "additional variables")
	root_cmd.AddCommand(playbook_cmd)
}
