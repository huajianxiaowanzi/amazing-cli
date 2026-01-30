# Release v0.1.0 - Ready to Publish

## Status: Ready for Tag Push ✅

All code changes have been prepared and the repository is ready for the first release (v0.1.0).

### What Has Been Done

1. ✅ **Version Support Added**: Added `version` variable to `main.go` that will be injected at build time
2. ✅ **Git Configuration**: Configured `push.followTags` to automatically push tags with commits
3. ✅ **Helper Script**: Created `push-release-tag.sh` to simplify future releases
4. ✅ **Tag Created Locally**: Tag `v0.1.0` has been created locally and points to the latest commit
5. ✅ **All Changes Committed**: All code changes have been committed and pushed to the branch

### What Needs To Be Done

Due to repository protection rules, the tag `v0.1.0` could not be pushed automatically. This is expected behavior as tag creation is typically restricted to prevent unauthorized releases.

**To complete the release, a repository owner/maintainer needs to push the tag:**

#### Option 1: Using the helper script (Recommended)

```bash
# Clone/pull the latest changes
git checkout copilot/release-initial-version
git pull

# The tag should already exist locally. Push it:
./push-release-tag.sh v0.1.0 "Initial release"
```

#### Option 2: Manual tag push

```bash
# Clone/pull the latest changes
git checkout copilot/release-initial-version
git pull

# Verify the tag exists
git tag -l v0.1.0

# If the tag doesn't exist, create it:
git tag -a v0.1.0 -m "Initial release"

# Push the tag
git push origin v0.1.0
```

#### Option 3: Merge to main and tag from main branch

```bash
# Merge this PR to main
# Then from main branch:
git checkout main
git pull
git tag -a v0.1.0 -m "Initial release"
git push origin v0.1.0
```

### What Happens Next

Once the tag is pushed:

1. GitHub Actions will automatically trigger the "Release" workflow
2. GoReleaser will build binaries for all platforms:
   - Linux (amd64, arm64, 386)
   - macOS (amd64, arm64)  
   - Windows (amd64, 386)
3. A GitHub Release will be created with:
   - Pre-built binaries
   - Checksums for verification
   - Automated changelog
4. The installation scripts (`install.sh` and `install.ps1`) will work correctly

### Monitoring the Release

After pushing the tag, monitor the release process:

- **Actions**: https://github.com/huajianxiaowanzi/amazing-cli/actions
- **Releases**: https://github.com/huajianxiaowanzi/amazing-cli/releases

The release process typically takes 2-5 minutes to complete.

### Troubleshooting

If you encounter issues:

1. **Permission Denied**: Ensure you have write access to the repository and permission to create tags
2. **Tag Already Exists**: Delete the remote tag first with `git push origin :refs/tags/v0.1.0`
3. **Workflow Doesn't Trigger**: Check that `.github/workflows/release.yml` exists and is properly configured
4. **Build Failures**: Check the Actions logs for detailed error messages

### Repository Protection Rules

The repository currently has protection rules that prevent tag creation from the bot account. This is a good security practice. Only authorized users should be able to create release tags.

If you want to allow automated tag creation in the future, you can:
1. Go to Settings → Rules → Rulesets
2. Find the rule blocking tag creation
3. Add an exception for the bot account or CI/CD workflows

---

**Next Steps After Release:**

Once v0.1.0 is published, users will be able to install the CLI using:

```bash
# Linux/macOS
curl -fsSL https://raw.githubusercontent.com/huajianxiaowanzi/amazing-cli/main/install.sh | sh

# Windows
irm https://raw.githubusercontent.com/huajianxiaowanzi/amazing-cli/main/install.ps1 | iex
```

Or download binaries directly from the Releases page.
