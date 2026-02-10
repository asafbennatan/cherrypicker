# cherrypicker

A CLI tool that helps manage cherry-picking commits from `main` into release branches using the GitHub API.

## Prerequisites

- Go 1.21+
- A `GITHUB_TOKEN` environment variable with repo read access

## Installation

```bash
make build
```

The binary is written to `bin/cherrypicker`.

## Usage

### List missing commits

Show all commits on `main` that are not present on a release branch:

```bash
cherrypicker list <owner/repo> <release-branch>
```

Example:

```bash
cherrypicker list flightctl/flightctl release-1.1
```

### Filter by PR label

Only show commits whose associated PR carries a specific label:

```bash
cherrypicker list <owner/repo> <release-branch> --with-label <label>
```

Example:

```bash
cherrypicker list flightctl/flightctl release-1.1 --with-label backport-to-1.1
```

### Output formats

By default output is a human-readable table. Use `-o yaml` to produce a YAML file that can be fed into the `create` command:

```bash
cherrypicker list flightctl/flightctl release-1.1 --with-label backport-to-1.1 -o yaml > commits.yaml
```

The YAML file includes metadata (repo URL, release branch, label) and the full list of commits in topological order:

```yaml
repo: https://github.com/flightctl/flightctl
releaseBranch: release-1.1
label: backport-to-1.1
commits:
  - sha: abc123...
    date: 2025-01-15T10:30:00Z
    author: Jane Doe
    message: "EDM-1234: some change"
    pr: 456
```

### Create a cherry-pick branch

From inside a cloned repo, apply the commits from a YAML file:

```bash
cd ~/dev/flightctl
cherrypicker create -f commits.yaml
```

This will:

1. Checkout the release branch (creates a local tracking branch if needed)
2. Pull latest changes
3. Create a new branch named `cherrypick-<timestamp>`
4. Cherry-pick each commit in order

The branch is not pushed automatically. Review the result and push when ready.

If a cherry-pick hits a conflict, the tool stops and prints the failing SHA. Resolve the conflict, then run `git cherry-pick --continue` to proceed with the remaining commits.
