package detect

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type PathCategory string

const (
	CategorySkills        PathCategory = "skills"
	CategoryConfig        PathCategory = "config"
	CategoryConversations PathCategory = "conversations"
	CategoryMemory        PathCategory = "memory"
	CategoryRules         PathCategory = "rules"
)

type BackupPath struct {
	Category PathCategory
	Path     string
	Pattern  string
}

type Tool struct {
	Name        string
	Description string
	Dir         string
	AltDirs     []string
	Paths       []BackupPath
	Detected    bool
	DiskSize    int64
}

// KnownTools is the list of 16 known AI coding tools (Claude Dev removed as duplicate of Cline).
var KnownTools = []Tool{
	{
		Name:        "claude",
		Description: "Claude Code",
		Dir:         "~/.claude",
		Paths: []BackupPath{
			{CategorySkills, "~/.claude/skills", ""},
			{CategorySkills, "~/.claude/agents", ""},
			{CategorySkills, "~/.claude/commands", ""},
			{CategoryMemory, "~/.claude/projects", "*/memory/*.md"},
			{CategoryConfig, "~/.claude/settings.json", ""},
			{CategoryConfig, "~/.claude/company", ""},
			{CategoryConfig, "~/.claude/bin", ""},
			{CategoryConversations, "~/.claude/projects", "*.jsonl"},
		},
	},
	{
		Name:        "cursor",
		Description: "Cursor",
		Dir:         "~/.cursor",
		Paths: []BackupPath{
			{CategoryRules, "~/.cursor/rules", ""},
			{CategoryConfig, "~/.cursor/mcp.json", ""},
		},
	},
	{
		Name:        "codex",
		Description: "OpenAI Codex CLI",
		Dir:         "~/.codex",
		Paths: []BackupPath{
			{CategorySkills, "~/.codex/skills", ""},
			{CategoryMemory, "~/.codex/memories", ""},
			{CategoryConfig, "~/.codex/config.toml", ""},
			{CategoryConfig, "~/.codex/config.yaml", ""},
			{CategoryConversations, "~/.codex/sessions", ""},
		},
	},
	{
		Name:        "windsurf",
		Description: "Windsurf (Codeium)",
		Dir:         "~/.codeium/windsurf",
		Paths: []BackupPath{
			{CategoryMemory, "~/.codeium/windsurf/memories", ""},
			{CategoryRules, "~/.codeium/windsurf/rules", ""},
			{CategoryConfig, "~/.codeium/windsurf/mcp_config.json", ""},
		},
	},
	{
		Name:        "aider",
		Description: "Aider",
		Dir:         "~/.aider",
		Paths: []BackupPath{
			{CategoryConversations, "~/.aider/chat-history", ""},
			{CategoryConfig, "~/.aider.conf.yml", ""},
		},
	},
	{
		Name:        "continue-dev",
		Description: "Continue",
		Dir:         "~/.continue",
		Paths: []BackupPath{
			{CategoryConfig, "~/.continue/config.json", ""},
			{CategoryConfig, "~/.continue/config.ts", ""},
			{CategoryConfig, "~/.continue/config.yaml", ""},
			{CategoryRules, "~/.continue/rules", ""},
			{CategoryConversations, "~/.continue/sessions", ""},
		},
	},
	{
		Name:        "copilot",
		Description: "GitHub Copilot",
		Dir:         "~/.config/github-copilot",
		Paths: []BackupPath{
			{CategoryConfig, "~/.config/github-copilot", ""},
		},
	},
	{
		Name:        "amp",
		Description: "Amp (Sourcegraph)",
		Dir:         "~/.amp",
		Paths: []BackupPath{
			{CategoryConfig, "~/.amp/config.yaml", ""},
			{CategoryConversations, "~/.amp/threads", ""},
		},
	},
	{
		Name:        "cline",
		Description: "Cline",
		Dir:         "~/.cline",
		Paths: []BackupPath{
			{CategoryConfig, "~/.cline/config.json", ""},
			{CategoryRules, "~/.cline/rules", ""},
			{CategoryConversations, "~/.cline/tasks", ""},
		},
	},
	{
		Name:        "roo-code",
		Description: "Roo Code",
		Dir:         "~/.roo-code",
		AltDirs:     []string{"~/.roo"},
		Paths: []BackupPath{
			{CategoryConfig, "~/.roo-code/config.json", ""},
			{CategoryRules, "~/.roo-code/rules", ""},
			{CategoryConversations, "~/.roo-code/tasks", ""},
		},
	},
	{
		Name:        "tabnine",
		Description: "Tabnine",
		Dir:         "~/.tabnine",
		Paths: []BackupPath{
			{CategoryConfig, "~/.tabnine/tabnine_config.json", ""},
		},
	},
	{
		Name:        "supermaven",
		Description: "Supermaven",
		Dir:         "~/.supermaven",
		Paths: []BackupPath{
			{CategoryConfig, "~/.supermaven", ""},
		},
	},
	{
		Name:        "zed-ai",
		Description: "Zed AI",
		Dir:         "~/.config/zed",
		Paths: []BackupPath{
			{CategoryConfig, "~/.config/zed/settings.json", ""},
			{CategoryConfig, "~/.config/zed/keymap.json", ""},
			{CategoryConfig, "~/.config/zed/tasks.json", ""},
			{CategoryRules, "~/.config/zed/rules", ""},
			{CategoryConversations, "~/.config/zed/conversations", ""},
		},
	},
	{
		Name:        "warp",
		Description: "Warp AI",
		Dir:         "~/.warp",
		Paths: []BackupPath{
			{CategoryConfig, "~/.warp/themes", ""},
			{CategoryConfig, "~/.warp/workflows", ""},
		},
	},
	{
		Name:        "amazon-q",
		Description: "Amazon Q Developer",
		Dir:         "~/.aws/amazonq",
		Paths: []BackupPath{
			{CategoryConfig, "~/.aws/amazonq", ""},
		},
	},
	{
		Name:        "gemini-cli",
		Description: "Gemini CLI",
		Dir:         "~/.gemini",
		Paths: []BackupPath{
			{CategoryConfig, "~/.gemini/settings.json", ""},
			{CategoryConfig, "~/.gemini/GEMINI.md", ""},
			{CategoryConversations, "~/.gemini/history", ""},
		},
	},
}

func ExpandHome(path string) string {
	if len(path) > 1 && path[:2] == "~/" {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func Scan() []Tool {
	var found []Tool
	for _, tool := range KnownTools {
		t := tool
		detected := false

		// Check primary dir
		dir := ExpandHome(t.Dir)
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			detected = true
		}

		// Check alternate dirs if primary not found
		if !detected {
			for _, altDir := range t.AltDirs {
				expanded := ExpandHome(altDir)
				if info, err := os.Stat(expanded); err == nil && info.IsDir() {
					detected = true
					break
				}
			}
		}

		if !detected {
			continue
		}

		t.Detected = true
		t.DiskSize = dirSize(dir)

		var validPaths []BackupPath
		for _, p := range t.Paths {
			expanded := ExpandHome(p.Path)
			if _, err := os.Stat(expanded); err == nil {
				validPaths = append(validPaths, p)
			}
		}
		t.Paths = validPaths
		if len(validPaths) == 0 {
			continue
		}
		found = append(found, t)
	}
	return found
}

func dirSize(path string) int64 {
	var size int64
	_ = filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return nil
			}
			size += info.Size()
		}
		return nil
	})
	return size
}

func FormatSize(bytes int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case bytes >= gb:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
