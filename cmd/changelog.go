package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ademajagon/gix/changelog"
	"github.com/ademajagon/gix/config"
	"github.com/ademajagon/gix/git"
	"github.com/ademajagon/gix/provider"
	"github.com/ademajagon/gix/utils"
	"github.com/spf13/cobra"
)

var (
	changelogFrom string
	changelogAll  bool
)

var changelogCmd = &cobra.Command{
	Use:   "changelog",
	Short: "Generate a changelog entry using AI",
	Run: func(cmd *cobra.Command, args []string) {
		if !git.IsGitRepo() {
			fmt.Fprintln(os.Stderr, "fatal: not a git repository")
			os.Exit(1)
		}

		if changelogAll && cmd.Flags().Changed("from") {
			fmt.Fprintln(os.Stderr, "error: --from and --all are mutually exclusive")
			os.Exit(1)
		}

		// Determine the base tag
		var tag string
		if changelogAll {
			tag = ""
			fmt.Println("Generating changelog from all commits")
		} else if cmd.Flags().Changed("from") {
			tag = changelogFrom
			fmt.Printf("From tag:   %s\n", tag)
		} else {
			// Auto-detect latest tag
			var err error
			tag, err = git.GetLatestSemverTag()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}

			if tag == "" {
				fmt.Println("No previous tags found")
			} else {
				fmt.Printf("Latest tag: %s\n", tag)
			}
		}

		// Parse base version
		baseTag := tag
		if baseTag == "" {
			baseTag = "v0.0.0"
		}
		baseVersion, err := changelog.ParseVersion(baseTag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		// Get commits since tag
		commits, err := git.GetCommitsSince(tag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		if len(commits) == 0 {
			if tag != "" {
				fmt.Printf("no new commits since %s\n", tag)
			} else {
				fmt.Println("no commits found")
			}
			os.Exit(0)
		}

		// Determine bump and next version
		subjects := make([]string, len(commits))
		for i, c := range commits {
			subjects[i] = c.Subject
		}

		bump := changelog.DetermineBump(subjects)
		nextVersion := baseVersion.Bump(bump)

		fmt.Printf("Commits:    %d\n", len(commits))
		fmt.Printf("Next:       %s (%s)\n", nextVersion, changelog.BumpLabel(bump))

		// Read existing CHANGELOG.md for unreleased entries
		existingContent := ""
		if data, err := os.ReadFile("CHANGELOG.md"); err == nil {
			existingContent = string(data)
		}

		unreleasedEntries, _ := changelog.ExtractUnreleased(existingContent)
		if strings.TrimSpace(unreleasedEntries) != "" {
			fmt.Println("Found unreleased entries in CHANGELOG.md — incorporating them")
		}

		// Load config and create AI provider
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: config not loaded: %v\n", err)
			fmt.Fprintln(os.Stderr, "hint: run `gix config set-key` to set your API key")
			os.Exit(1)
		}

		p, err := provider.New(cfg.ResolveProvider(), cfg.APIKey())
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		// Build the prompt for AI
		commitLog := changelog.FormatCommitLog(subjects)
		aiPrompt := buildChangelogPrompt(commitLog, unreleasedEntries)

		// Generate changelog with spinner
		spinner := utils.NewSpinner()
		spinner.Start()
		body, err := p.GenerateChangelog(aiPrompt)
		spinner.Stop()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: AI provider failed: %v\n", err)
			os.Exit(1)
		}

		// Compose full section
		header := changelog.FormatVersionHeader(nextVersion.String())
		section := header + "\n\n" + strings.TrimSpace(body)

		// Interactive prompt loop
		section, err = promptChangelog(section, aiPrompt, p, header)
		if err != nil {
			os.Exit(0)
		}

		// Write to CHANGELOG.md
		if err := changelog.WriteChangelog(section, existingContent); err != nil {
			fmt.Fprintf(os.Stderr, "error: failed to write CHANGELOG.md: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✓ CHANGELOG.md updated")
	},
}

func init() {
	changelogCmd.Flags().StringVar(&changelogFrom, "from", "", "Base tag to generate changelog from (e.g., v0.1.0)")
	changelogCmd.Flags().BoolVar(&changelogAll, "all", false, "Generate changelog from all commits, ignoring tags")
	rootCmd.AddCommand(changelogCmd)
}

func buildChangelogPrompt(commitLog, unreleasedEntries string) string {
	if strings.TrimSpace(unreleasedEntries) != "" {
		return fmt.Sprintf(provider.ChangelogUserPromptWithExistingTemplate, unreleasedEntries, commitLog)
	}
	return provider.ChangelogUserPromptTemplate + commitLog
}

func promptChangelog(section, aiPrompt string, p provider.AIProvider, header string) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	displayChangelog(section)

	for {
		fmt.Print("[Enter] to save  [e]dit  [r]egen  [c]ancel: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "":
			return section, nil
		case "e":
			section = utils.EditInEditor(section)
			displayChangelog(section)
		case "r":
			spinner := utils.NewSpinner()
			spinner.Start()
			body, err := p.GenerateChangelog(aiPrompt)
			spinner.Stop()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: AI provider failed: %v\n", err)
				continue
			}
			section = header + "\n\n" + strings.TrimSpace(body)
			displayChangelog(section)
		case "c":
			fmt.Println("canceled")
			return "", fmt.Errorf("canceled")
		default:
			fmt.Println("invalid input")
		}
	}
}

func displayChangelog(section string) {
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	utils.TypingEffect(section, 3*time.Millisecond)
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
}
