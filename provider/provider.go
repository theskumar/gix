package provider

const (
	// CommitMessageSystemPrompt is the system instruction for generating commit messages
	CommitMessageSystemPrompt = `Generate conventional commit messages: <type>[scope]: <description>
Types: feat, fix, docs, style, refactor, perf, test, build, ci, chore. Use ! for breaking changes (e.g., feat!:). Keep first line under 72 chars.`

	// CommitMessageUserPromptTemplate is the template for the user prompt when generating commit messages
	CommitMessageUserPromptTemplate = `Generate a conventional commit message for this diff (imperative mood, add body/footers if helpful):

`

	// ChangelogSystemPrompt instructs the AI to produce keepachangelog-formatted entries.
	ChangelogSystemPrompt = `You generate changelog entries following the Keep a Changelog format (https://keepachangelog.com/en/1.1.0/).

Rules:
- Group changes under these categories ONLY: Added, Changed, Deprecated, Removed, Fixed, Security
- Omit categories that have no entries
- Write entries from the user's perspective — describe what changed, not how
- Combine related commits into single meaningful entries
- Skip internal-only changes (refactors, CI tweaks) unless they affect users
- Do NOT include commit hashes
- Do NOT include a version header (## [version] - date) — just the categorized entries
- Do NOT wrap output in code fences
- Keep entries concise but descriptive
- When existing curated entries are provided, preserve them as-is and only add entries for changes not already covered`

	// ChangelogUserPromptTemplate is the template when generating from commits only.
	ChangelogUserPromptTemplate = "Generate a changelog from these commits:\n\n"

	// ChangelogUserPromptWithExistingTemplate is used when curated unreleased entries exist.
	ChangelogUserPromptWithExistingTemplate = "The following changelog entries have already been curated by the maintainer. Preserve them exactly and add any missing entries based on the commits below.\n\nExisting entries:\n%s\n\nCommits:\n%s"
)

// AIProvider abstracts chat completion and embedding capabilities
// so that different backends (OpenAI, Gemini, etc.) can be used interchangeably.
type AIProvider interface {
	GenerateCommitMessage(diff string) (string, error)
	GenerateChangelog(commitLog string) (string, error)
	GetEmbeddings(texts []string) ([][]float32, error)
}
