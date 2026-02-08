package config

import "testing"

func TestParseMode(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    Mode
		expectError bool
	}{
		{name: "upsert", input: "upsert", expected: ModeUpsert, expectError: false},
		{name: "exists", input: "exists", expected: ModeExists, expectError: false},
		{name: "match", input: "match", expected: ModeMatch, expectError: false},
		{name: "invalid", input: "invalid", expected: "", expectError: true},
		{name: "empty", input: "", expected: "", expectError: true},
		{name: "uppercase", input: "UPSERT", expected: "", expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseMode(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("ParseMode(%q) expected error, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ParseMode(%q) unexpected error: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("ParseMode(%q) = %q, want %q", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid upsert config",
			config: &Config{
				Mode:     ModeUpsert,
				RepoPath: ".github/workflows/ci.yml",
				NewFile:  "/path/to/ci.yml",
			},
			expectError: false,
		},
		{
			name: "valid exists config",
			config: &Config{
				Mode:     ModeExists,
				RepoPath: "Dockerfile",
				NewFile:  "/path/to/Dockerfile",
			},
			expectError: false,
		},
		{
			name: "valid match config",
			config: &Config{
				Mode:         ModeMatch,
				RepoPath:     ".github/workflows/release.yml",
				NewFile:      "/path/to/release.yml",
				ExpectSHA256: "abc123def456",
			},
			expectError: false,
		},
		{
			name: "missing mode",
			config: &Config{
				RepoPath: ".github/workflows/ci.yml",
				NewFile:  "/path/to/ci.yml",
			},
			expectError: true,
		},
		{
			name: "missing repo-path",
			config: &Config{
				Mode:    ModeUpsert,
				NewFile: "/path/to/ci.yml",
			},
			expectError: true,
		},
		{
			name: "missing new-file",
			config: &Config{
				Mode:     ModeUpsert,
				RepoPath: ".github/workflows/ci.yml",
			},
			expectError: true,
		},
		{
			name: "match mode without expect-sha256",
			config: &Config{
				Mode:     ModeMatch,
				RepoPath: ".github/workflows/release.yml",
				NewFile:  "/path/to/release.yml",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError && err == nil {
				t.Error("Validate() expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Repo != "." {
		t.Errorf("DefaultConfig().Repo = %q, want %q", cfg.Repo, ".")
	}
	if cfg.Remote != "origin" {
		t.Errorf("DefaultConfig().Remote = %q, want %q", cfg.Remote, "origin")
	}
}

func TestGetCommitMessage(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name:     "custom message",
			config:   &Config{CommitMessage: "fix: update config", RepoPath: ".github/workflows/ci.yml"},
			expected: "fix: update config",
		},
		{
			name:     "default message",
			config:   &Config{RepoPath: ".github/workflows/ci.yml"},
			expected: "chore: update .github/workflows/ci.yml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetCommitMessage()
			if result != tt.expected {
				t.Errorf("GetCommitMessage() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetPRTitle(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name:     "custom title",
			config:   &Config{PRTitle: "Fix CI workflow", RepoPath: ".github/workflows/ci.yml"},
			expected: "Fix CI workflow",
		},
		{
			name:     "default title",
			config:   &Config{RepoPath: ".github/workflows/ci.yml"},
			expected: "Update .github/workflows/ci.yml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetPRTitle()
			if result != tt.expected {
				t.Errorf("GetPRTitle() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetPRBody(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name:     "custom body",
			config:   &Config{PRBody: "Custom PR body content", RepoPath: ".github/workflows/ci.yml"},
			expected: "Custom PR body content",
		},
		{
			name:     "default body",
			config:   &Config{RepoPath: ".github/workflows/ci.yml"},
			expected: "This PR updates the standardized file at `.github/workflows/ci.yml`.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetPRBody()
			if result != tt.expected {
				t.Errorf("GetPRBody() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetExpectedHashes(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected []string
	}{
		{
			name:     "empty string",
			config:   &Config{ExpectSHA256: ""},
			expected: []string{},
		},
		{
			name:     "single hash",
			config:   &Config{ExpectSHA256: "abc123def456"},
			expected: []string{"abc123def456"},
		},
		{
			name:     "two hashes",
			config:   &Config{ExpectSHA256: "abc123def456,xyz789ghi012"},
			expected: []string{"abc123def456", "xyz789ghi012"},
		},
		{
			name:     "three hashes",
			config:   &Config{ExpectSHA256: "hash1,hash2,hash3"},
			expected: []string{"hash1", "hash2", "hash3"},
		},
		{
			name:     "hashes with spaces",
			config:   &Config{ExpectSHA256: "abc123def456 , xyz789ghi012"},
			expected: []string{"abc123def456", "xyz789ghi012"},
		},
		{
			name:     "hashes with trailing comma",
			config:   &Config{ExpectSHA256: "abc123def456,xyz789ghi012,"},
			expected: []string{"abc123def456", "xyz789ghi012"},
		},
		{
			name:     "real world example",
			config:   &Config{ExpectSHA256: "17ca04878ed554fc89bc73332e013fa8528c7999352a7cea17788e48fecabac6,6bbb6e1ef2fbd220c4dc6853dc40d80e1d060b32f3dfae245f2f4dc8858ccfa1"},
			expected: []string{"17ca04878ed554fc89bc73332e013fa8528c7999352a7cea17788e48fecabac6", "6bbb6e1ef2fbd220c4dc6853dc40d80e1d060b32f3dfae245f2f4dc8858ccfa1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetExpectedHashes()
			if len(result) != len(tt.expected) {
				t.Errorf("GetExpectedHashes() returned %d hashes, want %d", len(result), len(tt.expected))
			}
			for i, hash := range result {
				if hash != tt.expected[i] {
					t.Errorf("GetExpectedHashes()[%d] = %q, want %q", i, hash, tt.expected[i])
				}
			}
		})
	}
}

