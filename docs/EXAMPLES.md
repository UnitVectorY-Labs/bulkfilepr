# bulkfilepr Examples

This guide provides practical examples of using bulkfilepr for common scenarios.

## Basic Examples

### Upsert a Workflow File

Create or update a CI workflow file:

```bash
bulkfilepr apply \
  --mode upsert \
  --repo-path .github/workflows/ci.yml \
  --new-file ~/standards/ci.yml
```

### Update Only If File Exists

Update a Dockerfile only if one already exists in the repo:

```bash
bulkfilepr apply \
  --mode exists \
  --repo-path Dockerfile \
  --new-file ~/standards/Dockerfile
```

### Update Only If Hash Matches

Update a release workflow only if it matches a specific version:

```bash
bulkfilepr apply \
  --mode match \
  --repo-path .github/workflows/release.yml \
  --new-file ~/standards/release.yml \
  --expect-sha256 4b7c0e1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e
```

### Update If Hash Matches Any of Multiple Versions

Update a file if it matches any of several known baseline versions:

```bash
bulkfilepr apply \
  --mode match \
  --repo-path .github/workflows/ci.yml \
  --new-file ~/standards/ci-v3.yml \
  --expect-sha256 17ca04878ed554fc89bc73332e013fa8528c7999352a7cea17788e48fecabac6,6bbb6e1ef2fbd220c4dc6853dc40d80e1d060b32f3dfae245f2f4dc8858ccfa1
```

This is useful when you want to upgrade from multiple previous versions (e.g., v1 or v2) to a new version (v3).

## Dry Run Examples

### Preview Changes Without Making Them

Always use `--dry-run` first when updating multiple repos:

```bash
bulkfilepr apply \
  --mode upsert \
  --repo-path .github/workflows/ci.yml \
  --new-file ~/standards/ci.yml \
  --dry-run
```

Output will show:
```
Default branch: main
Mode: upsert
Action: would update (dry run)
Branch: bulkfilepr/a1b2c3d4e5f6
```

### Audit Multiple Repos

Check what would change across all repos in a directory:

```bash
for d in ~/work/acme/*; do
  [ -d "$d/.git" ] || continue
  echo "=== $d ==="
  (cd "$d" && bulkfilepr apply \
    --mode exists \
    --repo-path .github/workflows/ci.yml \
    --new-file ~/standards/ci.yml \
    --dry-run)
  echo ""
done
```

## Customization Examples

### Custom Branch Name

Specify a custom branch name instead of auto-generated:

```bash
bulkfilepr apply \
  --mode upsert \
  --repo-path .github/workflows/ci.yml \
  --new-file ~/standards/ci.yml \
  --branch update-ci-workflow-2024
```

### Custom Commit Message and PR Title

```bash
bulkfilepr apply \
  --mode upsert \
  --repo-path .github/workflows/ci.yml \
  --new-file ~/standards/ci.yml \
  --commit-message "chore: upgrade CI to v2 standards" \
  --pr-title "Upgrade CI workflow to v2 standards" \
  --pr-body "This PR updates the CI workflow to match the organization's v2 standard.

See INFRA-1234 for details."
```

### Create Draft PR

Create a draft PR for review before publishing:

```bash
bulkfilepr apply \
  --mode upsert \
  --repo-path .github/CODEOWNERS \
  --new-file ~/standards/CODEOWNERS \
  --draft
```

## Batch Operations

### Update Multiple Repos

Loop through multiple repositories and apply changes:

```bash
#!/bin/bash
STANDARD_FILE=~/standards/ci.yml
TARGET_PATH=.github/workflows/ci.yml

for repo_dir in ~/work/acme/*; do
  [ -d "$repo_dir/.git" ] || continue
  
  echo "Processing: $repo_dir"
  cd "$repo_dir"
  
  bulkfilepr apply \
    --mode exists \
    --repo-path "$TARGET_PATH" \
    --new-file "$STANDARD_FILE"
  
  echo ""
done
```

### Two-Phase Approach: Audit Then Apply

First, audit all repos to see what would change:

```bash
#!/bin/bash
# Phase 1: Audit
for repo_dir in ~/work/acme/*; do
  [ -d "$repo_dir/.git" ] || continue
  echo "=== $repo_dir ==="
  (cd "$repo_dir" && bulkfilepr apply \
    --mode exists \
    --repo-path .github/workflows/ci.yml \
    --new-file ~/standards/ci.yml \
    --dry-run)
done > audit-results.txt
```

Review `audit-results.txt`, then apply:

```bash
#!/bin/bash
# Phase 2: Apply (after reviewing audit results)
for repo_dir in ~/work/acme/*; do
  [ -d "$repo_dir/.git" ] || continue
  (cd "$repo_dir" && bulkfilepr apply \
    --mode exists \
    --repo-path .github/workflows/ci.yml \
    --new-file ~/standards/ci.yml)
done
```

## Match Mode with SHA-256

### Getting the SHA-256 of a File

To use `--mode match`, you need the SHA-256 hash of the current file:

```bash
# On macOS
shasum -a 256 .github/workflows/ci.yml

# On Linux
sha256sum .github/workflows/ci.yml
```

### Rolling Update with Hash Verification

Update only repos that have a specific version of the file:

```bash
# Old file hash (get this from your current standard)
OLD_HASH="4b7c0e1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e"

bulkfilepr apply \
  --mode match \
  --repo-path .github/workflows/ci.yml \
  --new-file ~/standards/ci-v2.yml \
  --expect-sha256 "$OLD_HASH"
```

This ensures only repos with the v1 workflow get upgraded to v2, leaving repos with custom modifications untouched.

### Rolling Update from Multiple Baseline Versions

Update repos that have any of several known baseline versions:

```bash
# v1 and v2 hashes
V1_HASH="4b7c0e1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e"
V2_HASH="17ca04878ed554fc89bc73332e013fa8528c7999352a7cea17788e48fecabac6"

bulkfilepr apply \
  --mode match \
  --repo-path .github/workflows/ci.yml \
  --new-file ~/standards/ci-v3.yml \
  --expect-sha256 "$V1_HASH,$V2_HASH"
```

This upgrades repos that have either v1 or v2 to v3, while leaving customized versions untouched.

## Working with Different Remotes

### Push to a Different Remote

If you need to push to a remote other than `origin`:

```bash
bulkfilepr apply \
  --mode upsert \
  --repo-path .github/workflows/ci.yml \
  --new-file ~/standards/ci.yml \
  --remote upstream
```

## Error Handling

### Handling Non-Clean Working Tree

If bulkfilepr reports "working tree is not clean", you have uncommitted changes:

```bash
# Option 1: Commit your changes first
git add .
git commit -m "WIP: current work"

# Option 2: Stash your changes
git stash
bulkfilepr apply ...
git stash pop
```

### Handling Wrong Branch

If bulkfilepr reports "not on default branch":

```bash
# Switch to the default branch first
git checkout main
bulkfilepr apply ...
```
