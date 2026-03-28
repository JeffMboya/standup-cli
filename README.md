# standup

A CLI tool that scans your local git repositories and generates a standup summary of what you worked on — grouped by repo, ready to paste into Slack, email, or a standup bot.

```
$ standup

Your standup — last 24h
───────────────────────

my-api  (3 commits)
  • fix: user auth token expiry bug
  • feat: add rate limiting middleware
  • chore: update deps

frontend  (1 commit)
  • feat: dark mode toggle

───────────────────────
2 repos, 4 commits total
✓ Copied to clipboard
```

No config needed to get started. Point it at your code directory once and it finds your repos, filters to your commits, and copies the result to clipboard.

---

## Why

Writing a standup update every morning means mentally reconstructing what you actually did yesterday across however many repos you touched. Most people either skip it, write something vague, or spend more time on the update than it deserves.

`standup` does the reconstruction for you. Run it before your morning standup and paste.

**Use cases:**

- **Daily standups** — "What did you do yesterday?" answered in one command.
- **Weekly recaps** — Use `--since week` for a week-in-review for 1:1s, Friday team updates, or status emails.
- **Async teams** — Post your update to Slack automatically before anyone asks.
- **Freelancers/contractors** — Track billable work across multiple client repos without keeping a manual log.
- **Performance reviews** — `--since "3 months ago"` to remind yourself what you shipped.
- **Context switching** — Coming back after a few days away? `standup --since 5d` to re-orient before touching any code.
- **Team leads** — Set `author` to a teammate's email to see what they've been working on.

---

## Install

### Option 1 — Download a binary (no Go required)

Go to the [Releases page](https://github.com/JeffMboya/standup-cli/releases/latest) and download the binary for your platform:

| Platform | File |
|----------|------|
| macOS (Apple Silicon) | `standup-darwin-arm64` |
| macOS (Intel) | `standup-darwin-amd64` |
| Linux x86_64 | `standup-linux-amd64` |
| Linux ARM64 | `standup-linux-arm64` |
| Windows | `standup-windows-amd64.exe` |

**macOS / Linux:**
```bash
# Replace the filename with your platform's binary
curl -L https://github.com/JeffMboya/standup-cli/releases/latest/download/standup-darwin-arm64 -o standup
chmod +x standup
sudo mv standup /usr/local/bin/
```

**Windows:** Download the `.exe` and put it somewhere on your `PATH`.

### Option 2 — Install with Go

```bash
go install github.com/JeffMboya/standup-cli@latest
```

### Option 3 — Build from source

```bash
git clone https://github.com/JeffMboya/standup-cli
cd standup-cli
go build -o standup .
mv standup ~/.local/bin/
```

### Linux clipboard support

The tool auto-copies output to your clipboard. Install one of:
```bash
sudo apt install xclip         # X11 (most common)
sudo apt install xsel          # X11 alternative
sudo apt install wl-clipboard  # Wayland
```

macOS uses `pbcopy` (built-in). Windows uses `clip` (built-in). If none is found, the tool still works — it just skips the clipboard step.

---

## Usage

```bash
standup                       # last 24 hours
standup --since week          # last 7 days
standup --since yesterday     # since yesterday
standup --since monday        # since last Monday
standup --since 3d            # last 3 days
standup --since 2w            # last 2 weeks
standup --since "2024-01-15"  # since a specific date

standup --slack               # also post to Slack (requires config)
standup --no-clip             # print only, skip clipboard
standup --no-slack            # skip Slack even if webhook is configured
standup --dir ~/clients/acme  # scan a specific directory
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--since` | `24h` | Time window (see formats below) |
| `--week` | — | Alias for `--since=week` |
| `--dir` | auto | Directory to search (overrides config) |
| `--slack` | — | Post to Slack webhook from config |
| `--no-slack` | — | Skip Slack post even if webhook is configured |
| `--no-clip` | — | Skip clipboard copy |

### `--since` formats

| Input | Means |
|-------|-------|
| `24h` | Last 24 hours (default) |
| `week` | Last 7 days |
| `yesterday` | Since yesterday |
| `today` | Since midnight |
| `monday` … `sunday` | Since last occurrence of that day |
| `2d`, `5d` | Last N days |
| `1w`, `3w` | Last N weeks |
| `12h`, `48h` | Last N hours |
| `"2024-01-15"` | Since a specific date |
| `"3 months ago"` | Any git-compatible relative date |

---

## Config file

Create `~/.standup.yaml` to set your preferences once:

```yaml
# Directories to scan for git repos (~ is expanded)
dirs:
  - ~/code
  - ~/clients/acme

# Slack incoming webhook URL — set this to enable Slack posting
# Get one at: https://api.slack.com/messaging/webhooks
slack_webhook: https://hooks.slack.com/services/T.../B.../...

# Default time window (same formats as --since)
default_since: 24h

# Set to true to skip clipboard by default
no_clip: false

# Set to true to skip Slack posting even if webhook is set
no_slack: false

# Override author detection (leave blank to use git config user.email)
author: ""
```

With `slack_webhook` set, `standup` will post to Slack automatically on every run. Use `--no-slack` to suppress it or add `no_slack: true` to config to opt out and only post manually with `--slack`.

### Auto-post via cron

Post your standup to Slack every weekday morning at 9am:

```bash
# Add to crontab: crontab -e
0 9 * * 1-5 /home/yourname/.local/bin/standup --no-clip
```

---

## How it works

1. **Finds your repos** — walks your code directory up to 6 levels deep. Skips `node_modules`, `vendor`, `.venv`, `build`, `dist`, and other non-code directories.

2. **Filters to your commits** — reads your identity from `git config --global user.email` (falls back to `user.name`). Override with `author` in config.

3. **Formats the output** — groups commits by repo, strips merge commits, strips the short hash prefix so you just see the message.

4. **Copies to clipboard** — uses `xclip`, `xsel`, or `wl-copy` on Linux; `pbcopy` on macOS; `clip` on Windows.

5. **Posts to Slack** — if `slack_webhook` is configured, sends a Slack-formatted version to your channel automatically.

---

## Requirements

- git in your PATH
- `git config --global user.email` (or `user.name`) set — used to filter commits to yours
- A clipboard tool on Linux (xclip, xsel, or wl-copy) — optional, only needed for clipboard support

---

## License

MIT
