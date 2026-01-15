package config

import "fmt"

// Mode represents the file update mode for the apply command.
type Mode string

const (
	// ModeUpsert always writes the new file (create if missing, update if exists).
	ModeUpsert Mode = "upsert"
	// ModeExists only updates if the file already exists.
	ModeExists Mode = "exists"
	// ModeMatch only updates if the file exists and matches the expected SHA-256 hash.
	ModeMatch Mode = "match"
)

// ParseMode converts a string to a Mode, returning an error if invalid.
func ParseMode(s string) (Mode, error) {
	switch s {
	case string(ModeUpsert):
		return ModeUpsert, nil
	case string(ModeExists):
		return ModeExists, nil
	case string(ModeMatch):
		return ModeMatch, nil
	default:
		return "", fmt.Errorf("invalid mode: %q, must be one of: upsert, exists, match", s)
	}
}

// Config holds all configuration options for the apply command.
type Config struct {
	// Mode specifies the file update mode (upsert, exists, match).
	Mode Mode
	// RepoPath is the destination file path inside the repo, relative to repo root.
	RepoPath string
	// NewFile is the path to the new file content to write.
	NewFile string
	// Repo is the repository directory (default: .).
	Repo string
	// Branch is the name of the branch to create (optional, auto-generated if empty).
	Branch string
	// CommitMessage is the commit message (default: "chore: update {repo-path}").
	CommitMessage string
	// PRTitle is the PR title (default: "Update {repo-path}").
	PRTitle string
	// PRBody is the PR body content.
	PRBody string
	// Draft indicates whether to create the PR as a draft.
	Draft bool
	// DryRun indicates whether to run in dry-run mode (no changes made).
	DryRun bool
	// Remote is the git remote name (default: origin).
	Remote string
	// ExpectSHA256 is the expected SHA-256 hash for match mode.
	ExpectSHA256 string
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() *Config {
	return &Config{
		Repo:   ".",
		Remote: "origin",
	}
}

// Validate checks that the configuration is valid.
func (c *Config) Validate() error {
	if c.Mode == "" {
		return fmt.Errorf("mode is required")
	}
	if c.RepoPath == "" {
		return fmt.Errorf("repo-path is required")
	}
	if c.NewFile == "" {
		return fmt.Errorf("new-file is required")
	}
	if c.Mode == ModeMatch && c.ExpectSHA256 == "" {
		return fmt.Errorf("expect-sha256 is required when mode is 'match'")
	}
	return nil
}

// GetCommitMessage returns the commit message, substituting defaults if necessary.
func (c *Config) GetCommitMessage() string {
	if c.CommitMessage != "" {
		return c.CommitMessage
	}
	return fmt.Sprintf("chore: update %s", c.RepoPath)
}

// GetPRTitle returns the PR title, substituting defaults if necessary.
func (c *Config) GetPRTitle() string {
	if c.PRTitle != "" {
		return c.PRTitle
	}
	return fmt.Sprintf("Update %s", c.RepoPath)
}

// GetPRBody returns the PR body, substituting defaults if necessary.
func (c *Config) GetPRBody() string {
	if c.PRBody != "" {
		return c.PRBody
	}
	return fmt.Sprintf("This PR updates the standardized file at `%s`.", c.RepoPath)
}
