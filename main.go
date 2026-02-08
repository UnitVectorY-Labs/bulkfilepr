package main

import (
	"flag"
	"fmt"
	"os"

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
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("bulkfilepr", flag.ContinueOnError)

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
		expectSHA256  = fs.String("expect-sha256", "", "Expected SHA-256 hash (required for match mode)")
		showVersion   = fs.Bool("version", false, "Print version")
	)

	// Check for subcommand
	if len(args) == 0 {
		printUsage(fs)
		return exitInvalidUsage
	}

	// Handle "apply" subcommand
	if args[0] == "apply" {
		args = args[1:]
	} else if args[0] == "-version" || args[0] == "--version" {
		fmt.Printf("bulkfilepr version %s\n", Version)
		return exitSuccess
	} else if args[0] != "-h" && args[0] != "--help" && args[0] != "-help" {
		fmt.Fprintf(os.Stderr, "Error: unknown command %q\n", args[0])
		fmt.Fprintf(os.Stderr, "Usage: bulkfilepr apply [options]\n")
		return exitInvalidUsage
	}

	// Parse flags
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return exitSuccess
		}
		return exitInvalidUsage
	}

	if *showVersion {
		fmt.Printf("bulkfilepr version %s\n", Version)
		return exitSuccess
	}

	// Validate required flags
	if *mode == "" {
		fmt.Fprintln(os.Stderr, "Error: --mode is required")
		printUsage(fs)
		return exitInvalidUsage
	}
	if *repoPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --repo-path is required")
		printUsage(fs)
		return exitInvalidUsage
	}
	if *newFile == "" {
		fmt.Fprintln(os.Stderr, "Error: --new-file is required")
		printUsage(fs)
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
		ExpectSHA256:  *expectSHA256,
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

func printUsage(fs *flag.FlagSet) {
	fmt.Fprintln(os.Stderr, "Usage: bulkfilepr apply [options]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Batch-update standardized files across repositories.")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Required options:")
	fmt.Fprintln(os.Stderr, "  --mode <mode>         Update mode: upsert, exists, or match")
	fmt.Fprintln(os.Stderr, "  --repo-path <path>    Destination file path inside the repo")
	fmt.Fprintln(os.Stderr, "  --new-file <path>     Path to the new file content")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Optional options:")
	fmt.Fprintln(os.Stderr, "  --repo <dir>          Repository directory (default: .)")
	fmt.Fprintln(os.Stderr, "  --branch <name>       Branch name (auto-generated if empty)")
	fmt.Fprintln(os.Stderr, "  --commit-message <msg> Commit message")
	fmt.Fprintln(os.Stderr, "  --pr-title <title>    PR title")
	fmt.Fprintln(os.Stderr, "  --pr-body <body>      PR body")
	fmt.Fprintln(os.Stderr, "  --draft               Create PR as draft")
	fmt.Fprintln(os.Stderr, "  --dry-run             Perform checks only, no changes")
	fmt.Fprintln(os.Stderr, "  --remote <name>       Git remote name (default: origin)")
	fmt.Fprintln(os.Stderr, "  --expect-sha256 <hex> Expected SHA-256 (required for match mode)")
	fmt.Fprintln(os.Stderr, "                        Multiple hashes can be comma-separated")
	fmt.Fprintln(os.Stderr, "  --version             Print version")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Modes:")
	fmt.Fprintln(os.Stderr, "  upsert  - Always write (create if missing, update if exists)")
	fmt.Fprintln(os.Stderr, "  exists  - Only update if file already exists")
	fmt.Fprintln(os.Stderr, "  match   - Only update if file exists and matches expected hash")
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

