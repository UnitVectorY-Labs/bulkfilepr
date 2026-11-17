package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/UnitVectorY-Labs/bulkfilepr/internal/apply"
	"github.com/UnitVectorY-Labs/bulkfilepr/internal/config"
	"github.com/UnitVectorY-Labs/bulkfilepr/internal/git"
)

var Version = "dev" // Set by the build system to the release version

const (
	exitSuccess      = 0
	exitOperational  = 1
	exitInvalidUsage = 2
)

func main() {
	// Set the build version from the build info if not set by the build system
	if Version == "dev" || Version == "" {
		if bi, ok := debug.ReadBuildInfo(); ok {
			if bi.Main.Version != "" && bi.Main.Version != "(devel)" {
				Version = bi.Main.Version
			}
		}
	}

	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("bulkfilepr", flag.ContinueOnError)

	fs.Usage = func() {}
	fs.SetOutput(io.Discard)
	var expectSHA256 []string
	fs.Func("expect-sha256", "Allowed existing SHA-256; repeat or comma-separate", func(value string) error {
		expectSHA256 = append(expectSHA256, value)
		return nil
	})

	// Define flags
	var (
		mode          = fs.String("mode", "", "Update mode: upsert, exists, or match (required)")
		repoPath      = fs.String("repo-path", "", "Destination file path inside the repo (required)")
		newFile       = fs.String("new-file", "", "Path to the new file content (required)")
		repo          = fs.String("repo", ".", "Repository directory")
		branch        = fs.String("branch", "", "Branch name (auto-generated if empty)")
		commitMessage = fs.String("commit-message", "", "Commit message")
		prTitle       = fs.String("pr-title", "", "PR title")
		prBody        = fs.String("pr-body", "", "PR body")
		draft         = fs.Bool("draft", false, "Create PR as draft")
		dryRun        = fs.Bool("dry-run", false, "Perform checks only, no changes")
		remote        = fs.String("remote", "origin", "Git remote name")
		showVersion   = fs.Bool("version", false, "Print version")
	)

	// Check for subcommand
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Run bulkfilepr --help for usage.")
		return exitInvalidUsage
	}

	// Handle "apply" subcommand
	if args[0] == "apply" {
		args = args[1:]
	} else if args[0] == "-version" || args[0] == "--version" {
		fmt.Println(versionString())
		return exitSuccess
	} else if args[0] != "-h" && args[0] != "--help" && args[0] != "-help" {
		fmt.Fprintf(os.Stderr, "Error: unknown command %q\n", args[0])
		fmt.Fprintln(os.Stderr, "Run bulkfilepr --help for usage.")
		return exitInvalidUsage
	}

	// Parse flags
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			printUsage(fs)
			return exitSuccess
		}
		fmt.Fprintf(os.Stderr, "Error: %v; run bulkfilepr --help for usage.\n", err)
		return exitInvalidUsage
	}

	if fs.NArg() != 0 {
		fmt.Fprintf(os.Stderr, "Error: unexpected argument %q; put all options before positional arguments (none are accepted).\n", fs.Arg(0))
		return exitInvalidUsage
	}

	if *showVersion {
		fmt.Println(versionString())
		return exitSuccess
	}

	// Validate required flags
	if *mode == "" {
		fmt.Fprintln(os.Stderr, "Error: --mode is required")
		fmt.Fprintln(os.Stderr, "Run bulkfilepr --help for usage.")
		return exitInvalidUsage
	}

	if *repoPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --repo-path is required")
		fmt.Fprintln(os.Stderr, "Run bulkfilepr --help for usage.")
		return exitInvalidUsage
	}
	if *newFile == "" {
		fmt.Fprintln(os.Stderr, "Error: --new-file is required")
		fmt.Fprintln(os.Stderr, "Run bulkfilepr --help for usage.")
		return exitInvalidUsage
	}

	// Parse mode
	parsedMode, err := config.ParseMode(*mode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return exitInvalidUsage
	}

	// Build config
	cfg := &config.Config{
		Mode:          parsedMode,
		RepoPath:      *repoPath,
		NewFile:       *newFile,
		Repo:          *repo,
		Branch:        *branch,
		CommitMessage: *commitMessage,
		PRTitle:       *prTitle,
		PRBody:        *prBody,
		Draft:         *draft,
		DryRun:        *dryRun,
		Remote:        *remote,
		ExpectSHA256:  strings.Join(expectSHA256, ","),
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return exitInvalidUsage
	}

	// Read new file content
	newContent, err := os.ReadFile(*newFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to read new file: %v\n", err)
		return exitOperational
	}

	// Create git operations
	gitOps := git.NewRealOperations(*repo)

	// Create and run applier
	applier := apply.NewApplier(cfg, gitOps, newContent)
	result, err := applier.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return exitOperational
	}

	// Print result
	printResult(cfg, result)
	return exitSuccess
}

