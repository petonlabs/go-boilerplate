# Contributing Guide

Thank you for considering contributing to go-boilerplate! This guide will help you get started.

---

## 🤝 How to Contribute

### Reporting Issues

Found a bug or have a feature request?

1. **Check existing issues** to avoid duplicates
2. **Open a new issue** with:
   - Clear, descriptive title
   - Steps to reproduce (for bugs)
   - Expected vs actual behavior
   - Environment details (OS, Go version, etc.)
   - Relevant code snippets or logs

### Submitting Pull Requests

1. **Fork the repository**
   ```bash
   git clone https://github.com/your-username/go-boilerplate.git
   cd go-boilerplate
   ```

2. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make your changes**
   - Follow existing code style
   - Add tests for new features
   - Update documentation

4. **Test your changes**
   ```bash
   cd apps/backend
   
   # Run tests
   go test ./...
   
   # Run linter
   golangci-lint run ./...
   
   # Check formatting
   go fmt ./...
   ```

5. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat: add amazing feature"
   ```

6. **Push to your fork**
   ```bash
   git push origin feature/your-feature-name
   ```

7. **Open a Pull Request**
   - Provide clear description
   - Reference related issues
   - Explain your changes
   - Wait for review

---

## 📋 Development Guidelines

### Code Style

Follow standard Go conventions:
- Use `gofmt` for formatting
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use meaningful variable names
- Keep functions small and focused
- Add comments for exported functions

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples**:
```
feat(auth): add password reset functionality
fix(database): resolve connection pool exhaustion
docs(readme): update installation instructions
test(handler): add unit tests for user endpoints
```

### Testing Requirements

- **Unit tests** for new features
- **Integration tests** for database operations
- **Maintain or improve** test coverage
- **All tests must pass** before merging

```bash
# Run all tests
go test ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Integration tests
go test -tags=integration ./...
```

### Documentation

Update documentation when:
- Adding new features
- Changing configuration
- Updating dependencies
- Fixing significant bugs

Documentation locations:
- **README.md**: Overview and quick start
- **docs/**: Detailed guides
- **Code comments**: Exported functions and complex logic

---

## 🔍 Code Review Process

### What We Look For

1. **Correctness**: Does it work as intended?
2. **Tests**: Are there adequate tests?
3. **Code Quality**: Is it readable and maintainable?
4. **Documentation**: Is it well-documented?
5. **Performance**: Any performance concerns?
6. **Security**: Any security implications?

### Review Timeline

- Initial review: Within 3 business days
- Follow-up reviews: Within 2 business days
- Complex PRs may take longer

### Addressing Feedback

- Be open to feedback
- Respond to all comments
- Make requested changes
- Re-request review when ready

---

## 🏗️ Project Structure

Understanding the structure helps you contribute effectively:

```
apps/backend/
├── cmd/              # Application entry points
├── internal/         # Private application code
│   ├── handler/     # HTTP handlers
│   ├── service/     # Business logic
│   ├── repository/  # Data access
│   ├── model/       # Domain models
│   └── ...
├── go.mod           # Dependencies
└── Dockerfile       # Container image
```

See [Architecture Guide](./docs/reference/ARCHITECTURE.md) for details.

---

## 🧪 Testing Strategy

### Test Organization

```
internal/handler/
├── user.go           # Implementation
└── user_test.go      # Tests
```

### Writing Tests

```go
func TestUserHandler_GetUser(t *testing.T) {
    // Arrange
    mockService := new(MockUserService)
    handler := NewUserHandler(mockService)
    
    // Act
    result, err := handler.GetUser("123")
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

See [Testing Guide](./docs/development/TESTING.md) for comprehensive examples.

---

## 🎨 Code Formatting

### Go Formatting

```bash
# Format all files
go fmt ./...

# Or use gofumpt (stricter)
gofumpt -w .
```

### Linting

```bash
# Run golangci-lint
golangci-lint run ./...

# Fix auto-fixable issues
golangci-lint run --fix ./...
```

### Pre-commit Hook

Install pre-commit hook to ensure quality:

```bash
# Copy hook
cp scripts/pre-commit.sh .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

---

## 🐛 Debugging

### Local Debugging

```bash
# Run with debugger
cd apps/backend
dlv debug ./cmd/go-boilerplate

# Debug specific test
dlv test ./internal/handler -- -test.run TestGetUser
```

### VS Code

Add to `.vscode/launch.json`:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/apps/backend/cmd/go-boilerplate"
    }
  ]
}
```

---

## 📦 Adding Dependencies

### Criteria

Only add dependencies that:
- Solve a significant problem
- Are well-maintained
- Have good documentation
- Don't duplicate existing functionality

### Process

1. **Discuss** in an issue first
2. **Add dependency**:
   ```bash
   cd apps/backend
   go get github.com/new/package
   ```
3. **Update documentation** in [Dependencies](./docs/reference/DEPENDENCIES.md)
4. **Explain rationale** in PR description

---

## 🚀 Release Process

### Versioning

We use [Semantic Versioning](https://semver.org/):
- `MAJOR.MINOR.PATCH`
- `1.0.0` → `1.0.1` (patch)
- `1.0.0` → `1.1.0` (minor)
- `1.0.0` → `2.0.0` (major)

### Creating a Release

1. Update version in relevant files
2. Update CHANGELOG.md
3. Create git tag:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
4. Create GitHub release with notes

---

## 💬 Community

### Getting Help

- **GitHub Discussions**: Ask questions
- **GitHub Issues**: Report bugs
- **Pull Requests**: Contribute code

### Code of Conduct

- Be respectful and inclusive
- Welcome newcomers
- Provide constructive feedback
- Focus on the code, not the person

---

## 📚 Resources

### Learning Go
- [A Tour of Go](https://go.dev/tour/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go by Example](https://gobyexample.com/)

### Project Resources
- [Architecture Guide](./docs/reference/ARCHITECTURE.md)
- [Best Practices](./docs/development/BEST_PRACTICES.md)
- [Testing Guide](./docs/development/TESTING.md)

### Tools
- [golangci-lint](https://golangci-lint.run/)
- [Delve Debugger](https://github.com/go-delve/delve)
- [VS Code Go Extension](https://marketplace.visualstudio.com/items?itemName=golang.go)

---

## ❓ FAQ

### How do I get started?

Follow the [Quick Start Guide](./docs/getting-started/QUICK_START.md) to set up your development environment.

### What should I work on?

Check issues labeled:
- `good first issue` - Great for newcomers
- `help wanted` - We need help with these
- `bug` - Bug fixes needed

### How long will my PR take to review?

Usually within 3 business days for initial review. Complex changes may take longer.

### Can I claim an issue?

Yes! Comment on the issue saying you'd like to work on it. We'll assign it to you.

### My tests are failing in CI

Run locally first:
```bash
./scripts/test-ci-locally.sh
```

This simulates the CI environment.

---

## 🙏 Thank You!

Thank you for contributing to go-boilerplate! Your efforts help make this project better for everyone.

---

**Questions?** Open an issue or start a discussion!
