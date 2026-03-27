package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate additional documentation and/or completions",
}

var generateShellCmd = &cobra.Command{
	Use:   "shell [shell]",
	Short: "Generate shell completions for the given shell to stdout",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		shell := args[0]
		var err error

		switch shell {
		case "bash":
			err = rootCmd.GenBashCompletionV2(cmd.OutOrStdout(), true)
		case "zsh":
			err = rootCmd.GenZshCompletion(cmd.OutOrStdout())
		case "fish":
			err = rootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
		case "powershell":
			err = rootCmd.GenPowerShellCompletion(cmd.OutOrStdout())
		default:
			return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish, powershell)", shell)
		}

		return err
	},
}

var generateManCmd = &cobra.Command{
	Use:   "man",
	Short: "Generate a man page for gjq to output directory if specified, else the current directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputDir, err := cmd.Flags().GetString("output-dir")
		if err != nil {
			return err
		}
		if outputDir == "" {
			outputDir, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("getting current directory: %w", err)
			}
		}

		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}

		header := &doc.GenManHeader{
			Title:   "GJQ",
			Section: "1",
		}

		if err := doc.GenManTree(rootCmd, header, outputDir); err != nil {
			return fmt.Errorf("generating man pages: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Generated man pages in %s\n", outputDir)
		return nil
	},
}

func init() {
	generateManCmd.Flags().StringP("output-dir", "o", "", "The output directory to write the man pages")

	generateCmd.AddCommand(generateShellCmd)
	generateCmd.AddCommand(generateManCmd)
	rootCmd.AddCommand(generateCmd)
}
