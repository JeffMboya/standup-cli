package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"unicode/utf8"
)

const maxDepth = 6

type RepoActivity struct {
	Name    string
	Commits []string
}

func main() {
	sinceFlag := flag.String("since", "", "Time window: 24h (default), week, yesterday, monday, 2d, 3w, etc.")
	weekFlag := flag.Bool("week", false, "Alias for --since=week")
	dirFlag := flag.String("dir", "", "Directory to search for git repos (overrides config)")
	slackFlag := flag.Bool("slack", false, "Post to Slack (uses slack_webhook from ~/.standup.yaml)")
	noSlackFlag := flag.Bool("no-slack", false, "Skip Slack post even if webhook is configured")
	noClipFlag := flag.Bool("no-clip", false, "Don't copy output to clipboard")
	flag.Parse()

	cfg := loadConfig()

	// Resolve time window: --week alias, then --since flag, then config default, then 24h
	sinceStr := cfg.DefaultSince
	if *weekFlag {
		sinceStr = "week"
	}
	if *sinceFlag != "" {
		sinceStr = *sinceFlag
	}
	if sinceStr == "" {
		sinceStr = "24h"
	}
	since := parseSince(sinceStr)

	// Resolve search directories: --dir flag overrides config
	var searchDirs []string
	if *dirFlag != "" {
		searchDirs = []string{*dirFlag}
	} else if len(cfg.Dirs) > 0 {
		searchDirs = cfg.Dirs
	} else {
		searchDirs = []string{detectCodeDir()}
	}

	// Resolve author: config overrides auto-detect
	author := cfg.Author
	if author == "" {
		author = getGitAuthor()
	}
	if author == "" {
		fmt.Fprintln(os.Stderr, "warning: could not detect git author — showing all commits")
	}

	// Find all repos across all search dirs
	var allRepos []string
	for _, dir := range searchDirs {
		fmt.Fprintf(os.Stderr, "Scanning %s...\n", dir)
		allRepos = append(allRepos, findGitRepos(dir, maxDepth)...)
	}
	if len(allRepos) == 0 {
		fmt.Fprintln(os.Stderr, "No git repos found.")
		fmt.Fprintln(os.Stderr, "Set dirs in ~/.standup.yaml or use: standup --dir ~/path/to/code")
		os.Exit(1)
	}

	// Collect activity
	var activities []RepoActivity
	for _, repo := range allRepos {
		commits := getCommits(repo, author, since.gitArg)
		if len(commits) > 0 {
			activities = append(activities, RepoActivity{
				Name:    filepath.Base(repo),
				Commits: commits,
			})
		}
	}

	if len(activities) == 0 {
		fmt.Printf("Nothing to report for %s.\n", since.label)
		return
	}

	// Print plain output
	output := formatOutput(activities, since.label)
	fmt.Print(output)

	// Clipboard
	noClip := *noClipFlag || cfg.NoClip
	if !noClip {
		if err := copyToClipboard(output); err != nil {
			fmt.Fprintf(os.Stderr, "clipboard: %v\n", err)
		} else {
			fmt.Fprintln(os.Stderr, "✓ Copied to clipboard")
		}
	}

	// Slack: post if webhook is set and not suppressed
	webhook := cfg.SlackWebhook
	shouldPost := webhook != "" && !*noSlackFlag && !cfg.NoSlack
	if *slackFlag && webhook == "" {
		fmt.Fprintln(os.Stderr, "error: --slack requires slack_webhook in ~/.standup.yaml")
		os.Exit(1)
	}
	if shouldPost || *slackFlag {
		slackText := formatSlack(activities, since.label)
		if err := postToSlack(webhook, slackText); err != nil {
			fmt.Fprintf(os.Stderr, "slack: %v\n", err)
		} else {
			fmt.Fprintln(os.Stderr, "✓ Posted to Slack")
		}
	}
}

// detectCodeDir looks for common developer code directories under home.
func detectCodeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	for _, candidate := range []string{"code", "projects", "dev", "work", "src", "repos"} {
		p := filepath.Join(home, candidate)
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			return p
		}
	}
	return home
}

// getGitAuthor returns the global git user email (or name as fallback).
func getGitAuthor() string {
	for _, field := range []string{"user.email", "user.name"} {
		out, err := exec.Command("git", "config", "--global", field).Output()
		if err == nil {
			if val := strings.TrimSpace(string(out)); val != "" {
				return val
			}
		}
	}
	return ""
}

// findGitRepos walks root up to maxD levels deep looking for .git directories.
func findGitRepos(root string, maxD int) []string {
	var repos []string

	skipDirs := map[string]bool{
		"node_modules": true,
		"vendor":       true,
		".cache":       true,
		"dist":         true,
		"build":        true,
		"target":       true,
		".npm":         true,
		".cargo":       true,
		"__pycache__":  true,
		".venv":        true,
		"venv":         true,
	}

	var walk func(dir string, depth int)
	walk = func(dir string, depth int) {
		if depth > maxD {
			return
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			name := e.Name()
			if name == ".git" {
				repos = append(repos, dir)
				return // don't recurse into repo subdirectories
			}
			if strings.HasPrefix(name, ".") || skipDirs[name] {
				continue
			}
			walk(filepath.Join(dir, name), depth+1)
		}
	}
	walk(root, 0)
	return repos
}

// getCommits returns commit messages from a repo since the given time, filtered by author.
func getCommits(repoPath, author, since string) []string {
	args := []string{"log", "--oneline", "--since=" + since, "--no-merges"}
	if author != "" {
		args = append(args, "--author="+author)
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil
	}

	var commits []string
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if _, msg, found := strings.Cut(line, " "); found {
			commits = append(commits, msg)
		} else {
			commits = append(commits, line)
		}
	}
	return commits
}

// formatOutput produces the plain-text standup output.
func formatOutput(activities []RepoActivity, label string) string {
	var b strings.Builder

	header := fmt.Sprintf("Your standup — %s", label)
	rule := strings.Repeat("─", utf8.RuneCountInString(header))

	b.WriteString(header + "\n")
	b.WriteString(rule + "\n\n")

	totalCommits := 0
	for _, a := range activities {
		n := len(a.Commits)
		totalCommits += n
		fmt.Fprintf(&b, "%s  (%d %s)\n", a.Name, n, pluralise("commit", n))
		for _, c := range a.Commits {
			fmt.Fprintf(&b, "  • %s\n", c)
		}
		b.WriteString("\n")
	}

	b.WriteString(rule + "\n")
	fmt.Fprintf(&b, "%d %s, %d %s total\n",
		len(activities), pluralise("repo", len(activities)),
		totalCommits, pluralise("commit", totalCommits),
	)

	return b.String()
}

// copyToClipboard writes text to the system clipboard.
func copyToClipboard(text string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "windows":
		cmd = exec.Command("clip")
	default:
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else if _, err := exec.LookPath("wl-copy"); err == nil {
			cmd = exec.Command("wl-copy")
		} else {
			return fmt.Errorf("no clipboard tool found — install xclip, xsel, or wl-copy")
		}
	}
	cmd.Stdin = bytes.NewBufferString(text)
	return cmd.Run()
}
