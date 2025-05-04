# Git Flow Workflow

## Overview

This project follows the Git Flow branching strategy to maintain a clean and organized development process. Git Flow provides a robust framework for managing larger projects and coordinating parallel development.

## Branch Structure

### Main Branches

- `main` (formerly master)
  - Contains production-ready code
  - Always deployable
  - Protected branch
  - Only updated through releases and hotfixes

- `develop`
  - Main development branch
  - Contains latest delivered development changes
  - Base for feature branches
  - Merged into main through releases

### Supporting Branches

- `feature/*`
  - Branch from: `develop`
  - Merge back into: `develop`
  - Naming: `feature/issue-number-short-description`
  - Purpose: New features and non-emergency bug fixes

- `release/*`
  - Branch from: `develop`
  - Merge back into: `main` and `develop`
  - Naming: `release/vX.Y.Z`
  - Purpose: Release preparation, version bumping, and release-specific fixes

- `hotfix/*`
  - Branch from: `main`
  - Merge back into: `main` and `develop`
  - Naming: `hotfix/issue-number-short-description`
  - Purpose: Urgent production fixes

- `support/*`
  - Branch from: `main`
  - Naming: `support/vX.Y`
  - Purpose: Version support branches

## Workflow

### Starting a New Feature

```bash
# Create and switch to a new feature branch
git flow feature start feature-name

# Make your changes and commit them
git commit -m "feat: add new feature"

# Push the feature branch to remote
git flow feature publish feature-name

# When ready to merge
git flow feature finish feature-name
```

### Creating a Release

```bash
# Start a new release
git flow release start v1.0.0

# Make release-specific changes
git commit -m "chore: bump version to v1.0.0"

# Finish the release
git flow release finish v1.0.0
```

### Creating a Hotfix

```bash
# Start a new hotfix
git flow hotfix start hotfix-name

# Make the necessary changes
git commit -m "fix: resolve critical issue"

# Finish the hotfix
git flow hotfix finish hotfix-name
```

## Commit Messages

Follow the Conventional Commits specification:

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `style:` - Code style changes (formatting, etc.)
- `refactor:` - Code changes that neither fix bugs nor add features
- `perf:` - Performance improvements
- `test:` - Adding or modifying tests
- `chore:` - Changes to the build process or auxiliary tools

Format:

```git
type(scope): description

[optional body]

[optional footer]
```

## Pull Requests

1. Create a pull request from your feature branch to `develop`
2. Ensure all tests pass
3. Get at least one review
4. Squash commits if necessary
5. Merge only after approval

## Versioning

Follow Semantic Versioning (MAJOR.MINOR.PATCH):

- MAJOR: Incompatible API changes
- MINOR: Backwards-compatible functionality
- PATCH: Backwards-compatible bug fixes

## Best Practices

1. Keep branches up to date with their parent
2. Delete branches after merging
3. Use meaningful branch names
4. Write clear commit messages
5. Review code before merging
6. Keep the history clean and linear when possible

## Common Issues and Solutions

### Branch Conflicts

```bash
# Update your feature branch with latest develop
git checkout develop
git pull
git checkout feature/your-feature
git merge develop

# Resolve conflicts and commit
git commit -m "chore: resolve merge conflicts"
```

### Reverting Changes

```bash
# Revert a specific commit
git revert <commit-hash>

# Revert a merge commit
git revert -m 1 <merge-commit-hash>
```

## Tools and Extensions

- Git Flow AVH Edition
- GitKraken
- SourceTree
- VS Code Git Flow extension

## Additional Resources

- [Git Flow Documentation](https://nvie.com/posts/a-successful-git-branching-model/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Semantic Versioning](https://semver.org/)
