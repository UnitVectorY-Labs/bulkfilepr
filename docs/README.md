---
layout: default
title: bulkfilepr
nav_order: 1
permalink: /
---

# bulkfilepr

Welcome to the bulkfilepr documentation. bulkfilepr is a command-line tool for batch-updating standardized files across many local GitHub repositories.

## What is bulkfilepr?

bulkfilepr makes a very specific kind of maintenance work fast, safe, and repeatable: rolling out the exact same file-level change across lots of repositories without manually opening each repo, creating branches, copying files, committing, pushing, and filing pull requests.

It targets standardized "shared surface area" files like:
- GitHub workflows (`.github/workflows/*.yml`)
- CODEOWNERS files
- Issue and PR templates
- Dockerfiles
- Lint configurations (`.eslintrc`, `.golangci.yml`, etc.)
- Other standardized artifacts where consistency matters

## Key Features

- **Conservative by design**: The tool discovers the repository's default branch, refuses to operate on anything but a clean working tree of that default branch, and supports modes that allow updates only when a file exists or when it matches an expected fingerprint.

- **Dry run support**: Can be used as a "what would change" audit step before making actual changes.

- **Clean PR workflow**: Produces a clean, reviewable PR using the same git and GitHub CLI flow you would do by hand.

## Documentation

- [Installation](INSTALL.md) - How to install bulkfilepr on various platforms
- [Usage](USAGE.md) - Detailed command-line reference and mode explanations
- [Examples](EXAMPLES.md) - Common usage patterns and examples

## Quick Example

```bash
# Update a CI workflow file across a repo
bulkfilepr apply \
  --mode upsert \
  --repo-path .github/workflows/ci.yml \
  --new-file ~/standards/ci.yml
```

This will:
1. Check that you're on the default branch with a clean working tree
2. Create a new branch (e.g., `bulkfilepr/a1b2c3d4e5f6`)
3. Copy the new file content to `.github/workflows/ci.yml`
4. Commit the changes
5. Push to origin
6. Create a pull request
7. Switch back to the default branch
