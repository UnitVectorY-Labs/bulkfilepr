package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMockOperations(t *testing.T) {
	mock := NewMockOperations()

	// Test default values
	if mock.DefaultBranch != "main" {
		t.Errorf("DefaultBranch = %q, want %q", mock.DefaultBranch, "main")
	}
	if mock.CurrentBranch != "main" {
		t.Errorf("CurrentBranch = %q, want %q", mock.CurrentBranch, "main")
	}
	if !mock.IsClean {
		t.Error("IsClean = false, want true")
	}

	// Test GetDefaultBranch
	branch, err := mock.GetDefaultBranch()
	if err != nil {
		t.Errorf("GetDefaultBranch() error = %v", err)
	}
	if branch != "main" {
		t.Errorf("GetDefaultBranch() = %q, want %q", branch, "main")
	}

	// Test GetCurrentBranch
	branch, err = mock.GetCurrentBranch()
	if err != nil {
		t.Errorf("GetCurrentBranch() error = %v", err)
	}
	if branch != "main" {
		t.Errorf("GetCurrentBranch() = %q, want %q", branch, "main")
	}

	// Test IsWorkingTreeClean
	clean, err := mock.IsWorkingTreeClean()
	if err != nil {
		t.Errorf("IsWorkingTreeClean() error = %v", err)
	}
	if !clean {
		t.Error("IsWorkingTreeClean() = false, want true")
	}

	// Test CreateBranch
	err = mock.CreateBranch("feature/test")
	if err != nil {
		t.Errorf("CreateBranch() error = %v", err)
	}
	if len(mock.CreatedBranches) != 1 || mock.CreatedBranches[0] != "feature/test" {
		t.Errorf("CreatedBranches = %v, want [feature/test]", mock.CreatedBranches)
	}

	// Test SwitchBranch
	err = mock.SwitchBranch("main")
	if err != nil {
		t.Errorf("SwitchBranch() error = %v", err)
	}
	if len(mock.SwitchedBranches) != 1 || mock.SwitchedBranches[0] != "main" {
		t.Errorf("SwitchedBranches = %v, want [main]", mock.SwitchedBranches)
	}

	// Test AddFile
	err = mock.AddFile("test.txt")
	if err != nil {
		t.Errorf("AddFile() error = %v", err)
	}
	if len(mock.AddedFiles) != 1 || mock.AddedFiles[0] != "test.txt" {
		t.Errorf("AddedFiles = %v, want [test.txt]", mock.AddedFiles)
	}

	// Test Commit
	err = mock.Commit("test commit")
	if err != nil {
		t.Errorf("Commit() error = %v", err)
	}
	if len(mock.Commits) != 1 || mock.Commits[0] != "test commit" {
		t.Errorf("Commits = %v, want [test commit]", mock.Commits)
	}

	// Test Push
	err = mock.Push("origin", "feature/test")
	if err != nil {
		t.Errorf("Push() error = %v", err)
	}
	if len(mock.Pushes) != 1 || mock.Pushes[0].Remote != "origin" || mock.Pushes[0].Branch != "feature/test" {
		t.Errorf("Pushes = %v, want [{origin feature/test}]", mock.Pushes)
	}

	// Test CreatePR
	url, err := mock.CreatePR("main", "feature/test", "Test PR", "Test body", false)
	if err != nil {
		t.Errorf("CreatePR() error = %v", err)
	}
	if url != mock.PRURLToReturn {
		t.Errorf("CreatePR() = %q, want %q", url, mock.PRURLToReturn)
	}
	if len(mock.CreatedPRs) != 1 {
		t.Errorf("CreatedPRs length = %d, want 1", len(mock.CreatedPRs))
	}
}

func TestFileOperations(t *testing.T) {
	tmpDir := t.TempDir()

	// Test FileExists for non-existent file
	if FileExists(tmpDir, "nonexistent.txt") {
		t.Error("FileExists() = true for nonexistent file, want false")
	}

	// Test WriteFile and FileExists
	content := []byte("test content")
	err := WriteFile(tmpDir, "test.txt", content)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if !FileExists(tmpDir, "test.txt") {
		t.Error("FileExists() = false after WriteFile, want true")
	}

	// Test ReadFile
	readContent, err := ReadFile(tmpDir, "test.txt")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(readContent) != string(content) {
		t.Errorf("ReadFile() = %q, want %q", readContent, content)
	}

	// Test WriteFile with nested path
	err = WriteFile(tmpDir, "nested/dir/file.txt", []byte("nested content"))
	if err != nil {
		t.Fatalf("WriteFile() with nested path error = %v", err)
	}
	if !FileExists(tmpDir, "nested/dir/file.txt") {
		t.Error("FileExists() = false for nested file, want true")
	}
}

func TestReadFileError(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := ReadFile(tmpDir, "nonexistent.txt")
	if err == nil {
		t.Error("ReadFile() expected error for nonexistent file, got nil")
	}
}

func TestWriteFilePermissionError(t *testing.T) {
	// Create a directory with no write permission
	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	if err := os.MkdirAll(readOnlyDir, 0444); err != nil {
		t.Fatalf("failed to create read-only dir: %v", err)
	}
	defer os.Chmod(readOnlyDir, 0755) // Cleanup: restore permissions

	err := WriteFile(readOnlyDir, "test.txt", []byte("content"))
	if err == nil {
		t.Error("WriteFile() expected error for read-only directory, got nil")
	}
}
