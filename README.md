# cb - Command Output Clipboard Manager

A lightweight, self-contained command-line tool for Linux that captures command output, saves it to a file, and copies it to your clipboard.

Mainly used in testing [Aura](https://github.com/AuraNull)!

## Features

- 📋 Run any command and copy its output to clipboard
- 💾 Save output to file with timestamp headers
- 🎯 Support for both X11 (xclip/xsel) and Wayland (wl-copy)
- 🔧 Flexible filtering (head/tail lines)
- 📝 Multiple output modes (quiet, verbose, raw)
- ⚡ Zero dependencies - uses only Go standard library

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/requiem-eco/clipboard-go.git
cd clipboard-go

# Build the binary
# Replace "cb" with whatever you want to call the binary, e.g.: xx, c, cbb, xcb, cbx, etc
go build -o cb .

# Install to /usr/local/bin (optional)
sudo cp cb /usr/local/bin/
```

### From Release

Download the latest pre-built binary from the [Releases](https://github.com/requiem-eco/clipboard-go/releases) page:

```bash
# Download for your architecture (example for amd64)
wget https://github.com/requiem-eco/clipboard-go/releases/latest/download/cb-linux-amd64

# Make it executable
chmod +x cb-linux-amd64

# Move to PATH (optional)
sudo mv cb-linux-amd64 /usr/local/bin/cb
```

### Prerequisites

For clipboard support, you need one of the following tools installed:
- **Wayland:** `wl-copy` (from wl-clipboard package)
- **X11:** `xclip` or `xsel`

```bash
# Fedora/RHEL
sudo dnf install wl-clipboard xclip

# Debian/Ubuntu
sudo apt install wl-clipboard xclip

# Arch
sudo pacman -S wl-clipboard xclip
```

## Usage

### Basic Syntax

```bash
cb [flags] <command> [args...]
```

### Quick Examples

```bash
# Copy command output to clipboard
cb echo "hello world"

# Copy directory listing
cb ls -la

# Copy only first 10 lines of dmesg
cb -h 10 dmesg

# Save to custom file and show verbose output
cb -f output.txt -v ps aux

# Capture stderr instead of stdout
cb -e gcc missing_file.c

# Quiet mode - copy but don't print
cb -q echo "silent copy"

# Show only 5 lines but copy everything
cb -q 5 find /var/log

# Append to file instead of overwriting
cb -a -f log.txt date

# File only, no clipboard
cb -n -f backup.txt cat important.txt
```

## Flags

| Flag | Description |
|------|-------------|
| `-h X` | Copy only the **head** X lines of output |
| `-t X` | Copy only the **tail** X lines of output |
| `-q [X]` | Suppress stdout entirely, or show only X lines |
| `-f FILE` | Write output to specific file instead of `/tmp/cb.txt` |
| `-c, --clear` | Clear clipboard before writing |
| `-e, --error` | Copy stderr instead of stdout |
| `-a, --append` | Append output to file instead of overwriting |
| `-n` | Disable clipboard, only save to file |
| `-v, --verbose` | Show what was copied / debug info |
| `--no-temp` | Skip writing to file entirely |
| `-r, --raw` | Preserve terminal formatting/ANSI codes |
| `--delay N` | Wait N seconds before copying output |
| `--trim` | Trim leading and trailing whitespace |
| `--version` | Show program version |
| `--help` | Display help message |

## Output Format

When saving to file, `cb` adds a timestamp header:

```
[2025-11-17 19:41:29 "echo Hello, World!"]:
Hello, World!
```

When using `-a/--append`, multiple runs create a log:

```
[2025-11-17 19:41:29 "echo First run"]:
First run

[2025-11-17 19:42:15 "echo Second run"]:
Second run
```

## Advanced Examples

### Filtering Large Outputs

```bash
# Get only last 20 lines of logs
cb -t 20 journalctl

# Get first 50 lines of a file
cb -h 50 cat large_file.txt
```

### Combining with Other Commands

```bash
# Copy git diff but don't clutter terminal
cb -q git diff

# Copy command output and append to daily log
cb -a -f ~/logs/daily-$(date +%Y-%m-%d).txt systemctl status
```

### Debugging

```bash
# Capture and see what was copied
cb -v -e make 2>&1

# Save compilation errors without clipboard
cb -n -f errors.log -e cargo build
```

### Preserving Formatting

```bash
# Keep colors and formatting
cb -r ls --color=always

# Copy formatted output with specific line limit
cb -r -h 30 bat colorful_file.rs
```

## Building from Source

### Requirements

- Go 1.21 or later
- Linux operating system

### Build Commands

```bash
# Standard build
go build -o cb .

# Optimized build (smaller binary)
go build -ldflags="-s -w" -o cb .

# Build for different architectures
GOOS=linux GOARCH=amd64 go build -o cb-amd64 .
GOOS=linux GOARCH=arm64 go build -o cb-arm64 .
```

## Development

### Running Tests

```bash
# Run Go tests
go test -v ./...

# Run formatting check
gofmt -s -l .

# Run go vet
go vet ./...
```

### Project Structure

```
cb/
├── main.go              # Main application
├── go.mod               # Go module definition
├── .github/
│   └── workflows/
│       ├── ci.yml       # Continuous integration
│       └── release.yml  # Release automation
├── .gitignore
├── LICENSE
├── README.md
└── CLAUDE.md           # Development instructions
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

See [LICENSE](LICENSE) file for details.

## Version

Current version: **00.00.001**

Check version:
```bash
cb --version
```
