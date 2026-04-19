# skill-vault

Back up your AI agent skills, config, and conversation logs. Detects installed tools automatically.

## Supported Tools

| Tool | Skills/Rules | Config | Memory | Conversations |
|------|-------------|--------|--------|---------------|
| Claude Code | `skills/`, `agents/`, `commands/` | `settings.json` | `projects/*/memory/` | `projects/*.jsonl` |
| Cursor | `rules/` | `settings.json`, `mcp.json` | | |
| Codex | | `config.yaml`, `instructions.md` | | |
| Windsurf | `rules/` | `settings.json` | `memories/` | |
| Aider | | `.aider.conf.yml` | | `chat-history/` |
| Continue | `rules/` | `config.json`, `config.ts`, `config.yaml` | | |
| Copilot | | config dir | | |
| Amp | | `config.yaml` | | `threads/` |
| Cline | `rules/` | `config.json` | | `tasks/` |
| Roo Code | `rules/` | `config.json` | | `tasks/` |
| Tabnine | | `config/` | | |
| Supermaven | | config dir | | |
| Zed AI | `rules/` | `settings.json`, `keymap.json` | | `conversations/` |
| Warp AI | | `config.yaml` | | `sessions/` |
| Amazon Q | | config dir | | |
| Gemini CLI | | `settings.json` | | `history/` |
| Claude Dev | | `config.json` | | `tasks/` |

## Backup Targets

| Target | What it backs up | Requires |
|--------|-----------------|----------|
| **Git** (GitHub/GitLab) | Skills, config, memory, rules | `git` |
| **AWS S3** | Conversation logs (compressed) | `aws` CLI |
| **Google Cloud Storage** | Conversation logs (compressed) | `gcloud` CLI |
| **Azure Blob Storage** | Conversation logs (compressed) | `az` CLI |
| **iCloud Drive** | Conversation logs (compressed) | macOS |
| **Time Machine** | Verifies tool dirs are included | macOS |

## Install

```bash
go install github.com/lorenjphillips/skill-vault@latest
```

## Usage

### Setup

```bash
skill-vault init
```

Interactive setup that:
1. Scans for installed AI tools
2. Lets you select which tools and categories to back up
3. Configures backup targets (git, cloud storage, iCloud, Time Machine)
4. Sets up a macOS launchd job for automatic backups

### Manual Sync

```bash
skill-vault sync
```

### Check Status

```bash
skill-vault status
```

## Config

Stored at `~/.skill-vault/config.yaml`:

```yaml
tools:
  claude:
    enabled: true
    categories: [skills, config, memory, conversations]
  cursor:
    enabled: true
    categories: [rules, config]
git:
  enabled: true
  provider: github
  repo: git@github.com:you/ai-backup.git
  local_path: ~/Development/ai-backup
s3:
  enabled: true
  bucket: my-ai-backups
  profile: default
  region: us-east-1
schedule:
  enabled: true
  interval: 24h
```

## Requirements

- macOS (launchd scheduling)
- `git` (for git backup)
- `rsync` (for file sync)
- Cloud CLI tools only if using that target (`aws`, `gcloud`, `az`)

## License

MIT
