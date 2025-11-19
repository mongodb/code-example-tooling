# Dependency Manager - Usage Guide

## Quick Start

### 1. Build the CLI

```bash
cd dependency-manager
go build -o depman
```

### 2. Run Your First Check

Check for dependency updates in the current directory:

```bash
./depman check
```

Check a specific directory:

```bash
./depman check --path /path/to/your/project
```

Check a specific dependency file:

```bash
./depman check --path ./package.json
```

## Commands

### `check` - Dry Run Mode

Lists all available dependency updates without making any changes.

**Usage:**
```bash
./depman check [flags]
```

**Example Output:**
```
Found 2 dependency management file(s):

Checking ./package.json (package.json)...
  Found 3 update(s):
  Package    Current  Latest  Type
  -------    -------  ------  ----
  react      18.2.0   18.3.1  minor
  axios      1.4.0    1.6.2   minor
  vite       4.4.5    5.0.10  major

Checking ./backend/go.mod (go.mod)...
  All dependencies are up to date!
```

**When to use:**
- Before making any changes to understand what updates are available
- In CI/CD pipelines to report outdated dependencies
- Regular dependency audits

---

### `update` - Update Files Only

Updates dependency management files to the latest versions but does NOT install the dependencies.

**Usage:**
```bash
./depman update [flags]
```

**What it does:**
- **package.json**: Runs `ncu -u` to update version numbers
- **pom.xml**: Runs `mvn versions:use-latest-releases`
- **requirements.txt**: Runs `pip-compile --upgrade`
- **go.mod**: Runs `go get -u ./...` and `go mod tidy`
- **.csproj**: Runs `dotnet add package` for each outdated package

**When to use:**
- You want to review changes before installing
- You're preparing a PR and want to commit file changes separately
- You want to update files but install dependencies later

**Example:**
```bash
./depman update --path ./my-project
```

---

### `install` - Full Update

Updates dependency files AND installs/syncs the new dependencies.

**Usage:**
```bash
./depman install [flags]
```

**What it does:**
- Updates the dependency file (same as `update` command)
- Then installs dependencies:
  - **package.json**: Runs `npm install`
  - **pom.xml**: Runs `mvn clean install`
  - **requirements.txt**: Runs `pip install -r requirements.txt --upgrade`
  - **go.mod**: Runs `go mod download` and `go mod verify`
  - **.csproj**: Runs `dotnet restore` and `dotnet build`

**When to use:**
- You want to fully update and test with new dependencies immediately
- Automated update workflows
- Local development environment updates

**Example:**
```bash
./depman install --path ./my-project
```

---

## Flags

### Global Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--path` | `-p` | `.` (current directory) | Starting filepath or directory to scan |
| `--help` | `-h` | - | Show help information |

### Examples

```bash
# Check current directory
./depman check

# Check specific directory
./depman check --path ~/projects/my-app

# Check specific file
./depman check -p ./package.json

# Update all dependencies in a monorepo
./depman install --path ./monorepo
```

---

## Workflow Examples

### Scenario 1: Regular Dependency Audit

```bash
# 1. Check what's outdated
./depman check --path ./my-project

# 2. Review the output and decide if you want to update

# 3. Update files only (to review changes)
./depman update --path ./my-project

# 4. Review the git diff
git diff

# 5. If satisfied, install dependencies
./depman install --path ./my-project

# 6. Test your application
npm test  # or appropriate test command

# 7. Commit changes
git add .
git commit -m "chore: update dependencies"
```

### Scenario 2: CI/CD Integration

```bash
# In your CI pipeline, check for outdated dependencies
./depman check --path . || echo "Some dependencies are outdated"

# Optionally fail the build if there are major updates
# (requires custom scripting to parse output)
```

### Scenario 3: Monorepo Update

```bash
# Update all dependency files across the entire monorepo
./depman install --path ./monorepo

# The tool will find and update:
# - ./monorepo/frontend/package.json
# - ./monorepo/backend/go.mod
# - ./monorepo/services/api/pom.xml
# - ./monorepo/ml-service/requirements.txt
# - etc.
```

### Scenario 4: Single File Update

```bash
# Update just the frontend dependencies
./depman install --path ./frontend/package.json

# Update just the backend dependencies
./depman install --path ./backend/go.mod
```

---

## Prerequisites by Package Manager

### npm (package.json)

**Required:**
- Node.js and npm installed
- For `update` and `install` commands: `npm install -g npm-check-updates`

**Check installation:**
```bash
npm --version
ncu --version
```

### Maven (pom.xml)

**Required:**
- Java JDK installed
- Maven installed

**Check installation:**
```bash
mvn --version
```

### pip (requirements.txt)

**Required:**
- Python installed
- pip installed
- For `update` and `install` commands: `pip install pip-tools`

**Check installation:**
```bash
pip --version
pip-compile --version
```

### Go modules (go.mod)

**Required:**
- Go installed (1.16+)

**Check installation:**
```bash
go version
```

### NuGet (.csproj)

**Required:**
- .NET SDK installed

**Check installation:**
```bash
dotnet --version
```

---

## Troubleshooting

### "No dependency management files found"

**Cause:** The specified path doesn't contain any supported dependency files.

**Solution:**
- Verify the path is correct
- Ensure you have at least one of: package.json, pom.xml, requirements.txt, go.mod, or .csproj

### "npm is not installed or not in PATH"

**Cause:** The package manager for that file type is not available.

**Solution:**
- Install the required package manager
- Ensure it's in your system PATH
- The tool will skip files for unavailable package managers

### "npm-check-updates (ncu) is not installed"

**Cause:** The `update` or `install` command requires ncu for npm projects.

**Solution:**
```bash
npm install -g npm-check-updates
```

### "pip-compile (pip-tools) is not installed"

**Cause:** The `update` or `install` command requires pip-tools for Python projects.

**Solution:**
```bash
pip install pip-tools
```

---

## Tips and Best practices

1. **Always run `check` first** to see what will be updated
2. **Review changes** after running `update` before installing
3. **Test thoroughly** after running `install`
4. **Use version control** - commit before running updates
5. **Update incrementally** - consider updating one project at a time in monorepos
6. **Check breaking changes** - major version updates may require code changes
7. **Run tests** after updates to catch compatibility issues

---

## Advanced Usage

### Combining with Git

```bash
# Create a branch for updates
git checkout -b update-dependencies

# Run the update
./depman install --path .

# Review changes
git diff

# Run tests
npm test  # or your test command

# Commit if tests pass
git add .
git commit -m "chore: update dependencies"
git push origin update-dependencies
```

### Selective Updates

If you want to update only specific types of files, you can run the command on specific subdirectories:

```bash
# Update only frontend (npm)
./depman install --path ./frontend

# Update only backend (Go)
./depman install --path ./backend

# Update only Python services
./depman install --path ./services/python-api
```

---

## Exit Codes

- `0`: Success
- `1`: Error occurred (check stderr for details)

---

## Getting Help

```bash
# General help
./depman --help

# Command-specific help
./depman check --help
./depman update --help
./depman install --help
```

