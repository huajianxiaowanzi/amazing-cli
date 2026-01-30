#!/bin/bash
# Helper script to create and push release tags
# Usage: ./push-release-tag.sh [tag-name] [tag-message]
#
# Example:
#   ./push-release-tag.sh v0.1.0 "Initial release"
#   ./push-release-tag.sh v0.2.0 "Add new features"
#
# Note: This script requires appropriate permissions to push tags.
# Repository protection rules may prevent tag creation/deletion.

set -e

TAG="${1:-v0.1.0}"
MESSAGE="${2:-Release $TAG}"

echo "Creating release tag: $TAG"
echo "Message: $MESSAGE"
echo ""

# Check if tag already exists locally
if git rev-parse "$TAG" >/dev/null 2>&1; then
    echo "⚠️  Tag $TAG already exists locally"
    read -p "Delete and recreate? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git tag -d "$TAG"
        echo "✓ Deleted local tag"
        
        # Check if tag exists remotely
        if git ls-remote --tags origin | grep -q "refs/tags/$TAG"; then
            echo ""
            echo "⚠️  Tag $TAG also exists on remote."
            echo "You may need to delete it remotely first:"
            echo "  git push origin :refs/tags/$TAG"
            echo ""
            read -p "Continue anyway? (y/N) " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                echo "Aborted."
                exit 1
            fi
        fi
    else
        echo "Aborted."
        exit 1
    fi
fi

# Create annotated tag
git tag -a "$TAG" -m "$MESSAGE"
echo "✓ Tag $TAG created locally"

# Push the tag
echo ""
echo "Pushing tag $TAG to origin..."
if git push origin "$TAG" 2>&1; then
    echo ""
    echo "✅ Tag pushed successfully!"
    echo "✅ GitHub Actions will now build and publish the release"
    echo ""
    echo "Check the release status at:"
    echo "https://github.com/huajianxiaowanzi/amazing-cli/actions"
    echo ""
    echo "Once complete, the release will be available at:"
    echo "https://github.com/huajianxiaowanzi/amazing-cli/releases/tag/$TAG"
else
    echo ""
    echo "❌ Failed to push tag. This may be due to:"
    echo "  - Insufficient permissions"
    echo "  - Repository protection rules"
    echo "  - Tag already exists remotely with different content"
    echo ""
    echo "Contact a repository administrator for assistance."
    exit 1
fi
