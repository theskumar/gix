package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func IsGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	err := cmd.Run()
	return err == nil
}

func HasStagedChanges() (bool, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	out, err := cmd.Output()

	if err != nil {
		return false, err
	}

	return strings.TrimSpace(string(out)) != "", nil
}

func GetStagedDiff() (string, error) {
	var stdout bytes.Buffer
	cmd := exec.Command("git", "diff", "--cached", "--unified=3")
	cmd.Stdout = &stdout
	err := cmd.Run()

	if err != nil {
		return "", err
	}

	diff := strings.TrimSpace(stdout.String())

	if diff == "" {
		return "", fmt.Errorf("no staged changes to commit.\nUse `git add <files>` to stage changes first")
	}

	return diff, nil
}

// GetCompactStagedDiff returns a token-optimized version of the staged diff
// by reducing context lines and filtering noise
func GetCompactStagedDiff() (string, error) {
	var stdout bytes.Buffer
	// Use --unified=0 to reduce context lines (saves 20-40% tokens)
	cmd := exec.Command("git", "diff", "--cached", "--unified=0")
	cmd.Stdout = &stdout
	err := cmd.Run()

	if err != nil {
		return "", err
	}

	diff := strings.TrimSpace(stdout.String())

	if diff == "" {
		return "", fmt.Errorf("no staged changes to commit.\nUse `git add <files>` to stage changes first")
	}

	// Apply additional compression filters
	return CompressDiff(diff), nil
}

// CompressDiff reduces token usage by filtering noise and compressing verbose content
func CompressDiff(diff string) string {
	var result strings.Builder
	lines := strings.Split(diff, "\n")

	inBinaryFile := false
	inLockFile := false
	currentFile := ""
	const maxLineLength = 200

	for _, line := range lines {
		// Track current file being processed
		if strings.HasPrefix(line, "diff --git") {
			inBinaryFile = false
			inLockFile = false
			currentFile = extractFilePath(line)

			// Check if this is a lock file or other noise file
			if isLockFile(currentFile) {
				inLockFile = true
				result.WriteString(line + "\n")
				result.WriteString("--- Dependencies updated ---\n")
				continue
			}

			result.WriteString(line + "\n")
			continue
		}

		// Handle binary files
		if strings.Contains(line, "Binary files") {
			inBinaryFile = true
			result.WriteString(fmt.Sprintf("Binary file changed: %s\n", currentFile))
			continue
		}

		// Skip content of lock files and binary files
		if inLockFile || inBinaryFile {
			continue
		}

		// Truncate very long lines (e.g., minified code, long strings)
		if len(line) > maxLineLength && !strings.HasPrefix(line, "diff --git") &&
		   !strings.HasPrefix(line, "---") && !strings.HasPrefix(line, "+++") {
			truncated := line[:maxLineLength] + "... [truncated]"
			result.WriteString(truncated + "\n")
			continue
		}

		result.WriteString(line + "\n")
	}

	return strings.TrimSpace(result.String())
}

// extractFilePath extracts the file path from a "diff --git a/path b/path" line
func extractFilePath(line string) string {
	parts := strings.Fields(line)
	if len(parts) >= 3 {
		// Remove "a/" prefix
		path := parts[2]
		if strings.HasPrefix(path, "a/") {
			return path[2:]
		}
		return path
	}
	return ""
}

// isLockFile checks if a file is a dependency lock file or similar noise
func isLockFile(path string) bool {
	lockFiles := []string{
		"package-lock.json",
		"yarn.lock",
		"pnpm-lock.yaml",
		"Gemfile.lock",
		"Cargo.lock",
		"go.sum",
		"poetry.lock",
		"Pipfile.lock",
		"composer.lock",
		"packages.lock.json",
	}

	for _, lockFile := range lockFiles {
		if strings.HasSuffix(path, lockFile) {
			return true
		}
	}

	return false
}
