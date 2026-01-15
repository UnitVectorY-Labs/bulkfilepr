# Installing bulkfilepr

## Prerequisites

Before installing bulkfilepr, ensure you have:

1. **Git** installed and configured
2. **GitHub CLI (`gh`)** installed and authenticated

### Installing GitHub CLI

If you don't have the GitHub CLI installed:

**macOS (using Homebrew):**
```bash
brew install gh
```

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install gh
```

**Windows (using winget):**
```bash
winget install --id GitHub.cli
```

**Windows (using Chocolatey):**
```bash
choco install gh
```

After installing, authenticate with GitHub:
```bash
gh auth login
```

## Installing bulkfilepr

### From Source (Go)

If you have Go 1.21 or later installed:

```bash
go install github.com/UnitVectorY-Labs/bulkfilepr@latest
```

This will install the `bulkfilepr` binary to your `$GOPATH/bin` directory.

### From Releases

Download the latest release binary for your platform from the [releases page](https://github.com/UnitVectorY-Labs/bulkfilepr/releases).

**macOS (Intel):**
```bash
curl -LO https://github.com/UnitVectorY-Labs/bulkfilepr/releases/latest/download/bulkfilepr-darwin-amd64
chmod +x bulkfilepr-darwin-amd64
sudo mv bulkfilepr-darwin-amd64 /usr/local/bin/bulkfilepr
```

**macOS (Apple Silicon):**
```bash
curl -LO https://github.com/UnitVectorY-Labs/bulkfilepr/releases/latest/download/bulkfilepr-darwin-arm64
chmod +x bulkfilepr-darwin-arm64
sudo mv bulkfilepr-darwin-arm64 /usr/local/bin/bulkfilepr
```

**Linux (x86_64):**
```bash
curl -LO https://github.com/UnitVectorY-Labs/bulkfilepr/releases/latest/download/bulkfilepr-linux-amd64
chmod +x bulkfilepr-linux-amd64
sudo mv bulkfilepr-linux-amd64 /usr/local/bin/bulkfilepr
```

**Windows:**

Download `bulkfilepr-windows-amd64.exe` from the releases page and add it to your PATH.

### Building from Source

```bash
# Clone the repository
git clone https://github.com/UnitVectorY-Labs/bulkfilepr.git
cd bulkfilepr

# Build
go build -o bulkfilepr .

# Optionally install to your PATH
sudo mv bulkfilepr /usr/local/bin/
```

## Verifying Installation

After installation, verify everything is working:

```bash
# Check bulkfilepr version
bulkfilepr --version

# Check that gh is authenticated
gh auth status

# Check git is available
git --version
```

## Upgrading

To upgrade to the latest version:

**If installed via `go install`:**
```bash
go install github.com/UnitVectorY-Labs/bulkfilepr@latest
```

**If installed from release binary:**
Download and replace with the latest binary from the releases page.
