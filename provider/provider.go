package provider

const (
	// CommitMessageSystemPrompt is the system instruction for generating commit messages
	CommitMessageSystemPrompt = `Generate conventional commit messages: <type>[scope]: <description>
Types: feat, fix, docs, style, refactor, perf, test, build, ci, chore. Use ! for breaking changes (e.g., feat!:). Keep first line under 72 chars.`

	// CommitMessageUserPromptTemplate is the template for the user prompt when generating commit messages
	CommitMessageUserPromptTemplate = `Generate a conventional commit message for this diff (imperative mood, add body/footers if helpful):

`
)

// AIProvider abstracts chat completion and embedding capabilities
// so that different backends (OpenAI, Gemini, etc.) can be used interchangeably.
type AIProvider interface {
	GenerateCommitMessage(diff string) (string, error)
	GetEmbeddings(texts []string) ([][]float32, error)
}
