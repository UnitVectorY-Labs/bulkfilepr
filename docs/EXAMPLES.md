---
layout: default
title: Examples
nav_order: 4
permalink: /examples
---

# Examples

This guide contains distinct, copy-paste-ready examples for common `bulkfilepr` workflows.

## Single-Repo Updates

### Create or update a file (`upsert`)

Use this when you want every target repo to end up with the file.

```bash
bulkfilepr apply \
  --mode upsert \
  --repo-path .github/workflows/ci.yml \
  --new-file ~/standards/ci.yml
```

### Update only if the file already exists (`exists`)

Use this when you do not want to create missing files.

```bash
bulkfilepr apply \
  --mode exists \
  --repo-path Dockerfile \
  --new-file ~/standards/Dockerfile
```

### Update only known baseline content (`match`)

Use this when you want to avoid overwriting custom modifications.

```bash
# Known baseline hashes (for example: v1 and v2)
V1_HASH="4b7c0e1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e"
V2_HASH="17ca04878ed554fc89bc73332e013fa8528c7999352a7cea17788e48fecabac6"

bulkfilepr apply \
  --mode match \
  --repo-path .github/workflows/ci.yml \
  --new-file ~/standards/ci-v3.yml \
  --expect-sha256 "$V1_HASH,$V2_HASH"
```

To calculate a file hash:

```bash
# macOS
shasum -a 256 .github/workflows/ci.yml

# Linux
sha256sum .github/workflows/ci.yml
```

## Safe Rollout Pattern (Many Repos)

### Phase 1: Audit with dry run

```bash
#!/usr/bin/env bash
set -euo pipefail

for repo_dir in ~/work/acme/*; do
  [ -d "$repo_dir/.git" ] || continue
  echo "=== $repo_dir ==="

  (
    cd "$repo_dir"
    bulkfilepr apply \
      --mode exists \
      --repo-path .github/workflows/ci.yml \
      --new-file ~/standards/ci.yml \
      --dry-run
  )
done
```

### Phase 2: Apply after review

```bash
#!/usr/bin/env bash
set -euo pipefail

for repo_dir in ~/work/acme/*; do
  [ -d "$repo_dir/.git" ] || continue

  (
    cd "$repo_dir"
    bulkfilepr apply \
      --mode exists \
      --repo-path .github/workflows/ci.yml \
      --new-file ~/standards/ci.yml
  )
done
```

## PR Metadata and Branch Control

### Set branch, commit, and PR metadata

```bash
bulkfilepr apply \
  --mode upsert \
  --repo-path .github/CODEOWNERS \
  --new-file ~/standards/CODEOWNERS \
  --branch chore/update-codeowners \
  --commit-message "chore: align CODEOWNERS with org standard" \
  --pr-title "Align CODEOWNERS with org standard" \
  --pr-body "Updates CODEOWNERS to the current standard used across repositories." \
  --draft
```

### Push to a non-default remote

```bash
bulkfilepr apply \
  --mode upsert \
  --repo-path .github/workflows/ci.yml \
  --new-file ~/standards/ci.yml \
  --remote upstream
```

## Common Preconditions

### Working tree must be clean

```bash
git status --short
# If needed:
git add -A && git commit -m "chore: checkpoint before bulkfilepr"
# or:
git stash push -u -m "temp before bulkfilepr"
```

### Run from the default branch

```bash
git fetch origin
git switch main   # or: git switch master, depending on the repo
```
