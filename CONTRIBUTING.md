# Contributing to SwiftLog

Thank you for your interest in contributing to SwiftLog! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Documentation](#documentation)
- [Community](#community)

## Code of Conduct

### Our Pledge

We are committed to providing a welcoming and inclusive environment for all contributors, regardless of background or experience level.

### Expected Behavior

- Be respectful and considerate
- Welcome newcomers and help them get started
- Provide constructive feedback
- Focus on what is best for the community
- Show empathy towards other community members

### Unacceptable Behavior

- Harassment or discriminatory language
- Personal attacks or trolling
- Publishing others' private information
- Other conduct that could be considered inappropriate

## Getting Started

### Prerequisites

Ensure you have the following installed:

- **Go** 1.21+ (for backend and CLI)
- **Node.js** 20+ (for frontend)
- **Docker** 24+ & **Docker Compose** v2+
- **Git**
- **PostgreSQL** client tools (optional, for debugging)

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/swiftlog.git
   cd swiftlog
   ```

3. Add upstream remote:
   ```bash
   git remote add upstream https://github.com/ORIGINAL_OWNER/swiftlog.git
   ```

4. Create a feature branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

### Setting Up Development Environment

1. **Copy environment configuration:**
   ```bash
   cp .env.example .env
   # Edit .env with your values
   ```

2. **Start infrastructure services:**
   ```bash
   make dev-up
   ```

3. **Run backend services locally:**
   ```bash
   # Terminal 1: Ingestor
   cd backend/cmd/ingestor && go run main.go

   # Terminal 2: API
   cd backend/cmd/api && go run main.go

   # Terminal 3: WebSocket
   cd backend/cmd/websocket && go run main.go

   # Terminal 4: AI Worker
   cd backend/cmd/ai-worker && go run main.go
   ```

4. **Run frontend locally:**
   ```bash
   # Terminal 5: Frontend
   cd frontend
   npm install
   npm run dev
   ```

5. **Build CLI:**
   ```bash
   cd cli
   go build -o swiftlog
   ```

## Development Workflow

### Branch Naming Convention

Use descriptive branch names:

- `feature/add-log-filtering` - New features
- `fix/api-auth-bug` - Bug fixes
- `docs/api-endpoints` - Documentation updates
- `refactor/database-layer` - Code refactoring
- `test/integration-tests` - Test additions

### Commit Message Guidelines

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**

```
feat(api): add log filtering endpoint

Add GET /api/v1/runs/:id/logs with query parameters for filtering
by log level and timestamp range.

Closes #123
```

```
fix(cli): handle broken pipe errors gracefully

Previously, the CLI would display confusing error messages when pipes
closed naturally. Now these errors are silenced.

Fixes #456
```

### Keeping Your Fork Updated

```bash
# Fetch upstream changes
git fetch upstream

# Merge upstream main into your branch
git checkout main
git merge upstream/main

# Rebase your feature branch
git checkout feature/your-feature-name
git rebase main
```

## Coding Standards

### Go Code Style

Follow [Effective Go](https://golang.org/doc/effective_go) and [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).

**Key Points:**
- Use `gofmt` to format code
- Use `golint` for linting
- Write clear variable names (avoid abbreviations)
- Document exported functions with comments
- Handle errors explicitly
- Use context for cancellation

**Example:**

```go
// GetProjectByID retrieves a project by its unique identifier.
// Returns ErrNotFound if the project doesn't exist.
func (r *ProjectRepository) GetProjectByID(ctx context.Context, id uuid.UUID) (*models.Project, error) {
    var project models.Project
    err := r.db.GetContext(ctx, &project, `
        SELECT id, name, user_id, created_at, updated_at
        FROM projects
        WHERE id = $1
    `, id)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("failed to get project: %w", err)
    }
    return &project, nil
}
```

### TypeScript/React Code Style

Follow [Airbnb JavaScript Style Guide](https://github.com/airbnb/javascript) and React best practices.

**Key Points:**
- Use TypeScript for all files
- Use functional components with hooks
- Add `'use client'` for client components
- Use async/await for asynchronous code
- Handle errors with try/catch
- Provide loading states

**Example:**

```tsx
'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { Project } from '@/types';

export default function ProjectList() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadProjects();
  }, []);

  const loadProjects = async () => {
    try {
      setLoading(true);
      const data = await api.listProjects();
      setProjects(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load projects');
    } finally {
      setLoading(false);
    }
  };

  if (loading) return <div>Loading...</div>;
  if (error) return <div>Error: {error}</div>;

  return (
    <div className="space-y-4">
      {projects.map(project => (
        <ProjectCard key={project.id} project={project} />
      ))}
    </div>
  );
}
```

### SQL Style

- Use uppercase for SQL keywords
- Indent nested queries
- Use parameterized queries (never string concatenation)
- Add comments for complex queries

**Example:**

```sql
-- Get all runs for a group with their status counts
SELECT
    lg.id,
    lg.name,
    COUNT(lr.id) as total_runs,
    COUNT(CASE WHEN lr.status = 'success' THEN 1 END) as successful_runs,
    COUNT(CASE WHEN lr.status = 'failed' THEN 1 END) as failed_runs
