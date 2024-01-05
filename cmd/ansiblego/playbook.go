package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/state-of-the-art/ansiblego/pkg/ansible"
	"github.com/state-of-the-art/ansiblego/pkg/ansible/inventory"
	"github.com/state-of-the-art/ansiblego/pkg/core"
	"github.com/state-of-the-art/ansiblego/pkg/log"
	"github.com/state-of-the-art/ansiblego/pkg/util"
)

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
		if p_extra_vars != nil {
			// Parsing extra variables and store to cfg
			cfg.ExtraVars = make(map[string]any)
			for _, keyval := range *p_extra_vars {
				extra_vars_data, err := util.ParseArgument("extra var", keyval, nil)
				if err != nil {
					return log.Errorf("Unable to process provided task arg %q: %v", keyval, err)
				}
				for key, val := range extra_vars_data {
					cfg.ExtraVars[key] = val
					log.Tracef("Provided extra var: %q=%v", key, val)
				}
			}
		}
		if p_skip_tags != nil {
			cfg.SkipTags = *p_skip_tags
		}
		if p_inventory != nil {
			if cfg.Inventory, err = inventory.New(*p_inventory); err != nil {
				return log.Errorf("Unable to process provided inventory: %v", err)
			}
		}

		ango, err := core.New(&cfg.CommonConfig)
		if err != nil {
			return err
		}

		log.Debug("AnsibleGo initialized")

		// Loading modules to use them for parsing of the playbook
		ansible.InitEmbeddedModules()
		// TODO: Load external modules from user as well

		//
		// Start parsing stage
		//
		var playbooks_to_apply []ansible.Playbook

		for _, pb_path := range args {
			log.Debug("Loading playbook:", pb_path)
			pf := ansible.PlaybookFile{}
			err := pf.Load(pb_path)
			if err != nil {
				return log.Error("Unable to load PlaybookFile:", err)
			}
			log.Debug("Parsed PlaybookFile", pb_path)

			// Making sure it's possible to represent the parsed playbooks
			y, err := pf.Yaml()
			if err != nil {
				return log.Error("Unable to represent PlaybookFile in YAML format:", err)
			}
			log.Tracef("PlaybookFile %s:\n%s", pb_path, y)
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
				if err := pb.Run(cfg, host); err != nil {
					return fmt.Errorf("Playbook execution error: %v", err)
				}
			}
		}

		ango.Close()

		log.Info("AnsibleGo exiting...")

		return fmt.Errorf("Not yet ready")
	},
}

func init() {
	playbook_cmd.Flags().SortFlags = false
	p_skip_tags = playbook_cmd.Flags().StringSlice("skip-tags", nil, "only run plays and tasks whose tags do not match these values (comma-separated)")
	p_inventory = playbook_cmd.Flags().StringSliceP("inventory", "i", nil, "specify inventory host path or comma separated host list.")
	root_cmd.AddCommand(playbook_cmd)
}
