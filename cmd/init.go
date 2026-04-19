package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/lorenjphillips/skill-vault/internal/config"
	"github.com/lorenjphillips/skill-vault/internal/detect"
	"github.com/lorenjphillips/skill-vault/internal/schedule"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")).
			MarginBottom(1)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactive setup — detect tools, choose backups, configure schedule",
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	fmt.Println(titleStyle.Render("skill-vault"))
	fmt.Println(dimStyle.Render("Back up your AI agent skills, config, and conversations.\n"))

	if config.Exists() {
		var overwrite bool
		huh.NewConfirm().
			Title("Existing config found. Overwrite?").
			Value(&overwrite).
			Run()
		if !overwrite {
			fmt.Println("Keeping existing config.")
			return nil
		}
	}

	fmt.Println(dimStyle.Render("Scanning for AI tools..."))
	tools := detect.Scan()

	if len(tools) == 0 {
		fmt.Println("No AI tools detected. Nothing to back up.")
		return nil
	}

	fmt.Println()
	for _, t := range tools {
		fmt.Printf("  %s %s %s\n",
			successStyle.Render("✓"),
			t.Description,
			dimStyle.Render(fmt.Sprintf("(%s)", detect.FormatSize(t.DiskSize))))
	}
	fmt.Println()

	selectedTools, err := selectTools(tools)
	if err != nil {
		return err
	}
	if len(selectedTools) == 0 {
		fmt.Println("No tools selected.")
		return nil
	}

	toolConfigs, err := selectCategories(tools, selectedTools)
	if err != nil {
		return err
	}

	githubCfg, err := configureGitHub()
	if err != nil {
		return err
	}

	s3Cfg, err := configureS3(toolConfigs)
	if err != nil {
		return err
	}

	scheduleCfg, err := configureSchedule()
	if err != nil {
		return err
	}

	cfg := &config.Config{
		Tools:    toolConfigs,
		GitHub:   githubCfg,
		S3:       s3Cfg,
		Schedule: scheduleCfg,
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Println()
	fmt.Println(successStyle.Render("Config saved to " + config.Path()))

	if scheduleCfg.Enabled {
		if err := schedule.Install(scheduleCfg.Interval); err != nil {
			return fmt.Errorf("installing schedule: %w", err)
		}
		fmt.Println(successStyle.Render("Scheduled sync installed (launchd)"))
	}

	fmt.Println()
	fmt.Println(dimStyle.Render("Run 'skill-vault sync' to back up now."))
	return nil
}

func selectTools(tools []detect.Tool) ([]string, error) {
	options := make([]huh.Option[string], len(tools))
	for i, t := range tools {
		label := fmt.Sprintf("%s (%s)", t.Description, detect.FormatSize(t.DiskSize))
		options[i] = huh.NewOption(label, t.Name).Selected(true)
	}

	var selected []string
	err := huh.NewMultiSelect[string]().
		Title("Which tools do you want to back up?").
		Options(options...).
		Value(&selected).
		Run()

	return selected, err
}

func selectCategories(tools []detect.Tool, selectedNames []string) (map[string]config.ToolConfig, error) {
	result := make(map[string]config.ToolConfig)
	selectedSet := make(map[string]bool)
	for _, n := range selectedNames {
		selectedSet[n] = true
	}

	for _, t := range tools {
		if !selectedSet[t.Name] {
			continue
		}

		categories := make(map[string]bool)
		for _, p := range t.Paths {
			categories[string(p.Category)] = true
		}

		if len(categories) <= 1 {
			cats := make([]string, 0, len(categories))
			for c := range categories {
				cats = append(cats, c)
			}
			result[t.Name] = config.ToolConfig{Enabled: true, Categories: cats}
			continue
		}

		options := make([]huh.Option[string], 0)
		for c := range categories {
			options = append(options, huh.NewOption(c, c).Selected(true))
		}

		var selected []string
		err := huh.NewMultiSelect[string]().
			Title(fmt.Sprintf("What to back up from %s?", t.Description)).
			Options(options...).
			Value(&selected).
			Run()
		if err != nil {
			return nil, err
		}

		result[t.Name] = config.ToolConfig{Enabled: true, Categories: selected}
	}

	return result, nil
}

func configureGitHub() (config.GitHubConfig, error) {
	var cfg config.GitHubConfig

	err := huh.NewConfirm().
		Title("Back up to a GitHub repository?").
		Description("Skills, config, and memory will be synced to a git repo").
		Value(&cfg.Enabled).
		Run()
	if err != nil || !cfg.Enabled {
		return cfg, err
	}

	home, _ := os.UserHomeDir()
	err = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("GitHub repo URL").
				Placeholder("git@github.com:you/ai-backup.git").
				Value(&cfg.Repo),
			huh.NewInput().
				Title("Local clone path").
				Placeholder(home+"/Development/ai-backup").
				Value(&cfg.LocalPath),
		),
	).Run()

	if cfg.LocalPath == "" {
		cfg.LocalPath = home + "/Development/ai-backup"
	}

	return cfg, err
}

func configureS3(toolConfigs map[string]config.ToolConfig) (config.S3Config, error) {
	var cfg config.S3Config

	hasConversations := false
	for _, t := range toolConfigs {
		for _, c := range t.Categories {
			if c == "conversations" {
				hasConversations = true
			}
		}
	}

	if !hasConversations {
		return cfg, nil
	}

	err := huh.NewConfirm().
		Title("Back up conversation logs to S3?").
		Description("Compressed daily snapshots — too large for git").
		Value(&cfg.Enabled).
		Run()
	if err != nil || !cfg.Enabled {
		return cfg, err
	}

	cfg.Region = "us-east-2"
	err = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("S3 bucket name").
				Placeholder("my-ai-backups").
				Value(&cfg.Bucket),
			huh.NewInput().
				Title("AWS CLI profile").
				Placeholder("default").
				Value(&cfg.Profile),
			huh.NewInput().
				Title("AWS region").
				Value(&cfg.Region),
		),
	).Run()

	if cfg.Profile == "" {
		cfg.Profile = "default"
	}

	return cfg, err
}

func configureSchedule() (config.ScheduleConfig, error) {
	var cfg config.ScheduleConfig

	err := huh.NewConfirm().
		Title("Set up automatic backups?").
		Description("Creates a macOS launchd job to sync on a schedule").
		Value(&cfg.Enabled).
		Run()
	if err != nil || !cfg.Enabled {
		return cfg, err
	}

	var intervalChoice string
	err = huh.NewSelect[string]().
		Title("How often should skill-vault sync?").
		Options(
			huh.NewOption("Every 6 hours", "6h"),
			huh.NewOption("Every 12 hours", "12h"),
			huh.NewOption("Every 24 hours (daily)", "24h"),
			huh.NewOption("Every 48 hours", "48h"),
			huh.NewOption("Custom", "custom"),
		).
		Value(&intervalChoice).
		Run()
	if err != nil {
		return cfg, err
	}

	if intervalChoice == "custom" {
		err = huh.NewInput().
			Title("Custom interval").
			Description("Go duration format: e.g. 8h, 72h (minimum 1h)").
			Placeholder("24h").
			Value(&intervalChoice).
			Run()
		if err != nil {
			return cfg, err
		}
	}

	cfg.Interval = intervalChoice
	_ = strings.TrimSpace(cfg.Interval)
	return cfg, nil
}