FROM log_groups lg
LEFT JOIN log_runs lr ON lr.group_id = lg.id
WHERE lg.project_id = $1
GROUP BY lg.id, lg.name
ORDER BY lg.created_at DESC;
```

## Testing

### Backend Tests

```bash
# Run all tests
cd backend
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/repository

# Run with verbose output
go test -v ./...
```

### Frontend Tests

```bash
cd frontend

# Run unit tests
npm test

# Run with coverage
npm test -- --coverage

# Run E2E tests
npm run test:e2e
```

### Integration Tests

Use the provided test scripts:

```bash
cd tests
./run_all_tests.sh
```

### Writing Tests

#### Go Test Example

```go
func TestProjectRepository_GetByID(t *testing.T) {
    db := setupTestDB(t)
    defer teardownTestDB(t, db)

    repo := repository.NewProjectRepository(db)

    // Create test project
    project := &models.Project{
        Name:   "test-project",
        UserID: testUserID,
    }
    err := repo.Create(context.Background(), project)
    require.NoError(t, err)

    // Test retrieval
    retrieved, err := repo.GetByID(context.Background(), project.ID)
    require.NoError(t, err)
    assert.Equal(t, project.Name, retrieved.Name)
    assert.Equal(t, project.UserID, retrieved.UserID)
}
```

#### React Test Example

```tsx
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import ProjectList from './ProjectList';

jest.mock('@/lib/api');

describe('ProjectList', () => {
  it('renders projects after loading', async () => {
    const mockProjects = [
      { id: '1', name: 'Project 1' },
      { id: '2', name: 'Project 2' }
    ];

    (api.listProjects as jest.Mock).mockResolvedValue(mockProjects);

    render(<ProjectList />);

    expect(screen.getByText('Loading...')).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText('Project 1')).toBeInTheDocument();
      expect(screen.getByText('Project 2')).toBeInTheDocument();
    });
  });
});
```

## Submitting Changes

### Before Submitting

1. **Run tests:**
   ```bash
   # Backend
   cd backend && go test ./...

   # Frontend
   cd frontend && npm test
   ```

2. **Format code:**
   ```bash
   # Go
   gofmt -w .

   # TypeScript
   cd frontend && npm run lint
   ```

3. **Update documentation** if needed

4. **Test manually** with the full stack running

### Pull Request Process

1. **Push your branch:**
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create a Pull Request** on GitHub

3. **Fill out the PR template:**
   - Description of changes
   - Related issues (e.g., "Closes #123")
   - Testing done
   - Screenshots (if UI changes)

4. **Wait for review:**
   - Address reviewer feedback
   - Push additional commits if needed
   - Maintainers will merge when approved

### Pull Request Template

```markdown
## Description
Brief description of changes

## Related Issues
Closes #123

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing Done
- [ ] Unit tests added/updated
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex code
- [ ] Documentation updated
- [ ] No new warnings generated
```

## Documentation

### When to Update Docs

Update documentation when you:

- Add a new feature
- Change an API endpoint
- Modify configuration options
- Fix a significant bug
- Add new dependencies

### Documentation Files

- `README.md` - Project overview and quick start
- `docs/API.md` - API reference
- `docs/ARCHITECTURE.md` - System architecture
- `cli/README.md` - CLI usage
- `frontend/README.md` - Frontend development
- `tests/README.md` - Testing guide

### Documentation Style

- Use clear, concise language
- Provide code examples
- Include screenshots for UI changes
- Use proper markdown formatting
- Add table of contents for long documents

## Community

### Getting Help

- **GitHub Issues**: Report bugs or request features
- **GitHub Discussions**: Ask questions or share ideas
- **Documentation**: Check docs first

### Issue Reporting

When reporting a bug, include:

- SwiftLog version
- Operating system
- Steps to reproduce
- Expected behavior
- Actual behavior
- Error messages or logs
- Screenshots (if applicable)

**Bug Report Template:**

```markdown
**Describe the bug**
A clear description of what the bug is.

**To Reproduce**
Steps to reproduce:
1. Run command '...'
2. Click on '...'
3. See error

**Expected behavior**
What you expected to happen.

**Screenshots**
If applicable, add screenshots.

**Environment:**
- OS: [e.g., macOS 14.0]
- SwiftLog version: [e.g., 1.0.0]
- Go version: [e.g., 1.21.0]
- Docker version: [e.g., 24.0.0]

**Additional context**
Any other context about the problem.
```

### Feature Requests

When requesting a feature:

- Explain the use case
- Describe the desired behavior
- Suggest implementation approach (optional)
- Mention alternatives you've considered

## Recognition

Contributors will be:

- Listed in CONTRIBUTORS.md
- Mentioned in release notes
- Credited in commit history

Thank you for contributing to SwiftLog!

---

## Quick Reference

**Commands:**
```bash
# Setup
make dev-up                    # Start infrastructure
make cli                       # Build CLI

# Development
cd backend/cmd/api && go run main.go   # Run API
cd frontend && npm run dev             # Run frontend

# Testing
go test ./...                  # Backend tests
npm test                       # Frontend tests
./tests/run_all_tests.sh       # Integration tests

# Cleanup
make dev-down                  # Stop infrastructure
```

**Resources:**
- [Main README](./README.md)
- [Architecture](./docs/ARCHITECTURE.md)
- [API Docs](./docs/API.md)
- [CLI Docs](./cli/README.md)

---

**Questions?** Open a GitHub Discussion or create an issue.
