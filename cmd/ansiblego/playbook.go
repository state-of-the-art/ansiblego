package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/state-of-the-art/ansiblego/pkg/ansible"
	"github.com/state-of-the-art/ansiblego/pkg/ansible/inventory"
	"github.com/state-of-the-art/ansiblego/pkg/core"
	"github.com/state-of-the-art/ansiblego/pkg/log"
)

var p_extra_vars *[]string
var p_skip_tags *[]string
var p_inventory *[]string

var playbook_cmd = &cobra.Command{
	Use:     "playbook",
	Version: "0.1",
	Short:   "Runs the playbooks you have",
	Long:    "Makes it possible to apply the playbook configuration to the host system",

	Args: cobra.MinimumNArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		log.Info("AnsibleGo Playbook running...")
		cfg := &core.PlaybookConfig{}
		err := core.ReadConfigFile(cfg, p_cfg_path)
		if err != nil {
			return log.Error("Unable to apply config file:", p_cfg_path, err)
		}
		cfg.Verbosity = log.Verbosity
		if len(*p_extra_vars) > 0 {
			cfg.ExtraVars = *p_extra_vars
		}
		if len(*p_skip_tags) > 0 {
			cfg.SkipTags = *p_skip_tags
		}
		if len(*p_inventory) > 0 {
			if cfg.Inventory, err = inventory.New(*p_inventory); err != nil {
				return log.Errorf("Unable to process provided inventory: %v", err)
			}
		}

		ango, err := core.New(&cfg.CommonConfig)
		if err != nil {
			return err
		}

		log.Debugf("PlaybookConfig: %q", cfg)

		log.Debug("AnsibleGo initialized")

		//
		// Start parsing stage
		//
		var playbooks_to_apply []ansible.Playbook

		// Loading modules to use them for parsing of the playbook
		ansible.InitEmbeddedModules()
		// TODO: Load external modules from user as well

		for _, pb_path := range args {
			log.Debug("Loading playbook:", pb_path)
			pf := ansible.PlaybookFile{}
			err := pf.Load(pb_path)
			if err != nil {
				return log.Error("Unable to load PlaybookFile:", err)
			}
			log.Debug("Parsed PlaybookFile", pb_path)

			// Making sure it's possible to represent the parsed playbooks
			yaml, err := pf.Yaml()
			if err != nil {
				return log.Error("Unable to represent PlaybookFile in YAML format:", err)
			}
			log.Tracef("PlaybookFile %s:\n%s", pb_path, yaml)
			playbooks_to_apply = append(playbooks_to_apply, pf...)
		}

		//
		// Start execute stage
		//
		if cfg.Inventory == nil || len(cfg.Inventory.Hosts) < 1 {
			log.Info("Inventory hosts are not specified, skipping playbooks execution")
			log.Debugf("Inv: %q", cfg.Inventory)
			return nil
		}

		for _, host := range cfg.Inventory.Hosts {
			for _, pb := range playbooks_to_apply {
				if err := pb.Run(host); err != nil {
					return fmt.Errorf("Playbook execution error: %v", err)
				}
			}
		}

		_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		ango.Close()

		log.Info("AnsibleGo exiting...")

		return fmt.Errorf("Not yet ready")
	},
}

func init() {
	playbook_cmd.Flags().SortFlags = false
	p_extra_vars = playbook_cmd.Flags().StringArrayP("extra-vars", "e", nil, "set additional variables as key=value or YAML/JSON, if filename prepend with @ (can occur multiple times)")
	p_skip_tags = playbook_cmd.Flags().StringSlice("skip-tags", nil, "only run plays and tasks whose tags do not match these values (comma-separated)")
	p_inventory = playbook_cmd.Flags().StringSliceP("inventory", "i", nil, "specify inventory host path or comma separated host list.")
	root_cmd.AddCommand(playbook_cmd)
}
