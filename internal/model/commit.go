package model

import "time"

// CherryPickFile is the top-level YAML structure shared between list (output)
// and create (input).
type CherryPickFile struct {
	Repo          string   `yaml:"repo"`
	ReleaseBranch string   `yaml:"releaseBranch"`
	Label         string   `yaml:"label,omitempty"`
	Commits       []Commit `yaml:"commits"`
}

// Commit is the serializable representation of a missing commit.
type Commit struct {
	SHA     string    `yaml:"sha"`
	Date    time.Time `yaml:"date"`
	Author  string    `yaml:"author"`
	Message string    `yaml:"message"`
	PR      int       `yaml:"pr,omitempty"`
}
