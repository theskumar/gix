package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ademajagon/gix/config"
	"github.com/ademajagon/gix/git"
	"github.com/ademajagon/gix/provider"
	"github.com/ademajagon/gix/semantics"
	"github.com/ademajagon/gix/utils"
	"github.com/spf13/cobra"
)

var splitCmd = &cobra.Command{
	Use:   "split",
	Short: "Suggest a split of the current staged diff into multiple semantic commits",
	Run: func(cmd *cobra.Command, args []string) {
		if !git.IsGitRepo() {
			fmt.Fprintf(os.Stderr, "fatal: not a git repository")
			os.Exit(1)
		}

		hasStaged, err := git.HasStagedChanges()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: checking staged changes: %v\n", err)
			os.Exit(1)
		}

		if !hasStaged {
			fmt.Fprintln(os.Stderr, "nothing to split (no staged changes)")
			os.Exit(0)
		}

		hunks, err := git.ParseHunks()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing hunks: %v\n", err)
			os.Exit(1)
		}

		if len(hunks) == 0 {
			fmt.Println("No hunks found in staged diff.")
			return
		}

		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
			os.Exit(1)
		}

		p, err := provider.New(cfg.ResolveProvider(), cfg.APIKey())
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		spinner := utils.NewSpinner()
		spinner.Start()
		groups, err := semantics.ClusterHunks(p, hunks)
		spinner.Stop()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error clustering hunks: %v\n", err)
			os.Exit(1)
		}

		if len(groups) == 0 {
			fmt.Fprintln(os.Stderr, "nothing to split (no semantic groups found)")
			os.Exit(0)
		}

		finalGroups, err := promptCommitGroups(groups, hunks, p)
		if err != nil {
			os.Exit(0)
		}

		if err := semantics.ApplyGroups(finalGroups); err != nil {
			fmt.Fprintf(os.Stderr, "error applying commits: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\n✓ Successfully created %d commits\n", len(finalGroups))
	},
}

func init() {
	rootCmd.AddCommand(splitCmd)
}

func promptCommitGroups(groups []semantics.HunkGroup, hunks []git.Hunk, p provider.AIProvider) ([]semantics.HunkGroup, error) {
	reader := bufio.NewReader(os.Stdin)
	currentGroups := groups

	displayGroups(currentGroups)

	for {
		fmt.Printf("\n[Enter] to proceed  [e]dit messages  [r]egen  [c]ancel: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "":
			return currentGroups, nil
		case "e":
			currentGroups = editGroupMessages(currentGroups)
			displayGroups(currentGroups)
		case "r":
			spinner := utils.NewSpinner()
			spinner.Start()
			newGroups, err := semantics.ClusterHunks(p, hunks)
			spinner.Stop()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: AI provider failed: %v\n", err)
				continue
			}
			currentGroups = newGroups
			displayGroups(currentGroups)
		case "c":
			fmt.Println("canceled")
			return nil, fmt.Errorf("canceled")
		default:
			fmt.Println("invalid input")
		}
	}
}

func displayGroups(groups []semantics.HunkGroup) {
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("Found %d semantic group(s) to commit:\n", len(groups))
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	for i, group := range groups {
		fmt.Printf("\n📦 Commit %d/%d:\n", i+1, len(groups))

		fileSet := make(map[string]bool)
		for _, hunk := range group.Hunks {
			fileSet[hunk.FilePath] = true
		}
		files := make([]string, 0, len(fileSet))
		for file := range fileSet {
			files = append(files, file)
		}
		fmt.Printf("Files: %s\n\n", strings.Join(files, ", "))

		fmt.Println(group.Message)
		fmt.Println()
	}
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
}

func editGroupMessages(groups []semantics.HunkGroup) []semantics.HunkGroup {
	reader := bufio.NewReader(os.Stdin)
	editedGroups := make([]semantics.HunkGroup, len(groups))
	copy(editedGroups, groups)

	for i := range editedGroups {
		fmt.Printf("\n📝 Edit commit %d/%d\n", i+1, len(editedGroups))
		fmt.Printf("Current: %s\n", editedGroups[i].Message)
		fmt.Printf("[Enter] to keep  [e]dit in editor  [t]ype new message: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "":
			continue
		case "e":
			editedGroups[i].Message = utils.EditInEditor(editedGroups[i].Message)
		case "t":
			fmt.Print("New message: ")
			newMsg, _ := reader.ReadString('\n')
			newMsg = strings.TrimSpace(newMsg)
			if newMsg != "" {
				editedGroups[i].Message = newMsg
			}
		default:
			fmt.Println("Invalid input, keeping original")
		}
	}

	return editedGroups
}
