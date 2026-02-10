package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/abennata/cherrypicker/internal/git"
	"github.com/abennata/cherrypicker/internal/model"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var inputFile string

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a cherry-pick branch from a list of commits",
	Long: `Create a new branch off the release branch and cherry-pick all commits
from the input file (YAML format produced by: list -o yaml).

The release branch is read from the YAML file metadata.

Steps performed:
  1. Checkout the release branch (creates local tracking branch if needed)
  2. Pull latest commits
  3. Create a new branch named cherrypick-<date>
  4. Cherry-pick each commit from the input file`,
	Args: cobra.NoArgs,
	RunE: runCreate,
}

func init() {
	createCmd.Flags().StringVarP(&inputFile, "file", "f", "", "path to YAML file containing commits (output of list -o yaml)")
	createCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	cpFile, err := parseCherryPickFile(inputFile)
	if err != nil {
		return fmt.Errorf("parsing input file: %w", err)
	}
	if len(cpFile.Commits) == 0 {
		return fmt.Errorf("no commits found in input file")
	}

	releaseBranch := cpFile.ReleaseBranch
	fmt.Fprintf(os.Stderr, "Repo:           %s\n", cpFile.Repo)
	fmt.Fprintf(os.Stderr, "Release branch: %s\n", releaseBranch)
	fmt.Fprintf(os.Stderr, "Commits:        %d\n", len(cpFile.Commits))

	if err := git.Checkout(releaseBranch); err != nil {
		return fmt.Errorf("checking out %s: %w", releaseBranch, err)
	}

	if err := git.Pull(); err != nil {
		return fmt.Errorf("pulling latest: %w", err)
	}

	branchName := fmt.Sprintf("cherrypick-%s", time.Now().Format("20060102-150405"))
	if err := git.CreateBranch(branchName); err != nil {
		return fmt.Errorf("creating branch %s: %w", branchName, err)
	}

	for i, c := range cpFile.Commits {
		fmt.Fprintf(os.Stderr, "\n[%d/%d] Cherry-picking %s\n", i+1, len(cpFile.Commits), c.SHA)
		if err := git.CherryPick(c.SHA); err != nil {
			return fmt.Errorf("cherry-pick failed on %s: %w\nResolve the conflict and run: git cherry-pick --continue", c.SHA, err)
		}
	}

	fmt.Fprintf(os.Stderr, "\nDone. %d commit(s) cherry-picked onto branch %s\n", len(cpFile.Commits), branchName)
	return nil
}

func parseCherryPickFile(path string) (*model.CherryPickFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cpFile model.CherryPickFile
	if err := yaml.Unmarshal(data, &cpFile); err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}

	if cpFile.ReleaseBranch == "" {
		return nil, fmt.Errorf("releaseBranch is missing from YAML file")
	}

	return &cpFile, nil
}
