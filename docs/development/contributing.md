# Contributing Guidelines

Thank you for contributing to rsyslox!

## Getting Started

1. Fork the repository
2. Clone your fork
3. Create a feature branch
4. Make your changes
5. Submit a pull request

## Development Setup

### Prerequisites

- Go 1.21+
- MySQL/MariaDB  
- rsyslog with MySQL support
- Docker (for testing)
- git
- make

### Clone Repository

```bash
git clone https://github.com/YOUR_USERNAME/rsyslox.git
cd rsyslox
```

### Install Dependencies

```bash
go mod download
go mod verify
```

### Build

```bash
# Development build
make build

# Static build (production)
make build-static

# Verify
./build/rsyslox --version
```

## Development Workflow

### 1. Create Feature Branch

```bash
git checkout -b feature/awesome-feature
```

Branch naming conventions:
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `refactor/` - Code refactoring
- `test/` - Test additions or fixes

### 2. Make Changes

#### Code Style

- Follow Go best practices
- Run `go fmt ./...` before committing
- Run `go vet ./...` to catch common errors
- Use meaningful variable names
- Add comments for complex logic

#### Project Structure

```
rsyslox/
├── cmd/                    # Command-line interface
├── internal/               # Internal packages
│   ├── config/            # Configuration handling
│   ├── database/          # Database operations
│   ├── handlers/          # HTTP handlers
│   ├── middleware/        # HTTP middleware
│   └── models/            # Data models
├── docker/                # Docker test environment
├── docs/                  # Documentation
├── main.go                # Application entry point
├── go.mod                 # Go modules
├── Makefile               # Build automation
└── rsyslox.service        # systemd service file
```

### 3. Testing

#### Unit Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/handlers/
```

#### Integration Tests (Docker)

```bash
# Build and test
make build-static
cd docker
docker-compose up -d

# Run manual tests
curl http://localhost:8000/health
curl "http://localhost:8000/logs?limit=5"

# Cleanup
docker-compose down -v
```

### 4. Code Quality

```bash
# Format code
go fmt ./...

# Run linter
go vet ./...

# Static analysis (optional)
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...
```

### 5. Commit Changes

```bash
git add .
git commit -m "Add awesome feature"
```

#### Commit Message Format

```
<type>: <subject>

<body>

<footer>
```

**Types:**
- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `style:` - Code style changes (formatting, etc.)
- `refactor:` - Code refactoring
- `test:` - Test additions or changes
- `chore:` - Build process or auxiliary tool changes

**Examples:**
```
feat: add support for PostgreSQL

Implement PostgreSQL driver alongside existing MySQL support.
Users can now choose between MySQL and PostgreSQL backends.

Closes #123
```

```
fix: handle null values in Message field

Previously null messages caused panic. Now properly handled
with empty string fallback.

Fixes #456
```

### 6. Push & Pull Request

```bash
git push origin feature/awesome-feature
```

Then create a Pull Request on GitHub:

1. Go to your fork on GitHub
2. Click "Pull Request"
3. Select your feature branch
4. Fill in the PR template
5. Submit

## Pull Request Guidelines

### PR Checklist

Before submitting a PR, ensure:

- [ ] Code follows Go best practices
- [ ] All tests pass
- [ ] New features have tests
- [ ] Documentation updated
- [ ] Commit messages are clear
- [ ] No merge conflicts
- [ ] Branch is up to date with main

### PR Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
How has this been tested?

## Checklist
- [ ] Code follows project style
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] No breaking changes (or documented)
```

## Code Review Process

1. **Automated Checks** - GitHub Actions run tests
2. **Maintainer Review** - Project maintainers review code
3. **Feedback** - Address review comments
4. **Approval** - PR approved by maintainer
5. **Merge** - PR merged to main branch

## Reporting Bugs

### Bug Report Template

```markdown
**Environment:**
- OS: Ubuntu 22.04
- Go Version: 1.21.5
- rsyslox Version: v0.2.3

**Expected Behavior:**
What should happen

**Actual Behavior:**
What actually happens

**Steps to Reproduce:**
1. Step one
2. Step two
3. ...

**Logs:**
```
Relevant log output
```
```

### Where to Report

- **Bugs:** [GitHub Issues](https://github.com/phil-bot/rsyslox/issues)
- **Security:** Email maintainers privately (see README)
- **Questions:** [GitHub Discussions](https://github.com/phil-bot/rsyslox/discussions)

## Feature Requests

### Before Requesting

1. Check existing issues
2. Search discussions
3. Review roadmap

### Feature Request Template

```markdown
**Problem:**
What problem does this solve?

**Proposed Solution:**
How should it work?

**Alternatives:**
What alternatives have you considered?

**Additional Context:**
Screenshots, examples, etc.
```

## Documentation

### Writing Docs

- Use clear, concise language
- Include code examples
- Add screenshots where helpful
- Test all commands/examples
- Follow existing structure

### Documentation Structure

```
docs/
├── getting-started/       # Installation, configuration
├── api/                   # API reference, examples
├── guides/                # Deployment, security, etc.
└── development/           # Docker, contributing
```

## Release Process

For maintainers:

1. Update version in code
2. Update CHANGELOG.md
3. Create git tag: `git tag -a v0.X.X -m "Release v0.X.X"`
4. Push tag: `git push origin v0.X.X`
5. GitHub Actions builds and publishes release

## Community Guidelines

### Code of Conduct

- Be respectful and inclusive
- Welcome newcomers
- Focus on constructive feedback
- Assume good intentions
- Report violations to maintainers

### Getting Help

- **Documentation:** Read the docs first
- **Discussions:** Ask questions in GitHub Discussions
- **Issues:** Report bugs via GitHub Issues
- **Discord:** Join our community (if available)

## Recognition

Contributors will be:
- Listed in CONTRIBUTORS.md
- Mentioned in release notes
- Credited in documentation

## Questions?

- **GitHub Discussions:** https://github.com/phil-bot/rsyslox/discussions
- **Issues:** https://github.com/phil-bot/rsyslox/issues

## Thank You!

Every contribution helps make rsyslox better for everyone! 🎉
