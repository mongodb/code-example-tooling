# Dependency Manager CLI

A powerful CLI tool built with Cobra that scans directories for dependency management files and helps you check and update dependencies across multiple package managers.

## Features

- 🔍 **Multi-language support**: Handles package.json, pom.xml, requirements.txt, go.mod, and .csproj files
- 📊 **Dry run mode**: Check for updates without making changes
- 🔄 **Selective updates**: Update dependency files without installing
- ⚡ **Full automation**: Update and install dependencies in one command
- 🌳 **Recursive scanning**: Automatically finds all dependency files in subdirectories

## Supported Package Managers

| Language/Framework | File Type | Package Manager | Commands Used |
|-------------------|-----------|-----------------|---------------|
| JavaScript/Node.js | package.json | npm | `npm outdated`, `ncu -u`, `npm install` |
| Java | pom.xml | Maven | `mvn versions:display-dependency-updates`, `mvn versions:use-latest-releases` |
| Python | requirements.txt | pip | `pip list --outdated`, `pip-compile --upgrade`, `pip install -r` |
| Go | go.mod | Go modules | `go list -u -m all`, `go get -u`, `go mod tidy` |
| C#/.NET | .csproj | NuGet | `dotnet list package --outdated`, `dotnet add package`, `dotnet restore` |

## Installation

### Prerequisites

Make sure you have Go 1.25 or later installed.

### Build from source

```bash
cd dependency-manager
go build -o depman
```

### Install globally

```bash
go install
```

## Usage

### Basic Commands

#### Check for updates (Dry Run)

Check for available dependency updates without making any changes:

```bash
depman check --path /path/to/project
```

Or use the current directory:

```bash
depman check
```

#### Update dependency files

Update dependency management files to the latest versions without installing:

```bash
depman update --path /path/to/project
```

#### Full update and install

Update dependency files and install the new dependencies:

```bash
depman install --path /path/to/project
```

### Flags

- `-p, --path`: Starting filepath or directory to scan (default: current directory)
- `--direct-only`: Only check direct dependencies (excludes indirect/dev dependencies)
- `--ignore`: Additional directory names to ignore during scanning (can be specified multiple times)

### Default Ignored Directories

The following directories are always ignored when scanning recursively:
- `node_modules` - npm packages
- `.git` - Git repository data
- `vendor` - Go/PHP vendor directories
- `target` - Maven/Rust build output
- `dist` - Distribution/build output
- `build` - Build output

### Examples

#### Check a single dependency file

```bash
depman check --path ./package.json
```

#### Scan entire project

```bash
depman check --path ./my-project
```

#### Update all dependencies in a monorepo

```bash
depman install --path ./monorepo
```

#### Check only direct dependencies

For Go modules, this excludes indirect dependencies. For npm, this excludes devDependencies:

```bash
depman check --path ./my-project --direct-only
```

#### Ignore additional directories

Ignore custom directories in addition to the default ignored directories:

```bash
depman check --path ./my-project --ignore .cache --ignore tmp
```

## How It Works

1. **Scanning**: The tool recursively scans the specified path for dependency management files
2. **Detection**: Identifies file types (package.json, pom.xml, etc.)
3. **Checking**: Uses the appropriate package manager to check for updates
4. **Updating**: Based on the command, either:
   - Shows available updates (check)
   - Updates the dependency file (update)
   - Updates and installs dependencies (install)

## Special Considerations

### npm (package.json)

- Requires `npm-check-updates` (ncu) for updating: `npm install -g npm-check-updates`
- Uses `npm outdated` for checking updates
- With `--direct-only`: excludes devDependencies (only checks/updates production dependencies)

### Maven (pom.xml)

- Uses Maven versions plugin
- Creates backup files (automatically cleaned up)

### pip (requirements.txt)

- Requires `pip-tools` for updating: `pip install pip-tools`
- Uses `pip list --outdated` for checking

### Go modules (go.mod)

- Uses native Go commands
- Automatically runs `go mod tidy` after updates
- With `--direct-only`: excludes indirect dependencies (only checks/updates direct dependencies)

### NuGet (.csproj)

- Uses `dotnet` CLI
- Runs `dotnet restore` and `dotnet build` for full updates

## Output Example

```
Found 3 dependency management file(s):

Checking ./frontend/package.json (package.json)...
  Found 5 update(s):
  Package              Current  Latest   Type
  -------              -------  ------   ----
  react                18.2.0   18.3.1   minor
  typescript           5.0.4    5.3.3    minor
  @types/react         18.2.0   18.2.48  patch
  eslint               8.45.0   8.56.0   minor
  vite                 4.4.5    5.0.10   major

Checking ./backend/go.mod (go.mod)...
  Found 2 update(s):
  Package                          Current  Latest   Type
  -------                          -------  ------   ----
  github.com/spf13/cobra           v1.7.0   v1.8.0   minor
  github.com/stretchr/testify      v1.8.4   v1.9.0   minor

Checking ./api/pom.xml (pom.xml)...
  All dependencies are up to date!
```

## Error Handling

The tool will:
- Skip files if the required package manager is not installed
- Continue processing other files if one fails
- Display clear error messages for troubleshooting

## Development

### Project Structure

```
dependency-manager/
├── cmd/                    # Cobra commands
│   ├── root.go            # Root command
│   ├── check.go           # Check command
│   ├── update.go          # Update command
│   └── install.go         # Install command
├── internal/
│   ├── scanner/           # File scanning logic
│   │   └── scanner.go
│   └── checker/           # Dependency checkers
│       ├── checker.go     # Interface and registry
│       ├── npm.go         # npm checker
│       ├── maven.go       # Maven checker
│       ├── pip.go         # pip checker
│       ├── gomod.go       # Go modules checker
│       └── nuget.go       # NuGet checker
├── main.go
├── go.mod
└── README.md
```

### Adding a New Package Manager

1. Create a new checker in `internal/checker/`
2. Implement the `Checker` interface
3. Register the checker in `cmd/check.go`, `cmd/update.go`, and `cmd/install.go`

## License

See LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

