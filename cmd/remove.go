package cmd

import (
	"fmt"

	"github.com/jerryagbesi/skipper/internal/sshconfig"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove <alias>",
	Aliases: []string{"rm"},
	Short:   "Remove a host entry from the SSH config",
	Long: `Remove the SSH config block matching <alias>.

Exits with an error when the alias is not found unless --force is supplied.`,
	Example: `  skipper remove devone
  skipper rm bastion
  skipper remove nonexistent --force`,
	Args: cobra.ExactArgs(1),
	RunE: runRemove,
}

var removeForce bool

func runRemove(_ *cobra.Command, args []string) error {
	path, err := resolveConfigPath(configPath)
	if err != nil {
		return err
	}

	alias := args[0]
	removed, err := sshconfig.RemoveHost(path, alias)
	if err != nil {
		return err
	}

	if !removed {
		if removeForce {
			return nil
		}
		return fmt.Errorf("host %q not found in %s", alias, path)
	}

	fmt.Printf("removed host %q\n", alias)
	return nil
}

func init() {
	removeCmd.Flags().BoolVarP(&removeForce, "force", "F", false, "exit 0 even when alias is not found")
	rootCmd.AddCommand(removeCmd)
}
