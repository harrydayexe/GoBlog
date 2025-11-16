# GitHub Actions Setup

## Required Secrets

The following secrets need to be configured in your repository settings (Settings → Secrets and variables → Actions):

### PAT_TOKEN (Required for Auto-Tagging)

**Purpose**: Allows the auto-tag workflow to create tags that trigger the release workflow.

**Why needed**: When using the default `GITHUB_TOKEN`, workflows triggered by that token cannot trigger other workflows (GitHub security feature). A Personal Access Token (PAT) bypasses this limitation.

**Setup steps**:

1. Go to GitHub Settings → Developer settings → Personal access tokens → [Fine-grained tokens](https://github.com/settings/tokens?type=beta)

2. Click "Generate new token"

3. Configure the token:
   - **Token name**: `GoBlog Auto-Tag Token` (or similar)
   - **Expiration**: Choose an expiration period (recommend 90 days or 1 year)
   - **Repository access**: Select "Only select repositories" → Choose your GoBlog repository
   - **Repository permissions**:
     - Contents: **Read and write** (required to create tags and trigger workflows)

4. Click "Generate token" and copy the token value

5. Add the token to your repository:
   - Go to your repository → Settings → Secrets and variables → Actions
   - Click "New repository secret"
   - Name: `PAT_TOKEN`
   - Secret: Paste the token value
   - Click "Add secret"

**Token expiration**: When your token expires, the auto-tag workflow will fail. You'll need to generate a new token and update the secret.

### Other Secrets

#### DOCKERHUB_USERNAME (Optional)
- Required if using Docker workflows
- Your Docker Hub username

#### DOCKERHUB_TOKEN (Optional)
- Required if using Docker workflows
- Docker Hub access token (Settings → Security → New Access Token)

#### CODECOV_TOKEN (Optional)
- Required for code coverage reporting
- Get from [Codecov.io](https://codecov.io) after adding your repository

## Workflow Overview

### Tag Workflow (`.github/workflows/tag.yml`)
- **Trigger**: Push to `main` branch
- **Action**: Creates version tags based on conventional commits
- **Token used**: `PAT_TOKEN` (to trigger release workflow)

### Release Workflow (`.github/workflows/release.yml`)
- **Trigger**: Push of version tags (e.g., `v1.0.0`)
- **Action**: Builds binaries and creates GitHub release using GoReleaser
- **Token used**: `GITHUB_TOKEN` (automatically provided)

## Troubleshooting

### Release workflow doesn't run after tag is created
- **Cause**: `PAT_TOKEN` secret is missing or invalid
- **Solution**: Check that the `PAT_TOKEN` secret is configured correctly and hasn't expired

### "Resource not accessible by integration" error
- **Cause**: PAT doesn't have correct permissions
- **Solution**: Ensure PAT has "Contents: Read and write" permission

### Tag is created but release fails
- **Cause**: Issue with GoReleaser configuration or Go setup
- **Solution**: Check the release workflow logs for specific error messages
