# GitHub Actions Workflows

## Overview

This directory contains GitHub Actions workflows for the Confluent CLI.

## Workflows

### skills.yml - Skills Validation

Ensures that generated MCP skills are valid, build correctly across platforms, and stay synchronized with the codebase.

**Purpose:** Prevent invalid skill manifests from being merged and ensure cross-platform compatibility.

**Triggers:**
- Push to `main` branch
- Push to `feature/**` branches
- Pull requests to `main` branch

**Expected runtime:** 10-15 minutes for full CI run

## Jobs

### 1. validate-skills (ubuntu-latest)

**Purpose:** Generates and validates the skills manifest

**Steps:**
1. Checkout code
2. Setup Go using `.go-version` file
3. Generate skills manifest: `make generate-skills`
4. Validate manifest: `make validate-skills`
5. Check tool count and warn if >= 200 tools
6. Upload `skills.json` as artifact (retention: 30 days)

**Runtime:** ~1-2 minutes

**Artifacts:**
- `skills-manifest.zip` - Download from Actions tab → Run → Artifacts

### 2. build-with-skills (multi-platform matrix)

**Purpose:** Tests cross-platform builds with embedded skills

**Matrix:**
- **Platforms:** ubuntu-latest, macos-latest, windows-latest
- **Go version:** 1.25.7

**Steps:**
1. Checkout code
2. Setup Go with matrix version
3. Build with skills: `make build`
4. Verify binary exists (platform-specific checks)
5. Test embedded skills loaded (placeholder for future `confluent mcp --version`)

**Runtime:** ~3-5 minutes per platform (runs in parallel)

**Depends on:** validate-skills (won't run if validation fails)

### 3. compare-manifests (ubuntu-latest, PR only)

**Purpose:** Ensures the committed manifest is up to date

**Condition:** Only runs on pull requests

**Steps:**
1. Checkout code
2. Setup Go
3. Generate fresh manifest: `make generate-skills`
4. Compare with committed version using `git diff`
5. Fail if differences detected

**Runtime:** ~1 minute

**Common failure:** "Skills manifest out of date" - see Troubleshooting below

## Local Testing

Replicate CI checks locally before pushing:

### 1. Replicate CI validation

```bash
# Clean and regenerate skills
make clean
make generate-skills
make validate-skills
```

### 2. Replicate CI build

```bash
# Clean and build from scratch
make clean
make build
```

### 3. Replicate manifest comparison

```bash
# Generate fresh manifest and check for differences
make generate-skills
git diff pkg/mcp/skills.json

# If differences exist, commit them:
git add pkg/mcp/skills.json
git commit -m "chore: update skills manifest"
```

## Troubleshooting

### "Skills manifest out of date"

**Problem:** The committed `pkg/mcp/skills.json` doesn't match what CI generates

**Solution:**
```bash
make generate-skills
git add pkg/mcp/skills.json
git commit -m "chore: update skills manifest"
git push
```

**Why this happens:**
- Code changes affected command tree structure
- Manifest not regenerated after command changes
- Different Go version locally vs CI (rare)

### "Validation failed"

**Problem:** The manifest has validation errors (duplicate names, count mismatch, etc.)

**Solution:**
1. Check CI logs for specific error message
2. Common issues:
   - Duplicate tool names → Check command registration
   - Count mismatch → Regenerate manifest: `make generate-skills`
   - Missing required fields → Check command metadata
3. Fix the underlying issue in code
4. Regenerate and validate:
   ```bash
   make clean
   make generate-skills
   make validate-skills
   ```

### Build failures

**Problem:** Build fails on specific platform

**Solution:**
1. Verify local build works first:
   ```bash
   make clean
   make build
   ```
2. If local build succeeds but CI fails:
   - Check platform-specific dependencies
   - Check Go version matches `.go-version`
   - Check CI logs for platform-specific errors
3. For cross-platform testing locally:
   ```bash
   # Linux
   GOOS=linux GOARCH=amd64 make cross-build

   # Windows
   GOOS=windows GOARCH=amd64 make cross-build

   # macOS (Apple Silicon)
   GOARCH=arm64 make build
   ```

### Tool count warning

**Problem:** CI shows warning "Tool count (XXX) exceeds recommended limit (200)"

**Impact:** This is a warning, not an error. The build will still pass.

**What it means:**
- MCP protocol has soft limits on tool counts
- High tool counts may impact client performance
- Consider grouping commands or reducing exported tools

**Action required:** None immediately, but consider optimization if count grows significantly

## Workflow Validation

To validate workflow syntax locally:

```bash
# Using Python (requires PyYAML)
python3 -c "import yaml; yaml.safe_load(open('.github/workflows/skills.yml'))"
```

If syntax is valid, no output is shown. Invalid syntax will show error details.

## Adding New Workflows

When adding new workflows:

1. Create `.github/workflows/new-workflow.yml`
2. Follow existing patterns (checkout, setup-go, etc.)
3. Use `go-version-file: .go-version` for Go setup
4. Test locally before committing
5. Document in this README

## CI Environment

**Runner images:**
- `ubuntu-latest` - Ubuntu (latest stable)
- `macos-latest` - macOS (latest stable)
- `windows-latest` - Windows Server (latest stable)

**Go setup:**
- Uses `actions/setup-go@v4` with `go-version-file: .go-version`
- Ensures consistent Go version across all jobs
- Current version: 1.25.7 (see `.go-version`)

**Artifacts:**
- Retention: 30 days
- Download from: Actions tab → Workflow run → Artifacts section
- Useful for debugging manifest generation differences
