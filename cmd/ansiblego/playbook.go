package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"

	"github.com/state-of-the-art/ansiblego/pkg/core"
)

var extra_vars *[]string
var skip_tags *[]string
var inventory *[]string

var playbook_cmd = &cobra.Command{
	Use:     "playbook",
	Version: "0.1",
	Short:   "Runs the playbooks you have",
	Long:    "Makes it possible to apply the playbook configuration to the required system",

	Args: cobra.MinimumNArgs(1),

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
		if len(*skip_tags) > 0 {
			cfg.SkipTags = *skip_tags
		}
		if len(*inventory) > 0 {
			cfg.Inventory = *inventory
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

		for _, pb_path := range args {
			if err := ango.Playbook(pb_path); err != nil {
				return err
			}
		}

		_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		ango.Close()

		log.Println("AnsibleGo exiting...")

		return fmt.Errorf("Not yet ready")
	},
}

func init() {
	playbook_cmd.Flags().SortFlags = false
	extra_vars = playbook_cmd.Flags().StringArrayP("extra-vars", "e", nil, "set additional variables as key=value or YAML/JSON, if filename prepend with @ (can occur multiple times)")
	skip_tags = playbook_cmd.Flags().StringSlice("skip-tags", nil, "only run plays and tasks whose tags do not match these values (comma-separated)")
	inventory = playbook_cmd.Flags().StringSliceP("inventory", "i", nil, "specify inventory host path or comma separated host list.")
	root_cmd.AddCommand(playbook_cmd)
}
