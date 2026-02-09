package git

import (
	"bytes"
	"os/exec"
	"strings"
)

// Commit represents a single git commit with its hash and subject line.
type Commit struct {
	Hash    string
	Subject string
}

// GetLatestSemverTag returns the most recent semver tag (e.g. "v1.2.3").
// Returns "" (not an error) if no tags exist.
func GetLatestSemverTag() (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0", "--match", "v[0-9]*")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()

	if err != nil {
		// No tags found — not an error, just means fresh repo
		return "", nil
	}

	return strings.TrimSpace(stdout.String()), nil
}

// GetCommitsSince returns commits since the given ref (exclusive) up to HEAD.
// If sinceRef is empty, returns all commits.
func GetCommitsSince(sinceRef string) ([]Commit, error) {
	args := []string{"log", "--no-merges", "--format=%H %s"}
	if sinceRef != "" {
		args = append(args, sinceRef+"..HEAD")
	}

	cmd := exec.Command("git", args...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()

	if err != nil {
		return nil, err
	}

	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return nil, nil
	}

	lines := strings.Split(output, "\n")
	commits := make([]Commit, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format is "<hash> <subject>" — split on first space
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}
		commits = append(commits, Commit{
			Hash:    parts[0],
			Subject: parts[1],
		})
	}

	return commits, nil
}
