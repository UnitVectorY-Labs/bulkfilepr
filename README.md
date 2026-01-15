[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT) [![Work In Progress](https://img.shields.io/badge/Status-Work%20In%20Progress-yellow)](https://guide.unitvectorylabs.com/bestpractices/status/#work-in-progress)

# bulkfilepr

Batch-update one or more standardized files across many local GitHub repositories, then commit the changes on a new branch and open pull requests back to each repository's default branch.

## Overview

bulkfilepr is a command-line tool designed to simplify the maintenance of standardized files (like GitHub workflows, CODEOWNERS, templates, Dockerfiles, and lint configs) across multiple repositories. Instead of manually opening each repo, creating branches, copying files, committing, pushing, and filing pull requests, bulkfilepr automates this entire process.

## Quick Start

```bash
# Upsert a workflow file (create or update)
bulkfilepr apply \
  --mode upsert \
  --repo-path .github/workflows/ci.yml \
  --new-file ~/standards/ci.yml

# Dry run to see what would change
bulkfilepr apply \
  --mode exists \
  --repo-path Dockerfile \
  --new-file ~/standards/Dockerfile \
  --dry-run
```

## Documentation

- [Installation](docs/INSTALL.md) - How to install bulkfilepr
- [Usage](docs/USAGE.md) - Detailed command-line reference
- [Examples](docs/EXAMPLES.md) - Usage examples and patterns

## Features

- **Three update modes**: `upsert` (always write), `exists` (update only if file exists), `match` (update only if file matches expected hash)
- **Idempotent operation**: If the target branch already exists, exits successfully (exit code 0) assuming previous successful run
- **Smart branch handling**: Automatically switches to default branch when on non-default branch with clean working tree
- **Safety checks**: Ensures you're on the default branch with a clean working tree before making changes
- **Dry run mode**: Preview changes without making any modifications
- **Automatic branching**: Creates deterministic branch names based on file content hash
- **GitHub CLI integration**: Automatically creates pull requests via `gh pr create`

## Requirements

- Git installed and configured
- GitHub CLI (`gh`) installed and authenticated

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
