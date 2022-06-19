package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/state-of-the-art/ansiblego/pkg/core"
)

func main() {
	var playbook_mode bool
	var agent_mode bool
	var extra_vars *[]string
	var cfg_path string
	var verbose int

	cmd := &cobra.Command{
		Use:   "ansiblego",
		Short: "AnsibleGo simple image config management",
		Long:  `The simple replacement for ansible to build images purpose`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Println("AnsibleGo running...")

			cfg := &core.Config{}
			if err := cfg.ReadConfigFile(cfg_path); err != nil {
				log.Println("Unable to apply config file:", cfg_path, err)
				return err
			}
			if playbook_mode {
				cfg.Mode = "playbook"
			}
			if agent_mode {
				cfg.Mode = "agent"
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

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.BoolVar(&playbook_mode, "playbook", false, "execute in playbook mode")
	flags.BoolVar(&agent_mode, "agent", false, "execute in agent mode")
	extra_vars = flags.StringSliceP("extra-vars", "e", nil, "additional variables")
	flags.StringVarP(&cfg_path, "cfg", "c", "", "yaml configuration file")
	flags.IntVarP(&verbose, "verbose", "v", -1000, "verbose logging level: >0 more verbose, <0 less")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
