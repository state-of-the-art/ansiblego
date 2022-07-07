package main

import (
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var (
	cfg_path string
	verbose  int
	detach   bool

	root_cmd = &cobra.Command{
		Use:   "ansiblego",
		Short: "AnsibleGo simple image configurator",
		Long:  "A simplest replacement for ansible to build your images",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Println("AnsibleGo running...")
			return nil
		},
	}
)

func main() {
	cobra.EnableCommandSorting = false

	// Check if the first argument is a registered command with prefix "--"
	// This is needed for backward compatibility with ansible cmd interfaces
	if len(os.Args) > 1 {
		for _, cmd := range root_cmd.Commands() {
			tmp := []string{"--" + cmd.Name()}
			cmd.Aliases = append(tmp, cmd.Aliases...)
			if os.Args[1][:2] == "--" && os.Args[1][2:] == cmd.Name() {
				os.Args[1] = cmd.Name()
				break
			}
		}
	}

	flags := root_cmd.Flags()
	flags.SortFlags = false

	// Common flags
	flags.StringVarP(&cfg_path, "cfg", "c", "", "yaml configuration file")
	flags.IntVarP(&verbose, "verbose", "v", -1000, "verbose logging level: >0 more verbose, <0 less")
	flags.BoolVar(&detach, "detach", false, "detach from shell to background")

	if err := root_cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// Rerun the same application in background
func rerunDetached() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	var args []string
	for _, item := range os.Args {
		if item != "--detach" {
			args = append(args, item)
		}
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = cwd
	err = cmd.Start()
	if err != nil {
		return err
	}
	cmd.Process.Release()

	return nil
}
