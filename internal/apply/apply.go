package apply

import (
	"bytes"
	"fmt"

	"github.com/UnitVectorY-Labs/bulkfilepr/internal/config"
	"github.com/UnitVectorY-Labs/bulkfilepr/internal/git"
	"github.com/UnitVectorY-Labs/bulkfilepr/internal/hash"
)

// Result represents the outcome of an apply operation.
type Result struct {
	// DefaultBranch is the detected default branch name.
	DefaultBranch string
	// Action describes what action was taken or would be taken.
	Action string
	// BranchName is the name of the branch that was/would be created.
	BranchName string
	// PRURL is the URL of the created PR (only set in non-dry-run mode).
	PRURL string
	// NoActionReason explains why no action was taken (if applicable).
	NoActionReason string
}

// Applier handles the apply logic for updating files in a repository.
type Applier struct {
	cfg        *config.Config
	gitOps     git.Operations
	repoDir    string
	newContent []byte
}

// NewApplier creates a new Applier instance.
func NewApplier(cfg *config.Config, gitOps git.Operations, newContent []byte) *Applier {
	return &Applier{
		cfg:        cfg,
		gitOps:     gitOps,
		repoDir:    cfg.Repo,
		newContent: newContent,
	}
}

// Run executes the apply operation and returns the result.
func (a *Applier) Run() (*Result, error) {
	result := &Result{}

	// Step 1: Detect default branch
	defaultBranch, err := a.gitOps.GetDefaultBranch()
	if err != nil {
		return nil, fmt.Errorf("failed to detect default branch: %w", err)
	}
	result.DefaultBranch = defaultBranch

	// Step 2: Verify on default branch
	currentBranch, err := a.gitOps.GetCurrentBranch()
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}
	if currentBranch != defaultBranch {
		return nil, fmt.Errorf("not on default branch: current branch is %q, expected %q", currentBranch, defaultBranch)
	}

	// Step 3: Verify clean working tree
	clean, err := a.gitOps.IsWorkingTreeClean()
	if err != nil {
		return nil, fmt.Errorf("failed to check working tree status: %w", err)
	}
	if !clean {
		return nil, fmt.Errorf("working tree is not clean: please commit or stash your changes")
	}

	// Step 4: Evaluate mode conditions
	shouldUpdate, reason, err := a.evaluateMode()
	if err != nil {
		return nil, err
	}
	if !shouldUpdate {
		result.Action = "no action taken"
		result.NoActionReason = reason
		return result, nil
	}

	// Step 5: Determine branch name
	branchName := a.determineBranchName()
	result.BranchName = branchName

	// Step 6: Execute update (or report dry-run)
	if a.cfg.DryRun {
		result.Action = "would update"
		return result, nil
	}

	// Step 7: Create branch
	if err := a.gitOps.CreateBranch(branchName); err != nil {
		return nil, fmt.Errorf("failed to create branch: %w", err)
	}

	// Use a cleanup function to switch back to default branch on error
	var updateErr error
	defer func() {
		if updateErr != nil {
			// Best effort: switch back to default branch on error
			_ = a.gitOps.SwitchBranch(defaultBranch)
		}
	}()

	// Step 8: Write file
	if err := git.WriteFile(a.repoDir, a.cfg.RepoPath, a.newContent); err != nil {
		updateErr = fmt.Errorf("failed to write file: %w", err)
		return nil, updateErr
	}

	// Step 9: Stage file
	if err := a.gitOps.AddFile(a.cfg.RepoPath); err != nil {
		updateErr = fmt.Errorf("failed to stage file: %w", err)
		return nil, updateErr
	}

	// Step 10: Commit
	if err := a.gitOps.Commit(a.cfg.GetCommitMessage()); err != nil {
		updateErr = fmt.Errorf("failed to commit: %w", err)
		return nil, updateErr
	}

	// Step 11: Push
	if err := a.gitOps.Push(a.cfg.Remote, branchName); err != nil {
		updateErr = fmt.Errorf("failed to push: %w", err)
		return nil, updateErr
	}

	// Step 12: Create PR
	prURL, err := a.gitOps.CreatePR(defaultBranch, branchName, a.cfg.GetPRTitle(), a.cfg.GetPRBody(), a.cfg.Draft)
	if err != nil {
		updateErr = fmt.Errorf("failed to create PR: %w", err)
		return nil, updateErr
	}
	result.PRURL = prURL
	result.Action = "updated"

	// Step 13: Switch back to default branch (best effort)
	_ = a.gitOps.SwitchBranch(defaultBranch)

	return result, nil
}

// evaluateMode checks if the update should proceed based on the mode.
// Returns (shouldUpdate, reason, error).
func (a *Applier) evaluateMode() (bool, string, error) {
	fileExists := git.FileExists(a.repoDir, a.cfg.RepoPath)

	switch a.cfg.Mode {
	case config.ModeUpsert:
		// Always write unless content is identical
		if fileExists {
			existingContent, err := git.ReadFile(a.repoDir, a.cfg.RepoPath)
			if err != nil {
				return false, "", fmt.Errorf("failed to read existing file: %w", err)
			}
			if bytes.Equal(existingContent, a.newContent) {
				return false, "file content is already identical", nil
			}
		}
		return true, "", nil

	case config.ModeExists:
		if !fileExists {
			return false, "file does not exist", nil
		}
		existingContent, err := git.ReadFile(a.repoDir, a.cfg.RepoPath)
		if err != nil {
			return false, "", fmt.Errorf("failed to read existing file: %w", err)
		}
		if bytes.Equal(existingContent, a.newContent) {
			return false, "file content is already identical", nil
		}
		return true, "", nil

	case config.ModeMatch:
		if !fileExists {
			return false, "file does not exist", nil
		}
		existingContent, err := git.ReadFile(a.repoDir, a.cfg.RepoPath)
		if err != nil {
			return false, "", fmt.Errorf("failed to read existing file: %w", err)
		}
		existingHash := hash.SHA256Bytes(existingContent)
		if existingHash != a.cfg.ExpectSHA256 {
			return false, fmt.Sprintf("file hash mismatch: expected %s, got %s", a.cfg.ExpectSHA256, existingHash), nil
		}
		if bytes.Equal(existingContent, a.newContent) {
			return false, "file content is already identical", nil
		}
		return true, "", nil

	default:
		return false, "", fmt.Errorf("unknown mode: %s", a.cfg.Mode)
	}
}

// determineBranchName returns the branch name to use.
func (a *Applier) determineBranchName() string {
	if a.cfg.Branch != "" {
		return a.cfg.Branch
	}
	// Generate branch name from hash of new file content
	contentHash := hash.SHA256Bytes(a.newContent)
	truncatedHash := hash.TruncatedHash(contentHash, 12)
	return fmt.Sprintf("bulkfilepr/%s", truncatedHash)
}
