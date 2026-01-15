package main

import (
	"testing"
)

func TestRunMissingCommand(t *testing.T) {
	exitCode := run([]string{})
	if exitCode != exitInvalidUsage {
		t.Errorf("run([]) = %d, want %d", exitCode, exitInvalidUsage)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	exitCode := run([]string{"unknown"})
	if exitCode != exitInvalidUsage {
		t.Errorf("run([unknown]) = %d, want %d", exitCode, exitInvalidUsage)
	}
}

func TestRunVersion(t *testing.T) {
	exitCode := run([]string{"--version"})
	if exitCode != exitSuccess {
		t.Errorf("run([--version]) = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunMissingMode(t *testing.T) {
	exitCode := run([]string{"apply", "--repo-path", "test.txt", "--new-file", "test.txt"})
	if exitCode != exitInvalidUsage {
		t.Errorf("run() = %d, want %d", exitCode, exitInvalidUsage)
	}
}

func TestRunMissingRepoPath(t *testing.T) {
	exitCode := run([]string{"apply", "--mode", "upsert", "--new-file", "test.txt"})
	if exitCode != exitInvalidUsage {
		t.Errorf("run() = %d, want %d", exitCode, exitInvalidUsage)
	}
}

func TestRunMissingNewFile(t *testing.T) {
	exitCode := run([]string{"apply", "--mode", "upsert", "--repo-path", "test.txt"})
	if exitCode != exitInvalidUsage {
		t.Errorf("run() = %d, want %d", exitCode, exitInvalidUsage)
	}
}

func TestRunInvalidMode(t *testing.T) {
	exitCode := run([]string{"apply", "--mode", "invalid", "--repo-path", "test.txt", "--new-file", "test.txt"})
	if exitCode != exitInvalidUsage {
		t.Errorf("run() = %d, want %d", exitCode, exitInvalidUsage)
	}
}

func TestRunMatchModeWithoutExpectSHA256(t *testing.T) {
	exitCode := run([]string{"apply", "--mode", "match", "--repo-path", "test.txt", "--new-file", "test.txt"})
	if exitCode != exitInvalidUsage {
		t.Errorf("run() = %d, want %d", exitCode, exitInvalidUsage)
	}
}

func TestRunNewFileNotExist(t *testing.T) {
	exitCode := run([]string{"apply", "--mode", "upsert", "--repo-path", "test.txt", "--new-file", "/nonexistent/file.txt"})
	if exitCode != exitOperational {
		t.Errorf("run() = %d, want %d", exitCode, exitOperational)
	}
}

func TestRunHelp(t *testing.T) {
	exitCode := run([]string{"-h"})
	if exitCode != exitSuccess {
		t.Errorf("run([-h]) = %d, want %d", exitCode, exitSuccess)
	}
}

func TestRunApplyHelp(t *testing.T) {
	exitCode := run([]string{"apply", "-h"})
	if exitCode != exitSuccess {
		t.Errorf("run([apply -h]) = %d, want %d", exitCode, exitSuccess)
	}
}
