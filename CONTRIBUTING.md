# Contributing Guide

Thank you for considering contributing to LeetGaming Replay API! This document provides guidelines for contributing to the project.

## How Can I Contribute?

### Reporting Bugs

Before creating a bug report, please check if the issue has already been reported. If you can't find an existing issue, create a new one following the bug report template.

**How to write a good bug report:**
- Use a clear and descriptive title
- Describe the steps to reproduce the problem
- Describe the expected behavior
- Describe the current behavior
- Include screenshots if applicable
- Include information about your environment (OS, Go version, etc.)

### Suggesting Enhancements

Suggestions are always welcome! When creating an issue for a suggestion:

- Use a clear and descriptive title
- Provide a detailed description of the suggested functionality
- Explain why this functionality would be useful
- Include examples of how the functionality would be used

### Pull Requests

1. **Fork the repository**
2. **Create a branch** from `develop`:
   ```bash
   git checkout develop
   git pull origin develop
   git checkout -b feature/your-feature-name
   ```

3. **Make your changes**
   - Follow the project's code standards
   - Add tests for new functionality
   - Ensure all tests pass
   - Update documentation as needed

4. **Commit your changes**
   - Follow the [commit style guide](./COMMIT_STYLE.md)
   - Use descriptive commit messages
   - Small, focused commits are preferred

5. **Push to your branch**
   ```bash
   git push origin feature/your-feature-name
   ```

6. **Open a Pull Request**
   - Fill out the PR template
   - Link the corresponding issue (if any)
   - Wait for review

## Code Standards

### Go Style Guide

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` for formatting (runs automatically)
- Run `golangci-lint` before committing
- Keep functions small and focused
- Use descriptive names for variables and functions

### Architecture

This project follows **Hexagonal Architecture (Ports & Adapters)**:

- **Domain Layer** (`pkg/domain/`): Contains pure business logic
- **Infrastructure Layer** (`pkg/infra/`): Concrete implementations (DB, Kafka, etc.)
- **Application Layer** (`cmd/`): Entry points and controllers

**Important rules:**
- Controllers MUST NEVER access repositories directly
- Use use cases for all business logic
- Dependencies point inward (infra → domain, never the reverse)

See [AGENTS.md](./AGENTS.md) for more architecture details.

### Testing

- **DO NOT use mocks** - use real instances (MongoDB, Kafka, etc.)
- Integration tests should use `//go:build integration`
- Run `go test ./...` before committing
- Maintain high test coverage

See [TESTING.md](./TESTING.md) for more details.

### Documentation

- Document public functions
- Update README.md if necessary
- Add usage examples when relevant
- Keep architecture documentation up to date

## Review Process

1. **Code Review**: At least 1 reviewer must approve
2. **CI/CD**: All tests must pass
3. **Linting**: `golangci-lint` must pass without errors
4. **Merge**: After approval, PR will be merged into `develop`

## Development Environment

### Prerequisites

- Go 1.23+
- Docker Desktop
- Make (optional)

### Local Setup

```bash
# Clone the repository
git clone https://github.com/leetgaming-pro/replay-api.git
cd replay-api

# Install dependencies
go mod download

# Start test infrastructure
docker compose -f docker-compose.test.yml up -d

# Run tests
go test ./...

# Run the API
go run cmd/rest-api/main.go
```

## Questions?

If you have questions about how to contribute:

- Open an issue with the `question` label
- Contact the maintainers
- Consult the [documentation](./docs/README.md)

## Code of Conduct

This project adheres to the [Code of Conduct](./CODE_OF_CONDUCT.md). By participating, you agree to uphold this code.

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.

---

**Thank you for contributing! 🎮**
