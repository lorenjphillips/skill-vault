package backup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/lorenjphillips/skill-vault/internal/config"
	"github.com/lorenjphillips/skill-vault/internal/detect"
)

func expandHome(path string) string {
	if len(path) > 1 && path[:2] == "~/" {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func Run(cfg *config.Config) error {
	if cfg.GitHub.Enabled {
		if err := syncGitHub(cfg); err != nil {
			return fmt.Errorf("github sync: %w", err)
		}
	}
	if cfg.S3.Enabled {
		if err := syncS3(cfg); err != nil {
			return fmt.Errorf("s3 sync: %w", err)
		}
	}
	return nil
}

func syncGitHub(cfg *config.Config) error {
	repoDir := expandHome(cfg.GitHub.LocalPath)

	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		fmt.Printf("  Cloning %s...\n", cfg.GitHub.Repo)
		if err := run("git", "clone", cfg.GitHub.Repo, repoDir); err != nil {
			return fmt.Errorf("clone: %w", err)
		}
	}

	if err := runIn(repoDir, "git", "stash", "push", "--include-untracked", "-m",
		fmt.Sprintf("skill-vault auto-stash %s", time.Now().Format("2006-01-02 15:04"))); err != nil {
		// stash fails if nothing to stash — that's fine
	}

	if err := runIn(repoDir, "git", "pull", "--rebase", "origin", "main"); err != nil {
		return fmt.Errorf("pull: %w", err)
	}

	runIn(repoDir, "git", "stash", "drop")

	for name, tool := range cfg.Tools {
		if !tool.Enabled {
			continue
		}
		if err := syncTool(repoDir, name, tool); err != nil {
			return fmt.Errorf("sync %s: %w", name, err)
		}
	}

	if err := runIn(repoDir, "git", "add", "-A"); err != nil {
		return err
	}

	out, _ := exec.Command("git", "-C", repoDir, "diff", "--cached", "--quiet").CombinedOutput()
	if exec.Command("git", "-C", repoDir, "diff", "--cached", "--quiet").Run() == nil {
		fmt.Println("  No changes to sync")
		return nil
	}
	_ = out

	msg := fmt.Sprintf("skill-vault sync %s", time.Now().Format("2006-01-02 15:04"))
	if err := runIn(repoDir, "git", "commit", "-m", msg); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	if err := runIn(repoDir, "git", "push", "origin", "main"); err != nil {
		return fmt.Errorf("push: %w", err)
	}

	fmt.Println("  Pushed to GitHub")
	return nil
}

func syncTool(repoDir, name string, tool config.ToolConfig) error {
	var toolDef *detect.Tool
	for _, t := range detect.KnownTools {
		if t.Name == name {
			toolDef = &t
			break
		}
	}
	if toolDef == nil {
		return nil
	}

	enabledCategories := make(map[string]bool)
	for _, c := range tool.Categories {
		enabledCategories[c] = true
	}

	for _, bp := range toolDef.Paths {
		if !enabledCategories[string(bp.Category)] {
			continue
		}

		src := expandHome(bp.Path)
		if _, err := os.Stat(src); os.IsNotExist(err) {
			continue
		}

		destDir := filepath.Join(repoDir, name, string(bp.Category))
		os.MkdirAll(destDir, 0755)

		if bp.Pattern != "" {
			if err := copyGlob(src, bp.Pattern, destDir); err != nil {
				return err
			}
		} else {
			info, _ := os.Stat(src)
			if info != nil && info.IsDir() {
				dest := filepath.Join(repoDir, name, string(bp.Category), filepath.Base(src))
				if err := run("rsync", "-a", "--delete", src+"/", dest+"/"); err != nil {
					return err
				}
			} else {
				dest := filepath.Join(destDir, filepath.Base(src))
				if err := copyFile(src, dest); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func syncS3(cfg *config.Config) error {
	for name, tool := range cfg.Tools {
		if !tool.Enabled {
			continue
		}

		hasConversations := false
		for _, c := range tool.Categories {
			if c == "conversations" {
				hasConversations = true
				break
			}
		}
		if !hasConversations {
			continue
		}

		var toolDef *detect.Tool
		for _, t := range detect.KnownTools {
			if t.Name == name {
				toolDef = &t
				break
			}
		}
		if toolDef == nil {
			continue
		}

		for _, bp := range toolDef.Paths {
			if bp.Category != detect.CategoryConversations {
				continue
			}

			src := expandHome(bp.Path)
			if _, err := os.Stat(src); os.IsNotExist(err) {
				continue
			}

			archive := filepath.Join(os.TempDir(), fmt.Sprintf("%s-conversations-%s.tar.gz",
				name, time.Now().Format("20060102")))

			fmt.Printf("  Compressing %s conversations...\n", toolDef.Description)
			if err := run("tar", "czf", archive, "-C", filepath.Dir(src), filepath.Base(src)); err != nil {
				return fmt.Errorf("tar %s: %w", name, err)
			}

			s3Key := fmt.Sprintf("%s-conversations-%s.tar.gz", name, time.Now().Format("20060102"))
			fmt.Printf("  Uploading to s3://%s/%s...\n", cfg.S3.Bucket, s3Key)

			args := []string{"s3", "cp", archive,
				fmt.Sprintf("s3://%s/%s", cfg.S3.Bucket, s3Key),
				"--profile", cfg.S3.Profile, "--quiet"}
			if cfg.S3.Region != "" {
				args = append(args, "--region", cfg.S3.Region)
			}
			if err := run("aws", args...); err != nil {
				os.Remove(archive)
				return fmt.Errorf("s3 upload %s: %w", name, err)
			}

			os.Remove(archive)
			fmt.Printf("  Uploaded %s\n", s3Key)
		}
	}
	return nil
}

func copyGlob(root, pattern, destDir string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		matched, _ := filepath.Match(pattern, rel)
		if !matched {
			parts := strings.Split(rel, string(filepath.Separator))
			if len(parts) > 1 {
				matched, _ = filepath.Match(pattern, filepath.Join(parts[len(parts)-2], parts[len(parts)-1]))
			}
		}
		if !matched {
			for _, part := range strings.Split(pattern, "/") {
				if m, _ := filepath.Match(part, filepath.Base(path)); m {
					matched = true
					break
				}
			}
		}
		if matched {
			rel, _ := filepath.Rel(root, path)
			dest := filepath.Join(destDir, rel)
			os.MkdirAll(filepath.Dir(dest), 0755)
			return copyFile(path, dest)
		}
		return nil
	})
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runIn(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
