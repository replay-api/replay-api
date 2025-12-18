# Git Hooks

This directory contains git hooks for automatic code validation before commits.

## Installation

To install the hooks, run:

```bash
make install-hooks
```

Or manually:

```bash
chmod +x scripts/install-git-hooks.sh
./scripts/install-git-hooks.sh
```

## Available Hooks

### pre-commit

Runs the following validations before each commit:

1. **gofmt**: Checks code formatting
2. **go vet**: Checks for common code issues
3. **golangci-lint**: Runs linters (if available)
4. **goimports**: Checks for unused imports
5. **Tests**: Runs tests for modified packages

## Requirements

For the hooks to work completely, you need to have installed:

- `gofmt` (comes with Go)
- `goimports`: `go install golang.org/x/tools/cmd/goimports@latest`
- `golangci-lint` (optional, but recommended): `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`

## How It Works

When you run `git commit`, the `pre-commit` hook is automatically executed. If any validation fails, the commit will be blocked and you'll see an error message.

## Skip Hooks (Not Recommended)

If you need to skip hooks in an emergency (not recommended):

```bash
git commit --no-verify -m "message"
```

## Manual Testing

To test the hook without committing:

```bash
make hooks-test
```

Or run directly:

```bash
.githooks/pre-commit
```

## Customization

To add new validations, edit the `.githooks/pre-commit` file and add your validation commands.
