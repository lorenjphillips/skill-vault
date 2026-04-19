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
| Continue | `rules/` | `config.json`, `config.ts` | | |
| Copilot | | config dir | | |
| Amp | | `config.yaml` | | `threads/` |

## Backup Targets

- **GitHub** — skills, config, memory, rules synced to a git repo
- **S3** — conversation logs compressed and uploaded as daily snapshots (too large for git)

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
3. Configures GitHub repo and/or S3 bucket
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
    categories:
      - skills
      - config
      - memory
      - conversations
  cursor:
    enabled: true
    categories:
      - rules
      - config
github:
  enabled: true
  repo: git@github.com:you/ai-backup.git
  local_path: ~/Development/ai-backup
s3:
  enabled: true
  bucket: my-ai-backups
  profile: ai-backup
  region: us-east-2
schedule:
  enabled: true
  interval: 24h
```

## Requirements

- macOS (launchd scheduling)
- `git` (for GitHub backup)
- `aws` CLI (for S3 backup)
- `rsync` (for file sync)

## License

MIT
