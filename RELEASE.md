# Creating Releases

This document explains how to create releases for the amazing-cli project.

## Prerequisites

- You need push access to the repository
- The GoReleaser workflow is already configured in `.github/workflows/release.yml`
- The `.goreleaser.yml` configuration is set up

## Creating a Release

### Option 1: Using Git Tags (Recommended)

1. Make sure your code is committed and pushed to the main branch
2. Create and push a version tag:

```bash
# Create a new tag (replace v0.1.0 with your desired version)
git tag -a v0.1.0 -m "Release v0.1.0"

# Push the tag to GitHub
git push origin v0.1.0
```

3. GitHub Actions will automatically:
   - Build binaries for all platforms (Linux, Windows, macOS)
   - Create a GitHub release with the tag name
   - Upload pre-built binaries and checksums
   - Generate release notes from the changelog

### Option 2: Manual Release via GitHub Actions

1. Go to the Actions tab in GitHub
2. Select the "Release" workflow
3. Click "Run workflow"
4. Choose the branch and run it

### What Gets Built

The release process creates binaries for:
- **Windows**: `amazing-cli_Windows_x86_64.zip`, `amazing-cli_Windows_i386.zip`
- **macOS**: `amazing-cli_Darwin_x86_64.tar.gz`, `amazing-cli_Darwin_arm64.tar.gz`
- **Linux**: `amazing-cli_Linux_x86_64.tar.gz`, `amazing-cli_Linux_arm64.tar.gz`, `amazing-cli_Linux_i386.tar.gz`

Plus a `checksums.txt` file for verifying downloads.

## First Release

To create the first release (v0.1.0):

```bash
# From the main branch
git tag -a v0.1.0 -m "Initial release"
git push origin v0.1.0
```

This will make the installation scripts work, as they require at least one release to be available.

## Version Numbering

Follow semantic versioning (SemVer):
- `v1.0.0` - Major release (breaking changes)
- `v0.1.0` - Minor release (new features, backward compatible)
- `v0.0.1` - Patch release (bug fixes)

## Troubleshooting

### GitHub Actions Fails

- Check the Actions tab for detailed logs
- Verify the Go version in `.github/workflows/release.yml` matches your `go.mod`
- Ensure the `GITHUB_TOKEN` has sufficient permissions

### Installation Scripts Still Fail

- Wait a few minutes after creating the release for GitHub to propagate the artifacts
- Check that the release is public and not a draft
- Verify the asset names match what the installation scripts expect
