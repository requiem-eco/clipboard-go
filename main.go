package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const version = "00.00.001"

type Config struct {
	headLines   int
	tailLines   int
	quietLines  int
	quiet       bool
	file        string
	clear       bool
	stderr      bool
	append      bool
	noClipboard bool
	verbose     bool
	noTemp      bool
	raw         bool
	delay       int
	trim        bool
	help        bool
	showVersion bool
}

// quietFlag implements flag.Value for optional integer flag
type quietFlag struct {
	lines *int
	set   *bool
}

func (q *quietFlag) String() string {
	if *q.set && *q.lines == 0 {
		return "true"
	}
	return fmt.Sprintf("%d", *q.lines)
}

func (q *quietFlag) Set(value string) error {
	*q.set = true
	if value == "" || value == "true" {
		*q.lines = 0
		return nil
	}
	var lines int
	_, err := fmt.Sscanf(value, "%d", &lines)
	if err != nil {
		return err
	}
	*q.lines = lines
	return nil
}

func (q *quietFlag) IsBoolFlag() bool {
	return true
}

func main() {
	config := parseFlags()

	if config.help {
		printHelp()
		os.Exit(0)
	}

	if config.showVersion {
		fmt.Printf("cb version %s\n", version)
		os.Exit(0)
	}

	// Get command arguments
	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no command specified")
		printHelp()
		os.Exit(1)
	}

	// Execute command and capture output
	output, err := executeCommand(args, config.stderr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}

	// Apply delay if specified
	if config.delay > 0 {
		if config.verbose {
			fmt.Printf("Waiting %d seconds...\n", config.delay)
		}
		time.Sleep(time.Duration(config.delay) * time.Second)
	}

	// Strip ANSI codes unless -r/--raw is specified
	if !config.raw {
		output = stripANSI(output)
	}

	// Apply trim if specified
	if config.trim {
		output = strings.TrimSpace(output)
	}

	// Apply head/tail filtering
	output = applyLineFilters(output, config.headLines, config.tailLines)

	// Prepare output with timestamp header
	cmdStr := strings.Join(args, " ")
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	header := fmt.Sprintf("[%s \"%s\"]:\n", timestamp, cmdStr)
	fullOutput := header + output + "\n"

	// Write to file unless --no-temp is specified
	if !config.noTemp {
		targetFile := config.file
		if targetFile == "" {
			targetFile = "/tmp/cb.txt"
		}

		if err := writeToFile(targetFile, fullOutput, config.append); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
			os.Exit(1)
		}

		if config.verbose {
			fmt.Printf("Output written to %s\n", targetFile)
		}
	}

	// Copy to clipboard unless -n is specified
	if !config.noClipboard {
		if config.clear {
			if err := clearClipboard(); err != nil && config.verbose {
				fmt.Fprintf(os.Stderr, "Warning: could not clear clipboard: %v\n", err)
			}
		}

		if err := copyToClipboard(output); err != nil {
			fmt.Fprintf(os.Stderr, "Error copying to clipboard: %v\n", err)
			os.Exit(1)
		}

		if config.verbose {
			lines := strings.Count(output, "\n") + 1
			fmt.Printf("Copied %d lines to clipboard\n", lines)
		}
	}

	// Print output to stdout unless quiet mode
	if config.quiet && config.quietLines == 0 {
		// Suppress all output
		return
	}

	printOutput := output
	if config.quietLines > 0 {
		printOutput = limitLines(output, config.quietLines)
	}

	if printOutput != "" {
		fmt.Print(printOutput)
		if !strings.HasSuffix(printOutput, "\n") {
			fmt.Println()
		}
	}
}

func parseFlags() Config {
	config := Config{}

	// Custom quiet flag
	qFlag := &quietFlag{lines: &config.quietLines, set: &config.quiet}

	flag.IntVar(&config.headLines, "h", 0, "Copy only the head of output by X lines")
	flag.IntVar(&config.tailLines, "t", 0, "Copy only the tail of output by X lines")
	flag.Var(qFlag, "q", "Suppress stdout entirely, or show only X lines")
	flag.StringVar(&config.file, "f", "", "Write output to specific file instead of /tmp/cb.txt")
	flag.BoolVar(&config.clear, "c", false, "Clear clipboard before writing")
	flag.BoolVar(&config.clear, "clear", false, "Clear clipboard before writing")
	flag.BoolVar(&config.stderr, "e", false, "Copy stderr instead of stdout")
	flag.BoolVar(&config.stderr, "error", false, "Copy stderr instead of stdout")
	flag.BoolVar(&config.append, "a", false, "Append output to file instead of overwriting")
	flag.BoolVar(&config.append, "append", false, "Append output to file instead of overwriting")
	flag.BoolVar(&config.noClipboard, "n", false, "Disable clipboard, only save to file")
	flag.BoolVar(&config.verbose, "v", false, "Show what was copied / debug info")
	flag.BoolVar(&config.verbose, "verbose", false, "Show what was copied / debug info")
	flag.BoolVar(&config.noTemp, "no-temp", false, "Skip writing to /tmp entirely")
	flag.BoolVar(&config.raw, "r", false, "Preserve terminal formatting/ANSI codes")
	flag.BoolVar(&config.raw, "raw", false, "Preserve terminal formatting/ANSI codes")
	flag.IntVar(&config.delay, "delay", 0, "Wait N seconds before copying output")
	flag.BoolVar(&config.trim, "trim", false, "Trim leading and trailing whitespace")
	flag.BoolVar(&config.help, "help", false, "Display usage instructions")
	flag.BoolVar(&config.showVersion, "version", false, "Show program version")

	flag.Parse()

	return config
}

