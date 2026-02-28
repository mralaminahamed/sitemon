# Contributing to Sitemon

Thank you for your interest in contributing to Sitemon!

## How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported
2. Create a new issue with:
   - Clear title and description
   - Steps to reproduce
   - Expected vs actual behavior
   - Environment details (OS, Go version, etc.)

### Suggesting Features

1. Open a new issue with the `enhancement` tag
2. Describe the feature you'd like to see
3. Explain why it would be useful
4. Include any relevant examples or mockups

### Pull Requests

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes
4. Add tests if applicable
5. Ensure code quality:
   - Run `go mod tidy`
   - Run `go build` to verify compilation
6. Commit with clear messages
7. Push to your fork
8. Submit a pull request

## Development Setup

```bash
# Clone the repository
git clone https://github.com/mralaminahamed/sitemon.git
cd sitemon

# Install dependencies
go mod tidy

# Build
go build -o sitemon .

# Run tests (if any)
go test ./...
```

## Code Style

- Follow Go standard conventions
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions focused and small
- Run gofmt before committing

## Commit Messages

Use clear, descriptive commit messages:

```
feat: add latency percentiles to request command
fix: resolve timeout not being applied correctly
docs: update README with new examples
```

## Testing

Before submitting a PR:

```bash
# Build the project
go build -o sitemon .

# Test the commands
./sitemon check --url https://example.com
./sitemon request -u https://example.com -n 10
```

## License

By contributing to Sitemon, you agree that your contributions will be licensed under the MIT License.
