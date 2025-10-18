# Version Management Guide

## Overview

OpenJobs displays a discreet version string at the bottom of the dashboard that updates automatically on each build.

## Version Display

**Location:** Bottom of dashboard (footer)
**Format:** `OpenJobs v2025.10.18-0236` with optional build time below

**Appearance:**
- Small, light gray text (0.75rem, 60% opacity)
- Centered at bottom of page
- Shows version and build time
- Updates automatically from `/health` endpoint

## Building with Version

### Option 1: Automatic Version (Recommended)

Use the build script which automatically generates a version based on date/time:

```bash
./build.sh
```

**Version format:** `YYYY.MM.DD-HHMM`
**Example:** `2025.10.18-0236` (built on Oct 18, 2025 at 02:36)

### Option 2: Git Tag Version

If you have a git tag, the build script will use it:

```bash
git tag v1.2.3
./build.sh
```

**Version:** `v1.2.3`

### Option 3: Manual Version

Set version manually during build:

```bash
go build \
    -ldflags "-X 'main.Version=v1.2.3' -X 'main.BuildTime=$(date -u +"%Y-%m-%d %H:%M UTC")'" \
    -o openjobs \
    cmd/openjobs/main.go
```

### Option 4: Docker Build

For Docker builds, add build args:

```dockerfile
# In Dockerfile
ARG VERSION=dev
ARG BUILD_TIME=unknown

RUN go build \
    -ldflags "-X 'main.Version=${VERSION}' -X 'main.BuildTime=${BUILD_TIME}'" \
    -o openjobs \
    cmd/openjobs/main.go
```

Then build with:

```bash
docker build \
    --build-arg VERSION=$(date +"%Y.%m.%d-%H%M") \
    --build-arg BUILD_TIME="$(date -u +"%Y-%m-%d %H:%M UTC")" \
    -t openjobs:latest .
```

## Version in API

The version is also available via the `/health` endpoint:

```bash
curl https://your-openjobs.com/health
```

Response:
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "service": "openjobs",
    "version": "2025.10.18-0236",
    "build_time": "2025-10-18 00:36 UTC"
  }
}
```

## Easypanel Deployment

For Easypanel deployments, the version will be set during the build process.

### Manual Trigger

If you want to force a specific version in Easypanel:

1. Go to your OpenJobs service
2. Click "Environment"
3. Add build-time environment variable (if supported)
4. Rebuild

Or update your build command in Easypanel:

```bash
VERSION=$(date +"%Y.%m.%d-%H%M") && \
go build -ldflags "-X 'main.Version=$VERSION' -X 'main.BuildTime=$(date -u +"%Y-%m-%d %H:%M UTC")'" \
-o openjobs cmd/openjobs/main.go
```

## Version Scheme

### Date-Based (Default)
- Format: `YYYY.MM.DD-HHMM`
- Example: `2025.10.18-0236`
- Automatically generated on each build
- Easy to see when it was built
- No manual management needed

### Semantic Versioning (Optional)
- Format: `vMAJOR.MINOR.PATCH`
- Example: `v1.2.3`
- Requires git tags
- Use for releases

### Development
- Format: `dev`
- Used when building without version flags
- Indicates development build

## Checking Current Version

### From Dashboard
1. Open OpenJobs dashboard
2. Scroll to bottom
3. See version in footer

### From CLI
```bash
curl -s https://your-openjobs.com/health | jq -r '.data.version'
```

### From Container
```bash
docker exec openjobs-container ./openjobs --version
# (if --version flag is implemented)
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Build and Deploy

on:
  push:
    branches: [main]
    tags: ['v*']

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set version
        id: version
        run: |
          if [[ $GITHUB_REF == refs/tags/* ]]; then
            echo "VERSION=${GITHUB_REF#refs/tags/}" >> $GITHUB_OUTPUT
          else
            echo "VERSION=$(date +"%Y.%m.%d-%H%M")" >> $GITHUB_OUTPUT
          fi
      
      - name: Build
        run: |
          go build \
            -ldflags "-X 'main.Version=${{ steps.version.outputs.VERSION }}' \
                      -X 'main.BuildTime=$(date -u +"%Y-%m-%d %H:%M UTC")'" \
            -o openjobs \
            cmd/openjobs/main.go
```

## Troubleshooting

### Version shows "dev"
- You built without version flags
- Solution: Use `./build.sh` or set version manually

### Version doesn't update on dashboard
- Clear browser cache
- Check `/health` endpoint shows correct version
- Verify JavaScript is loading version from health data

### Build time shows "unknown"
- You didn't set BuildTime during build
- Solution: Use `./build.sh` or add `-X 'main.BuildTime=...'` flag

## Best Practices

1. **Always use build script** - Ensures consistent versioning
2. **Tag releases** - Use git tags for production releases
3. **Date-based for dev** - Use automatic date-based versions for development
4. **Document changes** - Keep a CHANGELOG.md for version history
5. **Automate in CI/CD** - Set version automatically in deployment pipeline