func executeCommand(args []string, captureStderr bool) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout

	if captureStderr {
		cmd.Stderr = &stdout // Combine stderr with stdout
	} else {
		cmd.Stderr = &stderr
	}

	err := cmd.Run()
	if err != nil {
		// Still return the output even if command failed
		if captureStderr {
			return stdout.String(), nil
		}
		// If there's stderr output, include it in error context
		if stderr.Len() > 0 {
			return stdout.String(), fmt.Errorf("%v (stderr: %s)", err, stderr.String())
		}
		return stdout.String(), err
	}

	return stdout.String(), nil
}

func stripANSI(input string) string {
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	return ansiRegex.ReplaceAllString(input, "")
}

func applyLineFilters(output string, headLines, tailLines int) string {
	if headLines <= 0 && tailLines <= 0 {
		return output
	}

	lines := strings.Split(output, "\n")

	if headLines > 0 && headLines < len(lines) {
		lines = lines[:headLines]
	}

	if tailLines > 0 && tailLines < len(lines) {
		lines = lines[len(lines)-tailLines:]
	}

	return strings.Join(lines, "\n")
}

func limitLines(output string, maxLines int) string {
	if maxLines <= 0 {
		return output
	}

	lines := strings.Split(output, "\n")
	if len(lines) <= maxLines {
		return output
	}

	return strings.Join(lines[:maxLines], "\n")
}

func writeToFile(path string, content string, appendMode bool) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	flags := os.O_CREATE | os.O_WRONLY
	if appendMode {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	f, err := os.OpenFile(path, flags, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(content)
	return err
}

func detectClipboardTool() (string, []string) {
	// Try Wayland first (wl-copy)
	if _, err := exec.LookPath("wl-copy"); err == nil {
		return "wl-copy", []string{}
	}

	// Try X11 (xclip)
	if _, err := exec.LookPath("xclip"); err == nil {
		return "xclip", []string{"-selection", "clipboard"}
	}

	// Try xsel as fallback
	if _, err := exec.LookPath("xsel"); err == nil {
		return "xsel", []string{"--clipboard", "--input"}
	}

	return "", nil
}

func copyToClipboard(content string) error {
	tool, args := detectClipboardTool()
	if tool == "" {
		return fmt.Errorf("no clipboard tool found (tried: wl-copy, xclip, xsel)")
	}

	cmd := exec.Command(tool, args...)
	cmd.Stdin = strings.NewReader(content)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("clipboard command failed: %v (stderr: %s)", err, stderr.String())
	}

	return nil
}

func clearClipboard() error {
	tool, args := detectClipboardTool()
	if tool == "" {
		return fmt.Errorf("no clipboard tool found")
	}

	cmd := exec.Command(tool, args...)
	cmd.Stdin = strings.NewReader("")

	return cmd.Run()
}

func printHelp() {
	fmt.Println(`cb - Command output clipboard manager

Usage:
  cb [flags] <command> [args...]

Description:
  Run any command, capture its output, save to file, and copy to clipboard.

Flags:
  -h X              Copy only the head of output by X lines
  -t X              Copy only the tail of output by X lines
  -q [X]            Suppress stdout entirely, or show only X lines
  -f FILE           Write output to specific file instead of /tmp/cb.txt
  -c, --clear       Clear clipboard before writing
  -e, --error       Copy stderr instead of stdout
  -a, --append      Append output to file instead of overwriting
  -n                Disable clipboard, only save to file
  -v, --verbose     Show what was copied / debug info
  --no-temp         Skip writing to /tmp entirely
  -r, --raw         Preserve terminal formatting/ANSI codes
  --delay N         Wait N seconds before copying output
  --trim            Trim leading and trailing whitespace
  --version         Show program version
  --help            Display this help message

Examples:
  cb echo "hello world"
  cb ls -l /home/user
  cb -h 10 dmesg
  cb -f output.txt -v ps aux
  cb -e -v somecommand

Notes:
  - Requires wl-copy (Wayland) or xclip/xsel (X11) for clipboard support
  - Output is saved with timestamp header: [date time "command"]:`)
}
