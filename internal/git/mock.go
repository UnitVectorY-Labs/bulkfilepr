package git

import "fmt"

// MockOperations is a mock implementation of Operations for testing.
type MockOperations struct {
	DefaultBranch     string
	CurrentBranch     string
	IsClean           bool
	CreatedBranches   []string
	SwitchedBranches  []string
	AddedFiles        []string
	Commits           []string
	Pushes            []struct{ Remote, Branch string }
	CreatedPRs        []struct{ Base, Head, Title, Body string; Draft bool }
	PRURLToReturn     string

	// Error fields for simulating failures
	DefaultBranchErr     error
	CurrentBranchErr     error
	IsCleanErr           error
	CreateBranchErr      error
	SwitchBranchErr      error
	AddFileErr           error
	CommitErr            error
	PushErr              error
	CreatePRErr          error
}

// NewMockOperations creates a new MockOperations with default successful behavior.
func NewMockOperations() *MockOperations {
	return &MockOperations{
		DefaultBranch: "main",
		CurrentBranch: "main",
		IsClean:       true,
		PRURLToReturn: "https://github.com/owner/repo/pull/1",
	}
}

// GetDefaultBranch returns the mock default branch.
func (m *MockOperations) GetDefaultBranch() (string, error) {
	if m.DefaultBranchErr != nil {
		return "", m.DefaultBranchErr
	}
	return m.DefaultBranch, nil
}

// GetCurrentBranch returns the mock current branch.
func (m *MockOperations) GetCurrentBranch() (string, error) {
	if m.CurrentBranchErr != nil {
		return "", m.CurrentBranchErr
	}
	return m.CurrentBranch, nil
}

// IsWorkingTreeClean returns the mock clean status.
func (m *MockOperations) IsWorkingTreeClean() (bool, error) {
	if m.IsCleanErr != nil {
		return false, m.IsCleanErr
	}
	return m.IsClean, nil
}

// CreateBranch records the created branch.
func (m *MockOperations) CreateBranch(name string) error {
	if m.CreateBranchErr != nil {
		return m.CreateBranchErr
	}
	m.CreatedBranches = append(m.CreatedBranches, name)
	m.CurrentBranch = name
	return nil
}

// SwitchBranch records the branch switch.
func (m *MockOperations) SwitchBranch(name string) error {
	if m.SwitchBranchErr != nil {
		return m.SwitchBranchErr
	}
	m.SwitchedBranches = append(m.SwitchedBranches, name)
	m.CurrentBranch = name
	return nil
}

// AddFile records the added file.
func (m *MockOperations) AddFile(path string) error {
	if m.AddFileErr != nil {
		return m.AddFileErr
	}
	m.AddedFiles = append(m.AddedFiles, path)
	return nil
}

// Commit records the commit.
func (m *MockOperations) Commit(message string) error {
	if m.CommitErr != nil {
		return m.CommitErr
	}
	m.Commits = append(m.Commits, message)
	return nil
}

// Push records the push.
func (m *MockOperations) Push(remote, branch string) error {
	if m.PushErr != nil {
		return m.PushErr
	}
	m.Pushes = append(m.Pushes, struct{ Remote, Branch string }{remote, branch})
	return nil
}

// CreatePR records the PR creation.
func (m *MockOperations) CreatePR(base, head, title, body string, draft bool) (string, error) {
	if m.CreatePRErr != nil {
		return "", m.CreatePRErr
	}
	m.CreatedPRs = append(m.CreatedPRs, struct{ Base, Head, Title, Body string; Draft bool }{base, head, title, body, draft})
	return m.PRURLToReturn, nil
}

// Validate checks that all expected operations were performed.
func (m *MockOperations) Validate(expectedBranch string) error {
	if len(m.CreatedBranches) > 0 && m.CreatedBranches[len(m.CreatedBranches)-1] != expectedBranch {
		return fmt.Errorf("expected branch %q, got %q", expectedBranch, m.CreatedBranches[len(m.CreatedBranches)-1])
	}
	return nil
}