var semverRe = regexp.MustCompile(`^\d+\.\d+\.\d+`)

func versionString() string {
	version := Version
	if semverRe.MatchString(version) && !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	return fmt.Sprintf("bulkfilepr version %s (%s, %s/%s)", version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

func printUsage(fs *flag.FlagSet) {
	fmt.Fprintln(os.Stderr, `Usage: bulkfilepr apply --mode <mode> --repo-path <path> --new-file <path> [options]
       bulkfilepr --help | --version

Update one file in one local GitHub repository, commit, push, and open a PR.
Run once per repository for bulk updates. Requires git, authenticated gh (gh auth
login), a clean working tree, and push/PR permissions. Fetch refs beforehand:
branch existence checks use local and cached remote refs; no fetch/pull is done.

Required:
  --mode <mode>          upsert: create or update; exists: update only if present;
                         match: update only if existing bytes match any allowed hash
  --repo-path <path>     Destination relative to repository root
  --new-file <path>      Replacement content; relative to process working directory

Options:
  --repo <dir>           Local repository directory (default: .)
  --expect-sha256 <hex>  Required for match; SHA-256 of EXISTING content, lowercase hex
                         Repeat this flag or comma-separate hashes (any may match)
  --branch <name>        Default: bulkfilepr/<first 12 hex of replacement SHA-256>
  --commit-message <msg> Default: chore: update <repo-path>
  --pr-title <title>     Default: Update <repo-path>
  --pr-body <body>       Default: This PR updates the standardized file at <repo-path>.
  --draft               Create a draft PR
  --remote <name>        Push remote (default: origin)
  --dry-run             Check and report without writing, committing, pushing, or PR
                         creation; may switch a clean checkout to the default branch
  --help, -h            Show this help (also: bulkfilepr apply --help)
  --version             Print version/build info

Behavior:
  Uses the default branch reported by gh; switches to it if the checkout is clean.
  Identical content, unmet mode conditions, or an existing branch are successful
  no-ops. An existing branch is assumed complete; its PR is not verified.
  After updating, switches back to the default branch (best effort).

Examples:
  bulkfilepr apply --repo ./my-repo --mode upsert --repo-path .github/CODEOWNERS --new-file ./CODEOWNERS --dry-run
  bulkfilepr apply --repo ./my-repo --mode match --repo-path config.yml --new-file ./config.yml --expect-sha256 <old-hash-1> --expect-sha256 <old-hash-2>
  Replace <old-hash-N> with full 64-character hashes; compute with sha256sum FILE
  or shasum -a 256 FILE. Remove --dry-run to commit, push, and create a PR.

Output: stdout reports Action, Reason (for no-ops), Branch, and PR URL as applicable.
Errors go to stderr. Exit codes: 0 success (including no-op/dry-run/help/version),
1 operational failure (file, repository, git/gh, push, or PR), 2 invalid usage.`)
}

func printResult(cfg *config.Config, result *apply.Result) {
	fmt.Printf("Default branch: %s\n", result.DefaultBranch)

	switch result.Action {
	case "no action taken":
		fmt.Printf("Mode: %s\n", cfg.Mode)
		fmt.Printf("Action: no action taken\n")
		fmt.Printf("Reason: %s\n", result.NoActionReason)
	case "would update":
		fmt.Printf("Mode: %s\n", cfg.Mode)
		fmt.Printf("Action: would update (dry run)\n")
		fmt.Printf("Branch: %s\n", result.BranchName)
	case "branch already exists":
		fmt.Printf("Mode: %s\n", cfg.Mode)
		fmt.Printf("Action: branch already exists (idempotent - no action taken)\n")
		fmt.Printf("Branch: %s\n", result.BranchName)
		fmt.Printf("Reason: branch already exists, assuming previous successful run\n")
	case "updated":
		fmt.Printf("Mode: %s\n", cfg.Mode)
		fmt.Printf("Action: updated\n")
		fmt.Printf("Branch: %s\n", result.BranchName)
		fmt.Printf("PR URL: %s\n", result.PRURL)
	}
}
