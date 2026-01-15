package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Operations defines the interface for git and GitHub CLI operations.
// This allows for mocking in tests.
type Operations interface {
	// GetDefaultBranch returns the default branch name for the repository.
	GetDefaultBranch() (string, error)
	// GetCurrentBranch returns the current branch name.
	GetCurrentBranch() (string, error)
	// IsWorkingTreeClean checks if the working tree is clean (no uncommitted changes).
	IsWorkingTreeClean() (bool, error)
	// BranchExists checks if a branch exists locally or on remote.
	BranchExists(name, remote string) (bool, error)
	// CreateBranch creates and switches to a new branch.
	CreateBranch(name string) error
	// SwitchBranch switches to an existing branch.
	SwitchBranch(name string) error
	// AddFile stages a file for commit.
	AddFile(path string) error
	// Commit commits staged changes with the given message.
	Commit(message string) error
	// Push pushes the current branch to the specified remote.
	Push(remote, branch string) error
	// CreatePR creates a pull request using GitHub CLI.
	CreatePR(base, head, title, body string, draft bool) (string, error)
}

// RealOperations implements Operations using actual git and gh commands.
type RealOperations struct {
	// RepoDir is the repository directory to operate on.
	RepoDir string
}

// NewRealOperations creates a new RealOperations instance for the given directory.
func NewRealOperations(repoDir string) *RealOperations {
	return &RealOperations{RepoDir: repoDir}
}

// runGit runs a git command in the repository directory.
func (r *RealOperations) runGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.RepoDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("git %s failed: %w\nOutput: %s", strings.Join(args, " "), err, output)
	}
	return strings.TrimSpace(string(output)), nil
}

// runGH runs a gh command in the repository directory.
func (r *RealOperations) runGH(args ...string) (string, error) {
	cmd := exec.Command("gh", args...)
	cmd.Dir = r.RepoDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("gh %s failed: %w\nOutput: %s", strings.Join(args, " "), err, output)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetDefaultBranch returns the default branch name using GitHub CLI.
func (r *RealOperations) GetDefaultBranch() (string, error) {
	output, err := r.runGH("repo", "view", "--json", "defaultBranchRef", "--jq", ".defaultBranchRef.name")
	if err != nil {
		return "", fmt.Errorf("failed to get default branch: %w", err)
	}
	if output == "" {
		return "", fmt.Errorf("failed to get default branch: empty response")
	}
	return output, nil
}

// GetCurrentBranch returns the current branch name.
func (r *RealOperations) GetCurrentBranch() (string, error) {
	output, err := r.runGit("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return output, nil
}

// IsWorkingTreeClean checks if the working tree is clean.
func (r *RealOperations) IsWorkingTreeClean() (bool, error) {
	output, err := r.runGit("status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("failed to check working tree status: %w", err)
	}
	return output == "", nil
}

// BranchExists checks if a branch exists locally or on remote.
// Note: This checks using locally available refs. For remote branches, this
// requires that refs have been fetched. It will not perform a git fetch.
func (r *RealOperations) BranchExists(name, remote string) (bool, error) {
	// First check if branch exists locally
	_, err := r.runGit("rev-parse", "--verify", name)
	if err == nil {
		return true, nil
	}

	// Check if branch exists on remote (using locally cached remote refs)
	remoteBranch := fmt.Sprintf("%s/%s", remote, name)
	_, err = r.runGit("rev-parse", "--verify", remoteBranch)
	if err == nil {
		return true, nil
	}

	// Branch doesn't exist locally or on remote
	return false, nil
}

// CreateBranch creates and switches to a new branch.
func (r *RealOperations) CreateBranch(name string) error {
	_, err := r.runGit("checkout", "-b", name)
	if err != nil {
		return fmt.Errorf("failed to create branch %s: %w", name, err)
	}
	return nil
}

// SwitchBranch switches to an existing branch.
func (r *RealOperations) SwitchBranch(name string) error {
	_, err := r.runGit("checkout", name)
	if err != nil {
		return fmt.Errorf("failed to switch to branch %s: %w", name, err)
	}
	return nil
}

// AddFile stages a file for commit.
func (r *RealOperations) AddFile(path string) error {
	_, err := r.runGit("add", path)
	if err != nil {
		return fmt.Errorf("failed to add file %s: %w", path, err)
	}
	return nil
}

// Commit commits staged changes with the given message.
func (r *RealOperations) Commit(message string) error {
	_, err := r.runGit("commit", "-m", message)
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	return nil
}

// Push pushes the current branch to the specified remote.
func (r *RealOperations) Push(remote, branch string) error {
	_, err := r.runGit("push", "-u", remote, branch)
	if err != nil {
		return fmt.Errorf("failed to push to %s/%s: %w", remote, branch, err)
	}
	return nil
}

// CreatePR creates a pull request using GitHub CLI.
func (r *RealOperations) CreatePR(base, head, title, body string, draft bool) (string, error) {
	args := []string{"pr", "create", "--base", base, "--head", head, "--title", title, "--body", body}
	if draft {
		args = append(args, "--draft")
	}
	output, err := r.runGH(args...)
	if err != nil {
		return "", fmt.Errorf("failed to create PR: %w", err)
	}
	return output, nil
}

// FileExists checks if a file exists in the repository.
func FileExists(repoDir, filePath string) bool {
	fullPath := filepath.Join(repoDir, filePath)
	_, err := os.Stat(fullPath)
	return err == nil
}

// ReadFile reads a file from the repository.
func ReadFile(repoDir, filePath string) ([]byte, error) {
	fullPath := filepath.Join(repoDir, filePath)
	return os.ReadFile(fullPath)
}

// WriteFile writes content to a file in the repository, creating directories as needed.
func WriteFile(repoDir, filePath string, content []byte) error {
	fullPath := filepath.Join(repoDir, filePath)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	return os.WriteFile(fullPath, content, 0644)
}
