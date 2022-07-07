package main

import (
	"log"

	"github.com/spf13/cobra"
)

var agent_cmd = &cobra.Command{
	Use:   "agent",
	Short: "Executes the provided task/s",
	Long:  "It runs remotely to execute the requests and is usable to test commands",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Println("AnsibleGo Agent running...")
		return nil
	},
}

func init() {
	agent_cmd.Flags().SortFlags = false
	root_cmd.AddCommand(agent_cmd)
}
