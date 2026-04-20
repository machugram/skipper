package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func newManCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "man [directory]",
		Short:   "generate man pages",
		Example: "skipper man ./dist/man",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			outDir := "dist/man"
			if len(args) == 1 {
				outDir = args[0]
			}

			if err := generateManPages(outDir); err != nil {
				return err
			}

			fmt.Printf("generated man pages in %s\n", outDir)
			return nil
		},
	}
}

func generateManPages(outDir string) error {
	outDir = strings.TrimSpace(outDir)
	if outDir == "" {
		outDir = "dist/man"
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create man output directory: %w", err)
	}

	rootCmd.DisableAutoGenTag = true

	header := &doc.GenManHeader{
		Title:   "SKIPPER",
		Section: "1",
		Source:  version,
		Manual:  "Skipper Manual",
	}

	if err := doc.GenManTree(rootCmd, header, outDir); err != nil {
		return fmt.Errorf("generate man pages: %w", err)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(newManCmd())
}
