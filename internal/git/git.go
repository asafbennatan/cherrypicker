package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Run executes a git command, forwarding stdout/stderr to the terminal.
func Run(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Fprintf(os.Stderr, "=> git %s\n", strings.Join(args, " "))
	return cmd.Run()
}

// Checkout switches to the given branch, creating a local tracking branch from
// origin if the branch doesn't exist locally.
func Checkout(branch string) error {
	if err := Run("checkout", branch); err == nil {
		return nil
	}
	return Run("checkout", "-b", branch, "origin/"+branch)
}

// Pull pulls the latest commits for the current branch.
func Pull() error {
	return Run("pull", "--ff-only")
}

// CreateBranch creates and switches to a new branch off the current HEAD.
func CreateBranch(name string) error {
	return Run("checkout", "-b", name)
}

// CherryPick cherry-picks a single commit by SHA.
func CherryPick(sha string) error {
	return Run("cherry-pick", sha)
}
