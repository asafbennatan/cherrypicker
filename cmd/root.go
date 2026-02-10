package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cherrypicker",
	Short: "A CLI tool for cherry-picking GitHub pull requests",
	Long:  `Cherrypicker is a CLI tool that interacts with the GitHub API to help manage and cherry-pick pull requests across branches.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
