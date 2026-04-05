package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/jerryagbesi/skipper/internal/connect"
	"github.com/jerryagbesi/skipper/internal/sshconfig"
	"github.com/jerryagbesi/skipper/internal/ui"

	"github.com/spf13/cobra"
)

var configPath string

var rootCmd = &cobra.Command{
	Use:     "skipper <command> [flags]",
	Version: "v1.0:beta",
	Short:   "skipper is a cli tool for managing ssh connections",
	Example: "skipper --version",
	Run: func(cmd *cobra.Command, args []string) {
		if configPath == "" {
			var err error
			configPath, err = sshconfig.DefaultConfigPath()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		hosts, err := sshconfig.ParseHosts(configPath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if len(hosts) == 0 {
			fmt.Println("no hosts found in config file")
			os.Exit(1)
		}

		// Render bubbletea UI and get selected host
		result, err := ui.Run(hosts)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if result.Cancelled {
			return
		}

		// Connect to selected host
		err = connect.Connect(result.Host, exec.Command)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

	},
	SilenceErrors: true,
	Long: `skipper is a cli tool for managing ssh connections, It allows you to select your preferred ssh host alias, connect to it, and execute commands.

Usage: skipper <command> [flags]

Flags:

	-c, --config string   path to ssh config file, defaults to ~/.ssh/config
	-h, --help            help for skipper
	-v, --version         print version information`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// rootCmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
	// 	return fmt.Errorf("invalid flag: %w \n please run 'skipper --help' for usage information", err)
	// })

	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "path to ssh config file, defaults to ~/.ssh/config")
	rootCmd.Flags().BoolP("version", "v", false, "print version information")
}
