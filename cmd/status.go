package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/lorenjphillips/skill-vault/internal/config"
	"github.com/lorenjphillips/skill-vault/internal/detect"
	"github.com/lorenjphillips/skill-vault/internal/schedule"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show backup configuration and health",
	RunE:  runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	fmt.Println(titleStyle.Render("skill-vault status"))

	cfg, err := config.Load()
	if err != nil {
		fmt.Println(dimStyle.Render("Not configured — run 'skill-vault init'"))
		fmt.Println()
		fmt.Println("Detected tools:")
		for _, t := range detect.Scan() {
			fmt.Printf("  %s %s %s\n",
				successStyle.Render("✓"),
				t.Description,
				dimStyle.Render(detect.FormatSize(t.DiskSize)))
		}
		return nil
	}

	fmt.Println("Tools:")
	for name, t := range cfg.Tools {
		if t.Enabled {
			fmt.Printf("  %s %s → %v\n", successStyle.Render("✓"), name, t.Categories)
		}
	}

	fmt.Println()
	fmt.Println("Backup targets:")
	if cfg.GitHub.Enabled {
		fmt.Printf("  %s GitHub: %s\n", successStyle.Render("✓"), cfg.GitHub.Repo)
	} else {
		fmt.Printf("  %s GitHub: disabled\n", dimStyle.Render("○"))
	}
	if cfg.S3.Enabled {
		fmt.Printf("  %s S3: %s (profile: %s)\n", successStyle.Render("✓"), cfg.S3.Bucket, cfg.S3.Profile)
	} else {
		fmt.Printf("  %s S3: disabled\n", dimStyle.Render("○"))
	}

	fmt.Println()
	fmt.Println("Schedule:")
	fmt.Printf("  Status: %s\n", schedule.Status())
	fmt.Printf("  Last sync: %s\n", schedule.LastRun())
	if cfg.Schedule.Enabled {
		fmt.Printf("  Interval: %s\n", cfg.Schedule.Interval)
	}

	return nil
}
