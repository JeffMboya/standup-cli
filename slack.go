package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// formatSlack produces Slack mrkdwn-formatted standup text.
func formatSlack(activities []RepoActivity, label string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "*Standup — %s*\n\n", label)

	totalCommits := 0
	for _, a := range activities {
		n := len(a.Commits)
		totalCommits += n
		fmt.Fprintf(&b, "*%s*  (%d %s)\n", a.Name, n, pluralise("commit", n))
		for _, c := range a.Commits {
			fmt.Fprintf(&b, "  • %s\n", c)
		}
		b.WriteString("\n")
	}

	fmt.Fprintf(&b, "_%d %s, %d %s total_",
		len(activities), pluralise("repo", len(activities)),
		totalCommits, pluralise("commit", totalCommits),
	)

	return b.String()
}

// postToSlack sends a message to a Slack incoming webhook URL.
func postToSlack(webhookURL, text string) error {
	payload, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return err
	}
	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("posting to Slack: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack returned HTTP %d", resp.StatusCode)
	}
	return nil
}
