#!/bin/bash
# Helper script to create and push release tags
# Usage: ./push-release-tag.sh [tag-name] [tag-message]
#
# Example:
#   ./push-release-tag.sh v0.1.0 "Initial release"
#   ./push-release-tag.sh v0.2.0 "Add new features"

set -e

TAG="${1:-v0.1.0}"
MESSAGE="${2:-Release $TAG}"

echo "Creating release tag: $TAG"
echo "Message: $MESSAGE"
echo ""

# Check if tag already exists
if git rev-parse "$TAG" >/dev/null 2>&1; then
    echo "⚠️  Tag $TAG already exists"
    read -p "Delete and recreate? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git tag -d "$TAG"
        echo "✓ Deleted local tag"
    else
        echo "Aborted."
        exit 1
    fi
fi

# Create annotated tag
git tag -a "$TAG" -m "$MESSAGE"
echo "✓ Tag $TAG created"

# Push the tag
echo ""
echo "Pushing tag $TAG to origin..."
git push origin "$TAG"

echo ""
echo "✅ Tag pushed successfully!"
echo "✅ GitHub Actions will now build and publish the release"
echo ""
echo "Check the release status at:"
echo "https://github.com/huajianxiaowanzi/amazing-cli/actions"
echo ""
echo "Once complete, the release will be available at:"
echo "https://github.com/huajianxiaowanzi/amazing-cli/releases/tag/$TAG"
