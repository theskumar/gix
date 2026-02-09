package changelog

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const changelogFile = "CHANGELOG.md"

const keepachangelogHeader = `# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
`

// FormatVersionHeader returns a keepachangelog version header like "## [v1.0.0] - 2026-02-09".
func FormatVersionHeader(version string) string {
	date := time.Now().Format("2006-01-02")
	return fmt.Sprintf("## [%s] - %s", version, date)
}

// FormatCommitLog formats commit subjects as a bullet list for AI input.
func FormatCommitLog(subjects []string) string {
	var b strings.Builder
	for _, s := range subjects {
		b.WriteString("- ")
		b.WriteString(s)
		b.WriteString("\n")
	}
	return b.String()
}

// ExtractUnreleased parses an existing CHANGELOG.md and separates the
// [Unreleased] section content from the rest. Returns the body text between
// ## [Unreleased] and the next ## [ header (or EOF), and the rest of the file.
func ExtractUnreleased(content string) (unreleased string, rest string) {
	lines := strings.Split(content, "\n")

	unreleasedStart := -1
	unreleasedEnd := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(strings.ToLower(line))
		if unreleasedStart == -1 {
			if strings.HasPrefix(trimmed, "## [unreleased]") {
				unreleasedStart = i
				continue
			}
		} else if unreleasedEnd == -1 {
			// Look for the next versioned section
			if strings.HasPrefix(trimmed, "## [") {
				unreleasedEnd = i
				break
			}
		}
	}

	if unreleasedStart == -1 {
		// No [Unreleased] section found
		return "", content
	}

	if unreleasedEnd == -1 {
		// [Unreleased] goes to end of file
		unreleasedEnd = len(lines)
	}

	// Body is everything between the [Unreleased] header and the next section
	body := strings.TrimSpace(strings.Join(lines[unreleasedStart+1:unreleasedEnd], "\n"))

	// Rest is everything before [Unreleased] header + everything from next section onward
	var restParts []string
	restParts = append(restParts, lines[:unreleasedStart]...)
	if unreleasedEnd < len(lines) {
		restParts = append(restParts, lines[unreleasedEnd:]...)
	}
	rest = strings.Join(restParts, "\n")

	return body, rest
}

// WriteChangelog writes the new version section to CHANGELOG.md, preserving existing content.
func WriteChangelog(newSection, existingContent string) error {
	var result string

	if existingContent == "" {
		// No file exists — create with keepachangelog header
		result = keepachangelogHeader + "\n## [Unreleased]\n\n" + newSection + "\n"
	} else if hasUnreleasedSection(existingContent) {
		// File has ## [Unreleased] — clear its body and insert new section after it
		result = replaceUnreleasedAndInsert(existingContent, newSection)
	} else {
		// File exists without ## [Unreleased] — insert before first ## [
		result = insertBeforeFirstVersion(existingContent, newSection)
	}

	return os.WriteFile(changelogFile, []byte(result), 0644)
}

func hasUnreleasedSection(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(strings.ToLower(line)), "## [unreleased]") {
			return true
		}
	}
	return false
}

func replaceUnreleasedAndInsert(content, newSection string) string {
	lines := strings.Split(content, "\n")
	var result strings.Builder

	unreleasedFound := false
	bodySkipped := false

	for i := 0; i < len(lines); i++ {
		trimmed := strings.TrimSpace(strings.ToLower(lines[i]))

		if !unreleasedFound && strings.HasPrefix(trimmed, "## [unreleased]") {
			// Write the Unreleased header with empty body
			result.WriteString(lines[i])
			result.WriteString("\n\n")
			// Insert new versioned section
			result.WriteString(newSection)
			result.WriteString("\n\n")
			unreleasedFound = true
			continue
		}

		if unreleasedFound && !bodySkipped {
			// Skip lines until we hit the next ## [ section
			if strings.HasPrefix(trimmed, "## [") {
				bodySkipped = true
				result.WriteString(lines[i])
				result.WriteString("\n")
				continue
			}
			// Skip the old unreleased body
			continue
		}

		result.WriteString(lines[i])
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

func insertBeforeFirstVersion(content, newSection string) string {
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(strings.ToLower(line))
		if strings.HasPrefix(trimmed, "## [") {
			// Insert new section before this line
			before := strings.Join(lines[:i], "\n")
			after := strings.Join(lines[i:], "\n")
			return before + newSection + "\n\n" + after
		}
	}

	// No versioned sections found — append after existing content
	return strings.TrimRight(content, "\n") + "\n\n" + newSection + "\n"
}
