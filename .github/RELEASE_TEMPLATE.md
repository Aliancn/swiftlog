# Release Checklist

This document describes the process for creating a new release.

## Pre-Release Checklist

- [ ] All tests passing
- [ ] Documentation updated
- [ ] CHANGELOG.md updated with changes
- [ ] VERSION file updated
- [ ] No breaking changes without major version bump

## Release Process

### 1. Update Version

Update the version in these files:
- `VERSION` - Version number
- `CHANGELOG.md` - Add release notes
- `cli/cmd/root.go` - Update default version

### 2. Create Release Tag

```bash
# Create annotated tag
git tag -a v0.1.0 -m "Release v0.1.0"

# Push tag to GitHub
git push origin v0.1.0
```

### 3. GitHub Actions

Once the tag is pushed, GitHub Actions will automatically:

1. **Build CLI binaries** for all platforms:
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64)

2. **Create GitHub Release**:
   - Upload all CLI binaries
   - Generate SHA256 checksums
   - Extract release notes from CHANGELOG.md

3. **Build and push Docker images**:
   - Build for multi-platform (amd64, arm64)
   - Push to GitHub Container Registry
   - Tag with version and latest

4. **Deploy to servers** (if configured):
   - Update production/staging environments
   - Run health checks
   - Verify deployment

### 4. Verify Release

After GitHub Actions completes:

- [ ] Check [Releases page](https://github.com/aliancn/swiftlog/releases) for new release
- [ ] Verify all CLI binaries are attached
- [ ] Check Docker images on [Packages](https://github.com/aliancn?tab=packages)
- [ ] Test download and installation of CLI
- [ ] Verify deployment (if auto-deploy is enabled)

### 5. Announce

- [ ] Update README.md if needed
- [ ] Post to discussions/social media
- [ ] Notify users of breaking changes

## Version Numbering

SwiftLog follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version for incompatible API changes
- **MINOR** version for new functionality (backwards compatible)
- **PATCH** version for bug fixes (backwards compatible)

Examples:
- `v0.1.0` - Initial release
- `v0.1.1` - Bug fix release
- `v0.2.0` - New features added
- `v1.0.0` - Production-ready, stable API

## Hotfix Process

For critical bug fixes:

```bash
# Create hotfix branch from tag
git checkout -b hotfix/v0.1.1 v0.1.0

# Make fixes
git commit -am "Fix critical bug"

# Create new tag
git tag -a v0.1.1 -m "Hotfix v0.1.1"
git push origin v0.1.1

# Merge back to main
git checkout main
git merge hotfix/v0.1.1
git push origin main
```

## Rollback

If a release has issues:

1. **Mark release as pre-release** on GitHub
2. **Deploy previous version**:
   ```bash
   git tag v0.1.0-rollback v0.0.9
   git push origin v0.1.0-rollback
   ```
3. **Fix issues and create new release**

## Release Notes Template

```markdown
## What's New in v0.x.0

### Features
- New feature description

### Improvements
- Enhancement description

### Bug Fixes
- Bug fix description

### Breaking Changes
- Breaking change description (if any)

### Installation

**CLI:**
Download for your platform:
- [Linux amd64](link)
- [Linux arm64](link)
- [macOS Intel](link)
- [macOS Apple Silicon](link)
- [Windows](link)

**Docker:**
\```bash
docker pull ghcr.io/aliancn/swiftlog/api:v0.x.0
\```

**Upgrade from previous version:**
\```bash
./deploy.sh -h server.com -u deploy -v v0.x.0
\```

### Full Changelog
See [CHANGELOG.md](CHANGELOG.md) for full details.
```
