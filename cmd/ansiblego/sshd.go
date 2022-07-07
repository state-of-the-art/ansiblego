package main

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/state-of-the-art/ansiblego/pkg/sshd"
)

var sshd_user string
var sshd_password string
var sshd_address string

var sshd_cmd = &cobra.Command{
	Use:   "sshd",
	Short: "Provides you a transport to the remote system",
	Long:  "An embedded cross-platform ssh server to simplify setup of the minimal environments",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Printf("AnsibleGo SSHD running on %s...\n", sshd_address)
		sshd.Run(sshd_user, sshd_password, sshd_address)
		return nil
	},
}

func init() {
	sshd_cmd.Flags().SortFlags = false
	sshd_cmd.Flags().StringVarP(&sshd_user, "user", "u", "", "sshd username for login or token if password is not set")
	sshd_cmd.Flags().StringVarP(&sshd_password, "password", "p", "", "sshd password for login")
	sshd_cmd.Flags().StringVarP(&sshd_address, "address", "a", "0.0.0.0:2222", "sshd address to listen on")
	root_cmd.AddCommand(sshd_cmd)
}
