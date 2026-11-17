package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/UnitVectorY-Labs/bulkfilepr/internal/hash"
	"strings"
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

func TestVersionStringWithPrefixedVersion(t *testing.T) {
	originalVersion := Version
	t.Cleanup(func() { Version = originalVersion })
	Version = "v1.2.3"

	got := versionString()
	want := "bulkfilepr version v1.2.3 (" + runtime.Version() + ", " + runtime.GOOS + "/" + runtime.GOARCH + ")"

	if got != want {
		t.Errorf("versionString() = %q, want %q", got, want)
	}
}

func TestVersionStringAddsVPrefixWhenMissing(t *testing.T) {
	originalVersion := Version
	t.Cleanup(func() { Version = originalVersion })
	Version = "1.2.3"

	got := versionString()

	if !strings.HasPrefix(got, "bulkfilepr version v1.2.3 ") {
		t.Errorf("versionString() = %q, want prefix %q", got, "bulkfilepr version v1.2.3 ")
	}
}

// Capture stderr without parallel tests: run uses the process streams.
func captureErrorOutput(t *testing.T, args []string) (int, string) {
	t.Helper()
	old := os.Stderr
	f, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = f
	defer func() { os.Stderr = old; f.Close() }()
	code := run(args)
	if _, err := f.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	output, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	return code, string(output)
}

func TestHelpAndUsageErrors(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"apply", "--help"}, {"-h"}} {
		code, output := captureErrorOutput(t, args)
		if code != exitSuccess {
			t.Fatalf("help exit = %d", code)
		}
		for _, want := range []string{"Examples:", "Exit codes:", "authenticated gh", "Repeat this flag", "may switch", "--new-file"} {
			if !strings.Contains(output, want) {
				t.Errorf("help missing %q", want)
			}
		}
	}
	for _, args := range [][]string{{"apply"}, {"apply", "--bogus"}, {"apply", "--mode"}, {"apply", "unexpected"}} {
		code, output := captureErrorOutput(t, args)
		if code != exitInvalidUsage || !strings.Contains(output, "Error:") || strings.Contains(output, "Examples:") {
			t.Errorf("%v: code=%d output=%q", args, code, output)
		}
	}
}

func TestRepeatedExpectedHashes(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture")
	}
	repo := t.TempDir()
	gitRun := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git: %v: %s", err, out)
		}
	}
	gitRun("init", "-b", "main")
	if err := os.WriteFile(filepath.Join(repo, "config.txt"), []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	gitRun("add", ".")
	gitRun("-c", "user.name=Test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false", "commit", "-m", "initial")
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "gh"), []byte("#!/bin/sh\necho main\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	replacement := filepath.Join(t.TempDir(), "new.txt")
	if err := os.WriteFile(replacement, []byte("new"), 0600); err != nil {
		t.Fatal(err)
	}
	// A matching value must survive whether it appears in the first or last flag.
	matching := hash.SHA256Bytes([]byte("old"))
	other := hash.SHA256Bytes([]byte("other"))
	for _, values := range [][]string{{matching, other}, {other, matching}, {other + "," + matching, other}} {
		args := []string{"apply", "--repo", repo, "--mode", "match", "--repo-path", "config.txt", "--new-file", replacement, "--dry-run"}
		for _, value := range values {
			args = append(args, "--expect-sha256", value)
		}
		old := os.Stdout
		f, err := os.CreateTemp(t.TempDir(), "stdout")
		if err != nil {
			t.Fatal(err)
		}
		os.Stdout = f
		code := run(args)
		os.Stdout = old
		f.Seek(0, 0)
		output, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			t.Fatal(err)
		}
		if code != exitSuccess || !strings.Contains(string(output), "Action: would update") {
			t.Fatalf("code=%d output=%s", code, output)
		}
	}
}
