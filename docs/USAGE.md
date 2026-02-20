---
layout: default
title: Usage
nav_order: 3
permalink: /usage
---

# Usage

## Command Structure

```
bulkfilepr apply [options]
```

bulkfilepr uses a single `apply` command with various options to control its behavior.

## Command-Line Options

| Option | Argument | Required | Notes |
|--------|----------|----------|-------|
| `--mode` | `<mode>` | Yes | Update mode: `upsert`, `exists`, or `match` |
| `--repo-path` | `<path>` | Yes | Destination file path inside the repository (relative to repo root) |
| `--new-file` | `<path>` | Yes | Path on disk to the new file content |
| `--repo` | `<dir>` | No | Repository directory to operate on (default: `.`) |
| `--branch` | `<name>` | No | Branch name for the changes (auto-generated if omitted) |
| `--commit-message` | `<msg>` | No | Commit message (default: `chore: update {repo-path}`) |
| `--pr-title` | `<title>` | No | Pull request title (default: `Update {repo-path}`) |
| `--pr-body` | `<body>` | No | Pull request body content |
| `--draft` | - | No | Create the PR as a draft |
| `--dry-run` | - | No | Perform checks only, make no actual changes |
| `--remote` | `<name>` | No | Git remote name to push to (default: `origin`) |
| `--expect-sha256` | `<hex>` | Conditional | Expected SHA-256 hash (required when `--mode match`). Multiple hashes can be comma-separated to match any of them |
| `--version` | - | No | Print version number and exit |

## Update Modes

The `--mode` option controls when files are updated:

### `upsert`
Always write the new file (create if missing, update if exists). If the file already exists and the content is identical to the new content, no action is taken.

**Use case**: Ensuring a standard file is present and up-to-date across all repositories.

### `exists`
Only update if the file already exists at the destination path. If the file is missing, no action is taken. If the file exists but content matches the new content, no action is taken.

**Use case**: Updating existing configuration files without creating them in repositories that don't have them.

### `match`
Only update if the file exists AND its SHA-256 hash matches one of the values provided in `--expect-sha256`. This ensures you only update files at specific known versions. Multiple comma-separated hashes can be provided to match any of several versions.

**Use case**: Safely updating files when you need to verify they haven't been customized from known baselines.

**Examples**:
- `.github/workflows/ci.yml`
- `Dockerfile`
- `.eslintrc.json`
- `CODEOWNERS`

## Branch Naming

When `--branch` is not specified, bulkfilepr automatically generates a branch name using the pattern:

```
bulkfilepr/{hash}
```

where `{hash}` is the first 12 characters of the SHA-256 hash of the new file content. This ensures:
- Deterministic branch names for identical content
- Different branches for different file versions
- Easy identification of bulkfilepr-managed branches

## Idempotency and Branch Existence

**bulkfilepr is designed to be idempotent.** If the branch to be created already exists (either locally or on the remote), the command exits successfully with exit code 0. This allows the command to be run multiple times safely without creating duplicate branches or PRs.

**Behavior when branch exists**:
- The tool checks if the branch exists before attempting to create it
- If found, it assumes a previous successful run and exits with code 0
- Output indicates: `Action: branch already exists (idempotent - no action taken)`
- No git operations are performed

This makes bulkfilepr safe to use in automation and retry scenarios.

## Branch State Handling

bulkfilepr has intelligent handling for repository branch states:

### On Default Branch
If you're already on the default branch with a clean working tree, bulkfilepr proceeds normally.

### On Non-Default Branch (Clean Working Tree)
If you're on a non-default branch but your working tree is clean (no uncommitted changes), bulkfilepr will automatically switch to the default branch and proceed with the operation.

**Example**:
```bash
# Currently on 'feature-branch' with no uncommitted changes
bulkfilepr apply --mode upsert --repo-path README.md --new-file ~/standard/README.md
# → Switches to default branch (e.g., 'main') and proceeds
```

### On Non-Default Branch (Dirty Working Tree)
If you're on a non-default branch AND have uncommitted changes, bulkfilepr exits with a non-zero exit code to prevent data loss.

**Error message**:
```
Error: not on default branch and working tree is dirty: current branch is "feature-branch", expected "main". Please commit or stash your changes
```

This safety check ensures you don't accidentally lose uncommitted work.

## Safety Checks

Before making any changes (in both normal and dry-run modes), bulkfilepr performs the following safety checks:

1. **Default Branch Detection**: Uses `gh repo view` to determine the repository's default branch name.

2. **Branch State Verification**: 
   - If on the default branch: proceeds to next check
   - If on non-default branch with clean working tree: switches to default branch
   - If on non-default branch with dirty working tree: exits with error

3. **Clean Working Tree**: Ensures there are no uncommitted changes after any branch switching. This prevents accidentally including unrelated changes in the PR.

4. **Branch Existence Check**: Verifies the target branch doesn't already exist (for idempotency).

If any of these checks fail, bulkfilepr exits with a non-zero exit code.

## Dry Run Mode

The `--dry-run` flag is a critical safety feature that performs all checks and reports what would happen, but makes no actual changes:

**What dry run does**:
- ✅ Detects default branch
- ✅ Checks current branch state
- ✅ Verifies working tree cleanliness
- ✅ Evaluates mode conditions (file existence, content matching)
- ✅ Determines branch name
- ✅ Reports what would be updated

**What dry run does NOT do**:
- ❌ Does not switch branches
- ❌ Does not create branches
- ❌ Does not write files
- ❌ Does not stage, commit, or push
- ❌ Does not create PRs

**Exit behavior**: Dry run exits with code 0 if safety checks pass (regardless of whether it would take action), or non-zero if safety checks fail.

## Mode Interactions with `--expect-sha256`

The `--expect-sha256` option has specific interactions with update modes:

| Mode | `--expect-sha256` | Behavior |
|------|-------------------|----------|
| `upsert` | Optional (ignored) | Always attempts to write file |
| `exists` | Optional (ignored) | Only updates if file exists |
| `match` | **Required** | Only updates if file exists AND hash matches |

When using `--mode match`, you must provide `--expect-sha256` or the command will exit with error code 2 (invalid usage).

**Example**:
```bash
# This will fail - missing --expect-sha256
bulkfilepr apply --mode match --repo-path config.yml --new-file ~/config.yml

# This works - hash provided
bulkfilepr apply --mode match --repo-path config.yml --new-file ~/config.yml \
  --expect-sha256 a1b2c3d4e5f6...
```

## Exit Codes

| Code | Meaning | Examples |
|------|---------|----------|
| 0 | Success | File updated, no action needed, branch already exists, dry run passed |
| 1 | Operational failure | Unsafe repo state, git/gh command failure, push/PR failure |
| 2 | Invalid usage | Missing required flags, invalid mode, missing `--expect-sha256` for match mode |

**Note**: Exit code 0 is returned for several "success" scenarios:
- File was successfully updated and PR created
- No action was needed (content already matches)
- Branch already exists (idempotent behavior)
- Mode conditions not met (e.g., file doesn't exist with `--mode exists`)

## Output Format

bulkfilepr provides clear, human-readable output for all operations:

```
Default branch: main
Mode: upsert
Action: updated
Branch: bulkfilepr/a1b2c3d4e5f6
PR URL: https://github.com/owner/repo/pull/123
```

**Possible actions**:
- `updated` - File was updated and PR created
- `no action taken` - Mode conditions not met or content already matches
- `would update (dry run)` - Dry run mode, would have updated
- `branch already exists (idempotent - no action taken)` - Branch exists, assuming previous success

Each action includes relevant context like branch name, reason for no action, or PR URL.
