# Contributing to Yggdrasil Mesh Chat

Thank you for your interest in contributing! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How to Contribute](#how-to-contribute)
- [Development Setup](#development-setup)
- [Code Style](#code-style)
- [Commit Messages](#commit-messages)
- [Pull Request Process](#pull-request-process)
- [Reporting Bugs](#reporting-bugs)
- [Suggesting Features](#suggesting-features)

## Code of Conduct

Please be respectful and inclusive in all interactions. We are committed to providing a welcoming and inspiring community for everyone.

## How to Contribute

### Reporting Bugs

1. Check existing [issues](https://github.com/amafjarkasi/yggchat/issues) to avoid duplicates
2. Create a new issue with:
   - Clear title and description
   - Steps to reproduce
   - Expected vs actual behavior
   - System information (OS, Go version)
   - Logs or error messages

### Suggesting Features

1. Check existing [issues](https://github.com/amafjarkasi/yggchat/issues) for similar ideas
2. Create a new issue with:
   - Clear description of the feature
   - Use case / motivation
   - Proposed implementation (if applicable)

### Submitting Code

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## Development Setup

### Prerequisites

- Go 1.20+
- Git

### Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/yggchat.git
cd yggchat

# Add upstream remote
git remote add upstream https://github.com/amafjarkasi/yggchat.git

# Install dependencies
go mod tidy

# Build
make build

# Run tests
make test
```

### Project Structure

```
yggchat/
├── main.go           # Entry point
├── config.go         # Configuration
├── ygg.go            # Yggdrasil network
├── helpers.go        # Security utilities
├── web_server.go     # HTTP/SSE server
├── tui.go            # Terminal UI
├── ui_styles.go      # Themes
├── discovery.go      # Network discovery
├── image_render.go   # Image preview
├── assets/           # Static files
├── scripts/          # Utility scripts
└── web/              # Web frontend
```

## Code Style

### Go Code

- Follow standard Go conventions
- Use `gofmt` for formatting
- Run `go vet` before committing
- Keep functions focused and small
- Add comments for exported functions

```bash
# Format code
make fmt

# Run linter
make lint

# Run vet
make vet
```

### Naming Conventions

- **Packages**: lowercase, single word (`config`, `helpers`)
- **Functions**: CamelCase for exported (`LoadConfig`), camelCase for private (`getConfigFilename`)
- **Constants**: CamelCase (`ChunkSize`, `TimeFormat`)
- **Variables**: camelCase (`configFilename`)

### Error Handling

```go
// Good: Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to load config: %w", err)
}

// Bad: Ignore errors
_ = someFunction()
```

### Testing

- Write tests for new functionality
- Use table-driven tests
- Aim for >80% coverage on new code

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"case 1", "input1", "expected1"},
        {"case 2", "input2", "expected2"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := MyFunction(tt.input)
            if result != tt.expected {
                t.Errorf("got %q, want %q", result, tt.expected)
            }
        })
    }
}
```

## Commit Messages

Use conventional commits format:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation |
| `style` | Formatting (no code change) |
| `refactor` | Code restructuring |
| `test` | Adding tests |
| `chore` | Build/tooling changes |

### Examples

```
feat(tui): Add dark mode theme
fix(security): Sanitize file paths in transfers
docs(readme): Update installation instructions
test(crypto): Add ECDH key exchange tests
```

## Pull Request Process

### Before Submitting

1. **Update from upstream**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run tests**
   ```bash
   make test
   ```

3. **Check formatting**
   ```bash
   make fmt
   make vet
   ```

4. **Update documentation** if needed

### PR Template

```markdown
## Description

Brief description of changes

## Type of Change

- [ ] Bug fix
- [ ] New feature
- [ ] Documentation update
- [ ] Refactoring

## Testing

- [ ] Tests pass locally
- [ ] New tests added (if applicable)

## Checklist

- [ ] Code follows project style
- [ ] Self-review completed
- [ ] Comments added for complex code
- [ ] Documentation updated
```

### Review Process

1. Maintainers will review your PR
2. Address any feedback
3. Once approved, PR will be merged

## Areas for Contribution

### Good First Issues

Look for issues labeled `good first issue` on GitHub.

### Feature Ideas

- [ ] Message reactions/emojis
- [ ] Group chat support
- [ ] Voice messages
- [ ] Message editing/deletion
- [ ] Custom themes
- [ ] Plugin system
- [ ] Mobile app

### Documentation

- Improve README
- Add tutorials
- Write blog posts
- Translate to other languages

## Questions?

Feel free to open an issue for any questions about contributing.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
