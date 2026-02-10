package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	gh "github.com/abennata/cherrypicker/internal/github"
	"github.com/abennata/cherrypicker/internal/model"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	withLabel    string
	withoutLabel string
	outputFormat string
)

var listCmd = &cobra.Command{
	Use:   "list <owner/repo> <release-branch-name>",
	Short: "List commits in main that are missing from a release branch",
	Long: `List all commits present in main but absent from the specified release branch.
Use --with-label to filter to only commits whose associated PR carries a specific label.
Use --without-label to exclude commits whose associated PR carries a specific label.`,
	Args: cobra.ExactArgs(2),
	RunE: runList,
}

func init() {
	listCmd.Flags().StringVar(&withLabel, "with-label", "", "filter to commits whose associated PR has this label")
	listCmd.Flags().StringVar(&withoutLabel, "without-label", "", "exclude commits whose associated PR has this label")
	listCmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "output format: table or yaml")
	listCmd.MarkFlagsMutuallyExclusive("with-label", "without-label")
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	if outputFormat != "table" && outputFormat != "yaml" {
		return fmt.Errorf("unsupported output format %q, must be table or yaml", outputFormat)
	}

	owner, repo, err := parseRepo(args[0])
	if err != nil {
		return err
	}
	releaseBranch := args[1]

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable is required")
	}

	client := gh.NewClient(token)
	ctx := context.Background()

	var commits []gh.MissingCommit
	switch {
	case withLabel != "":
		fmt.Fprintf(os.Stderr, "Fetching commits with label %q...\n", withLabel)
		commits, err = client.ListMissingCommitsWithLabel(ctx, owner, repo, releaseBranch, withLabel)
	case withoutLabel != "":
		fmt.Fprintf(os.Stderr, "Fetching commits without label %q...\n", withoutLabel)
		commits, err = client.ListMissingCommitsWithoutLabel(ctx, owner, repo, releaseBranch, withoutLabel)
	default:
		commits, err = client.ListMissingCommits(ctx, owner, repo, releaseBranch)
	}
	if err != nil {
		return fmt.Errorf("listing missing commits: %w", err)
	}

	if len(commits) == 0 {
		fmt.Println("No missing commits found.")
		return nil
	}

	switch outputFormat {
	case "yaml":
		return printYAML(args[0], releaseBranch, withLabel, commits)
	default:
		return printTable(commits, withLabel != "")
	}
}

func printTable(commits []gh.MissingCommit, showPR bool) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if showPR {
		fmt.Fprintf(w, "SHA\tDATE\tAUTHOR\tPR\tMESSAGE\n")
		for _, c := range commits {
			prRef := ""
			if len(c.PRs) > 0 {
				prRef = fmt.Sprintf("#%d", c.PRs[0].GetNumber())
			}
			fmt.Fprintf(w, "%.12s\t%s\t%s\t%s\t%s\n",
				c.SHA,
				c.Date.Time.Format("2006-01-02 15:04"),
				c.Author,
				prRef,
				truncate(c.Message, 72),
			)
		}
	} else {
		fmt.Fprintf(w, "SHA\tDATE\tAUTHOR\tMESSAGE\n")
		for _, c := range commits {
			fmt.Fprintf(w, "%.12s\t%s\t%s\t%s\n",
				c.SHA,
				c.Date.Time.Format("2006-01-02 15:04"),
				c.Author,
				truncate(c.Message, 80),
			)
		}
	}

	w.Flush()
	fmt.Fprintf(os.Stderr, "\nTotal: %d missing commits\n", len(commits))
	return nil
}

func printYAML(repo, releaseBranch, label string, commits []gh.MissingCommit) error {
	out := model.CherryPickFile{
		Repo:          "https://github.com/" + repo,
		ReleaseBranch: releaseBranch,
		Label:         label,
	}
	for _, c := range commits {
		mc := model.Commit{
			SHA:     c.SHA,
			Date:    c.Date.Time,
			Author:  c.Author,
			Message: c.Message,
		}
		if len(c.PRs) > 0 {
			mc.PR = c.PRs[0].GetNumber()
		}
		out.Commits = append(out.Commits, mc)
	}

	enc := yaml.NewEncoder(os.Stdout)
	enc.SetIndent(2)
	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("encoding yaml: %w", err)
	}
	return enc.Close()
}

func parseRepo(repoArg string) (string, string, error) {
	parts := strings.SplitN(repoArg, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("repo must be in owner/repo format, got %q", repoArg)
	}
	return parts[0], parts[1], nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
