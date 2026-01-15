# bulkfilepr Usage Guide

## Command Structure

```
bulkfilepr apply [options]
```

bulkfilepr uses a single `apply` command with various options to control its behavior.

## Required Options

### `--mode <mode>`

Specifies the update mode. Must be one of:

- **`upsert`**: Always write the new file (create if missing, update if exists). If the resulting content is identical to what is already there, no action is taken.

- **`exists`**: Only update if the file already exists at the destination path. If missing, no action is taken. If exists but content matches, no action is taken.

- **`match`**: Only update if the file exists AND its SHA-256 hash matches the `--expect-sha256` value. This is useful for ensuring you only update files that you expect to be at a specific version.

### `--repo-path <path>`

The destination file path inside the repository, relative to the repository root.

Examples:
- `.github/workflows/ci.yml`
- `Dockerfile`
- `.eslintrc.json`
- `CODEOWNERS`

### `--new-file <path>`

Path on disk to the new file content that will be written to `--repo-path`.

## Optional Options

### `--repo <dir>`

Repository directory to operate on. Defaults to `.` (current directory).

### `--branch <name>`

Branch name to create for the changes. If omitted, a branch name is automatically generated using the pattern `bulkfilepr/{hash}` where `{hash}` is the first 12 characters of the SHA-256 hash of the new file content.

### `--commit-message <msg>`

Commit message for the change. Defaults to `chore: update {repo-path}`.

### `--pr-title <title>`

Title for the pull request. Defaults to `Update {repo-path}`.

### `--pr-body <body>`

Body content for the pull request. Defaults to a short sentence referencing the standard file update.

### `--draft`

If set, create the PR as a draft PR.

### `--dry-run`

**Critical option.** Performs all safety checks and reports what would change, but makes no actual changes:
- Does not switch or create branches
- Does not write files
- Does not stage, commit, or push
- Does not create PRs

Dry run exits with code 0 if safety checks pass (regardless of whether it would take action), or non-zero if safety checks fail.

### `--remote <name>`

Git remote name to push to. Defaults to `origin`.

### `--expect-sha256 <hex>`

Required when `--mode match` is specified. The SHA-256 hash that the existing file at `--repo-path` must match for the update to proceed.

### `--version`

Print the version number and exit.

## Safety Checks

Before making any changes (in both normal and dry-run modes), bulkfilepr performs the following safety checks:

1. **Default Branch Detection**: Uses `gh repo view` to determine the repository's default branch.

2. **On Default Branch**: Verifies that the current checked-out branch is the default branch. This prevents accidentally branching off a feature branch.

3. **Clean Working Tree**: Runs `git status --porcelain` to ensure there are no uncommitted changes. This prevents accidentally including unrelated changes in the PR.

If any of these checks fail, bulkfilepr exits with a non-zero exit code.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success, including "no action taken" |
| 1 | Operational failure (unsafe repo state, git/gh failure, push/PR failure) |
| 2 | Invalid usage (missing required flags, invalid mode, missing `--expect-sha256` for match mode) |

## Output

bulkfilepr outputs human-readable information including:
- Default branch name
- Mode decision and reason (updated vs no-op vs condition not met)
- In dry run: "would update" plus planned branch name
- In non dry run: PR URL on success
