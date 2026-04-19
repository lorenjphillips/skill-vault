package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "skill-vault",
	Short: "Back up your AI agent skills, config, and conversation logs",
	Long: `skill-vault detects installed AI coding tools (Claude Code, Cursor, Codex,
Windsurf, Aider, Continue, Copilot) and backs up their skills, configs,
memory, and conversation logs to GitHub and/or S3.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
