package cmd

import (
	"fmt"
	"os/exec"

	"github.com/jerryagbesi/skipper/internal/connect"
	"github.com/jerryagbesi/skipper/internal/sshconfig"
	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:   "connect <alias>",
	Short: "Connect directly to a host by alias",
	Long: `Connect to a host by its SSH config alias without opening the interactive picker.

Useful for scripting or when you already know the alias.`,
	Example: `  skipper connect bastion
  skipper connect devone`,
	Args: cobra.ExactArgs(1),
	RunE: runConnect,
}

func runConnect(_ *cobra.Command, args []string) error {
	path, err := resolveConfigPath(configPath)
	if err != nil {
		return err
	}

	alias := args[0]
	host, err := findHostByAlias(path, alias)
	if err != nil {
		return err
	}

	return connect.Connect(host, exec.Command)
}

func findHostByAlias(path, alias string) (*sshconfig.Host, error) {
	hosts, err := sshconfig.ParseHosts(path)
	if err != nil {
		return nil, err
	}

	for i, h := range hosts {
		if h.Alias == alias {
			return &hosts[i], nil
		}
	}

	return nil, fmt.Errorf("no host with alias %q found in %s", alias, path)
}

func init() {
	rootCmd.AddCommand(connectCmd)
}
