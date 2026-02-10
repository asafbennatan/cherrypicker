package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v69/github"
)

type Client struct {
	gh *github.Client
}

type MissingCommit struct {
	SHA     string
	Message string
	Author  string
	Date    github.Timestamp
	HTMLURL string
	PRs     []*github.PullRequest
}

func NewClient(token string) *Client {
	return &Client{
		gh: github.NewClient(nil).WithAuthToken(token),
	}
}

func (c *Client) Raw() *github.Client {
	return c.gh
}

// ListMissingCommits returns commits present in main but not in the release branch,
// in topological order as they appear on main (oldest first).
func (c *Client) ListMissingCommits(ctx context.Context, owner, repo, releaseBranch string) ([]MissingCommit, error) {
	comparison, err := c.compareBranches(ctx, owner, repo, releaseBranch, "main")
	if err != nil {
		return nil, fmt.Errorf("comparing branches: %w", err)
	}

	commits := make([]MissingCommit, 0, len(comparison))
	for _, rc := range comparison {
		commits = append(commits, repoCommitToMissing(rc))
	}

	return commits, nil
}

// ListMissingCommitsWithLabel returns commits whose associated PR carries the given label
// and whose merge commit SHA is not present in the release branch.
// Results preserve topological order from the compare API so cherry-picks apply cleanly.
func (c *Client) ListMissingCommitsWithLabel(ctx context.Context, owner, repo, releaseBranch, label string) ([]MissingCommit, error) {
	labeledPRs, err := c.listMergedPRsWithLabel(ctx, owner, repo, label)
	if err != nil {
		return nil, fmt.Errorf("listing PRs with label %q: %w", label, err)
	}

	prBySHA := make(map[string]*github.PullRequest, len(labeledPRs))
	for _, pr := range labeledPRs {
		prBySHA[pr.GetMergeCommitSHA()] = pr
	}

	compareCommits, err := c.compareBranches(ctx, owner, repo, releaseBranch, "main")
	if err != nil {
		return nil, fmt.Errorf("comparing branches: %w", err)
	}

	// Iterate in compare (topological) order, keeping only labeled ones.
	var commits []MissingCommit
	for _, rc := range compareCommits {
		pr, found := prBySHA[rc.GetSHA()]
		if !found {
			continue
		}

		mc := repoCommitToMissing(rc)
		mc.PRs = []*github.PullRequest{pr}
		commits = append(commits, mc)
	}

	return commits, nil
}

// ListMissingCommitsWithoutLabel returns commits not in the release branch
// whose associated PR does NOT carry the given label.
// Results preserve topological order from the compare API.
func (c *Client) ListMissingCommitsWithoutLabel(ctx context.Context, owner, repo, releaseBranch, label string) ([]MissingCommit, error) {
	labeledPRs, err := c.listMergedPRsWithLabel(ctx, owner, repo, label)
	if err != nil {
		return nil, fmt.Errorf("listing PRs with label %q: %w", label, err)
	}

	labeledSHAs := make(map[string]struct{}, len(labeledPRs))
	for _, pr := range labeledPRs {
		labeledSHAs[pr.GetMergeCommitSHA()] = struct{}{}
	}

	compareCommits, err := c.compareBranches(ctx, owner, repo, releaseBranch, "main")
	if err != nil {
		return nil, fmt.Errorf("comparing branches: %w", err)
	}

	var commits []MissingCommit
	for _, rc := range compareCommits {
		if _, hasLabel := labeledSHAs[rc.GetSHA()]; hasLabel {
			continue
		}
		commits = append(commits, repoCommitToMissing(rc))
	}

	return commits, nil
}

func repoCommitToMissing(rc *github.RepositoryCommit) MissingCommit {
	return MissingCommit{
		SHA:     rc.GetSHA(),
		Message: strings.Split(rc.GetCommit().GetMessage(), "\n")[0],
		Author:  rc.GetCommit().GetAuthor().GetName(),
		Date:    *rc.GetCommit().GetAuthor().Date,
		HTMLURL: rc.GetHTMLURL(),
	}
}

// compareBranches returns all commits in head that are not in base.
// It paginates through all results.
func (c *Client) compareBranches(ctx context.Context, owner, repo, base, head string) ([]*github.RepositoryCommit, error) {
	var allCommits []*github.RepositoryCommit
	page := 1

	for {
		comp, resp, err := c.gh.Repositories.CompareCommits(ctx, owner, repo, base, head, &github.ListOptions{
			Page:    page,
			PerPage: 100,
		})
		if err != nil {
			return nil, err
		}

		allCommits = append(allCommits, comp.Commits...)

		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	return allCommits, nil
}

// listMergedPRsWithLabel uses the GitHub Search API to find all merged PRs with the given label.
func (c *Client) listMergedPRsWithLabel(ctx context.Context, owner, repo, label string) ([]*github.PullRequest, error) {
	query := fmt.Sprintf("repo:%s/%s is:pr is:merged label:%s", owner, repo, label)

	var allPRs []*github.PullRequest
	page := 1

	for {
		result, resp, err := c.gh.Search.Issues(ctx, query, &github.SearchOptions{
			Sort:  "created",
			Order: "desc",
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: 100,
			},
		})
		if err != nil {
			return nil, err
		}

		for _, issue := range result.Issues {
			pr, _, err := c.gh.PullRequests.Get(ctx, owner, repo, issue.GetNumber())
			if err != nil {
				return nil, fmt.Errorf("getting PR #%d: %w", issue.GetNumber(), err)
			}
			allPRs = append(allPRs, pr)
		}

		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	return allPRs, nil
}
