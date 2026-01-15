package apply

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/UnitVectorY-Labs/bulkfilepr/internal/config"
	"github.com/UnitVectorY-Labs/bulkfilepr/internal/git"
	"github.com/UnitVectorY-Labs/bulkfilepr/internal/hash"
)

func TestApplierUpsertModeNewFile(t *testing.T) {
	tmpDir := t.TempDir()
	mock := git.NewMockOperations()
	newContent := []byte("new content\n")

	cfg := &config.Config{
		Mode:     config.ModeUpsert,
		RepoPath: "test/file.txt",
		NewFile:  "/path/to/new.txt",
		Repo:     tmpDir,
		Remote:   "origin",
	}

	applier := NewApplier(cfg, mock, newContent)
	result, err := applier.Run()

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Action != "updated" {
		t.Errorf("Action = %q, want %q", result.Action, "updated")
	}
	if result.DefaultBranch != "main" {
		t.Errorf("DefaultBranch = %q, want %q", result.DefaultBranch, "main")
	}
	if result.PRURL == "" {
		t.Error("PRURL is empty, expected non-empty")
	}
	if len(mock.CreatedBranches) != 1 {
		t.Errorf("CreatedBranches length = %d, want 1", len(mock.CreatedBranches))
	}
	if len(mock.Commits) != 1 {
		t.Errorf("Commits length = %d, want 1", len(mock.Commits))
	}
}

func TestApplierUpsertModeIdenticalContent(t *testing.T) {
	tmpDir := t.TempDir()
	mock := git.NewMockOperations()
	content := []byte("existing content\n")

	// Create existing file with same content
	testDir := filepath.Join(tmpDir, "test")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testDir, "file.txt"), content, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cfg := &config.Config{
		Mode:     config.ModeUpsert,
		RepoPath: "test/file.txt",
		NewFile:  "/path/to/new.txt",
		Repo:     tmpDir,
		Remote:   "origin",
	}

	applier := NewApplier(cfg, mock, content)
	result, err := applier.Run()

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Action != "no action taken" {
		t.Errorf("Action = %q, want %q", result.Action, "no action taken")
	}
	if result.NoActionReason != "file content is already identical" {
		t.Errorf("NoActionReason = %q, want %q", result.NoActionReason, "file content is already identical")
	}
	if len(mock.CreatedBranches) != 0 {
		t.Errorf("CreatedBranches length = %d, want 0", len(mock.CreatedBranches))
	}
}

func TestApplierExistsModeFileNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	mock := git.NewMockOperations()
	newContent := []byte("new content\n")

	cfg := &config.Config{
		Mode:     config.ModeExists,
		RepoPath: "nonexistent/file.txt",
		NewFile:  "/path/to/new.txt",
		Repo:     tmpDir,
		Remote:   "origin",
	}

	applier := NewApplier(cfg, mock, newContent)
	result, err := applier.Run()

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Action != "no action taken" {
		t.Errorf("Action = %q, want %q", result.Action, "no action taken")
	}
	if result.NoActionReason != "file does not exist" {
		t.Errorf("NoActionReason = %q, want %q", result.NoActionReason, "file does not exist")
	}
}

func TestApplierExistsModeFileExists(t *testing.T) {
	tmpDir := t.TempDir()
	mock := git.NewMockOperations()
	oldContent := []byte("old content\n")
	newContent := []byte("new content\n")

	// Create existing file
	testDir := filepath.Join(tmpDir, "test")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testDir, "file.txt"), oldContent, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cfg := &config.Config{
		Mode:     config.ModeExists,
		RepoPath: "test/file.txt",
		NewFile:  "/path/to/new.txt",
		Repo:     tmpDir,
		Remote:   "origin",
	}

	applier := NewApplier(cfg, mock, newContent)
	result, err := applier.Run()

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Action != "updated" {
		t.Errorf("Action = %q, want %q", result.Action, "updated")
	}
}

func TestApplierMatchModeHashMismatch(t *testing.T) {
	tmpDir := t.TempDir()
	mock := git.NewMockOperations()
	existingContent := []byte("existing content\n")
	newContent := []byte("new content\n")

	// Create existing file
	testDir := filepath.Join(tmpDir, "test")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testDir, "file.txt"), existingContent, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cfg := &config.Config{
		Mode:         config.ModeMatch,
		RepoPath:     "test/file.txt",
		NewFile:      "/path/to/new.txt",
		Repo:         tmpDir,
		Remote:       "origin",
		ExpectSHA256: "wronghash123456789",
	}

	applier := NewApplier(cfg, mock, newContent)
	result, err := applier.Run()

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Action != "no action taken" {
		t.Errorf("Action = %q, want %q", result.Action, "no action taken")
	}
	if result.NoActionReason == "" {
		t.Error("NoActionReason is empty, expected hash mismatch message")
	}
}

func TestApplierMatchModeHashMatch(t *testing.T) {
	tmpDir := t.TempDir()
	mock := git.NewMockOperations()
	existingContent := []byte("existing content\n")
	newContent := []byte("new content\n")

	// Create existing file
	testDir := filepath.Join(tmpDir, "test")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testDir, "file.txt"), existingContent, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Calculate the correct hash
	expectedHash := hash.SHA256Bytes(existingContent)

	cfg := &config.Config{
		Mode:         config.ModeMatch,
		RepoPath:     "test/file.txt",
		NewFile:      "/path/to/new.txt",
		Repo:         tmpDir,
		Remote:       "origin",
		ExpectSHA256: expectedHash,
	}

	applier := NewApplier(cfg, mock, newContent)
	result, err := applier.Run()

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Action != "updated" {
		t.Errorf("Action = %q, want %q", result.Action, "updated")
	}
}

func TestApplierDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	mock := git.NewMockOperations()
	newContent := []byte("new content\n")

	cfg := &config.Config{
		Mode:     config.ModeUpsert,
		RepoPath: "test/file.txt",
		NewFile:  "/path/to/new.txt",
		Repo:     tmpDir,
		Remote:   "origin",
		DryRun:   true,
	}

	applier := NewApplier(cfg, mock, newContent)
	result, err := applier.Run()

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Action != "would update" {
		t.Errorf("Action = %q, want %q", result.Action, "would update")
	}
	if result.BranchName == "" {
		t.Error("BranchName is empty, expected non-empty")
	}
	// Verify no actual operations were performed
	if len(mock.CreatedBranches) != 0 {
		t.Errorf("CreatedBranches length = %d, want 0", len(mock.CreatedBranches))
	}
	if len(mock.Commits) != 0 {
		t.Errorf("Commits length = %d, want 0", len(mock.Commits))
	}
	if len(mock.CreatedPRs) != 0 {
		t.Errorf("CreatedPRs length = %d, want 0", len(mock.CreatedPRs))
	}
}

func TestApplierNotOnDefaultBranch(t *testing.T) {
	tmpDir := t.TempDir()
	mock := git.NewMockOperations()
	mock.CurrentBranch = "feature-branch"
	newContent := []byte("new content\n")

	cfg := &config.Config{
		Mode:     config.ModeUpsert,
		RepoPath: "test/file.txt",
		NewFile:  "/path/to/new.txt",
		Repo:     tmpDir,
		Remote:   "origin",
	}

	applier := NewApplier(cfg, mock, newContent)
	_, err := applier.Run()

	if err == nil {
		t.Error("Run() expected error for not on default branch, got nil")
	}
}

func TestApplierDirtyWorkingTree(t *testing.T) {
	tmpDir := t.TempDir()
	mock := git.NewMockOperations()
	mock.IsClean = false
	newContent := []byte("new content\n")

	cfg := &config.Config{
		Mode:     config.ModeUpsert,
		RepoPath: "test/file.txt",
		NewFile:  "/path/to/new.txt",
		Repo:     tmpDir,
		Remote:   "origin",
	}

	applier := NewApplier(cfg, mock, newContent)
	_, err := applier.Run()

	if err == nil {
		t.Error("Run() expected error for dirty working tree, got nil")
	}
}

func TestApplierCustomBranchName(t *testing.T) {
	tmpDir := t.TempDir()
	mock := git.NewMockOperations()
	newContent := []byte("new content\n")

	cfg := &config.Config{
		Mode:     config.ModeUpsert,
		RepoPath: "test/file.txt",
		NewFile:  "/path/to/new.txt",
		Repo:     tmpDir,
		Remote:   "origin",
		Branch:   "custom-branch-name",
	}

	applier := NewApplier(cfg, mock, newContent)
	result, err := applier.Run()

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.BranchName != "custom-branch-name" {
		t.Errorf("BranchName = %q, want %q", result.BranchName, "custom-branch-name")
	}
	if len(mock.CreatedBranches) != 1 || mock.CreatedBranches[0] != "custom-branch-name" {
		t.Errorf("CreatedBranches = %v, want [custom-branch-name]", mock.CreatedBranches)
	}
}

func TestApplierAutoBranchName(t *testing.T) {
	tmpDir := t.TempDir()
	mock := git.NewMockOperations()
	newContent := []byte("new content\n")

	cfg := &config.Config{
		Mode:     config.ModeUpsert,
		RepoPath: "test/file.txt",
		NewFile:  "/path/to/new.txt",
		Repo:     tmpDir,
		Remote:   "origin",
	}

	applier := NewApplier(cfg, mock, newContent)
	result, err := applier.Run()

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	// Branch name should be bulkfilepr/{12-char-hash}
	expectedHash := hash.SHA256Bytes(newContent)
	truncatedHash := hash.TruncatedHash(expectedHash, 12)
	expectedBranch := "bulkfilepr/" + truncatedHash

	if result.BranchName != expectedBranch {
		t.Errorf("BranchName = %q, want %q", result.BranchName, expectedBranch)
	}
}

func TestApplierDraftPR(t *testing.T) {
	tmpDir := t.TempDir()
	mock := git.NewMockOperations()
	newContent := []byte("new content\n")

	cfg := &config.Config{
		Mode:     config.ModeUpsert,
		RepoPath: "test/file.txt",
		NewFile:  "/path/to/new.txt",
		Repo:     tmpDir,
		Remote:   "origin",
		Draft:    true,
	}

	applier := NewApplier(cfg, mock, newContent)
	_, err := applier.Run()

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(mock.CreatedPRs) != 1 {
		t.Fatalf("CreatedPRs length = %d, want 1", len(mock.CreatedPRs))
	}
	if !mock.CreatedPRs[0].Draft {
		t.Error("CreatedPRs[0].Draft = false, want true")
	}
}

func TestApplierCustomCommitMessage(t *testing.T) {
	tmpDir := t.TempDir()
	mock := git.NewMockOperations()
	newContent := []byte("new content\n")

	cfg := &config.Config{
		Mode:          config.ModeUpsert,
		RepoPath:      "test/file.txt",
		NewFile:       "/path/to/new.txt",
		Repo:          tmpDir,
		Remote:        "origin",
		CommitMessage: "fix: update workflow",
	}

	applier := NewApplier(cfg, mock, newContent)
	_, err := applier.Run()

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(mock.Commits) != 1 {
		t.Fatalf("Commits length = %d, want 1", len(mock.Commits))
	}
	if mock.Commits[0] != "fix: update workflow" {
		t.Errorf("Commits[0] = %q, want %q", mock.Commits[0], "fix: update workflow")
	}
}
