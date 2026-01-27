package main

import (
	"fmt"
	"os"

	"github.com/mateusfdl/zeno/cmd"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	var commander interface{ Run([]string) error }

	switch command {
	case "parse":
		commander = cmd.NewParseCommand()
	case "merge":
		commander = cmd.NewMergeCommand()
	case "compare":
		commander = cmd.NewCompareCommand()
	case "view":
		commander = cmd.NewViewCommand()
	case "version", "--version", "-v":
		fmt.Printf("Zeno version %s\n", version)
		os.Exit(0)
	case "help", "--help", "-h":
		printUsage()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}

	if err := commander.Run(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf(`Zeno - Go Bench Analysis Tool v%s

A tool for parsing, storing, and analyzing Go benchmark results.

USAGE:
    zeno <command> [options]

COMMANDS:
    parse      Parse benchmark output from stdin and output JSON
    merge      Merge multiple benchmark JSON files
    compare    Compare two benchmark runs and detect regressions
    view       View benchmark results (TUI or HTML web report)
    version    Show version information
    help       Show this help message

EXAMPLES:
    # Parse benchmark output
    go test -bench=. -benchmem | zeno parse -o results.json

    # Parse with metadata
    go test -bench=. | zeno parse --version=v1.0.0 --tags=ci -o results.json

    # Merge benchmark files
    zeno merge -o combined.json file1.json file2.json file3.json

    # Compare benchmarks
    zeno compare baseline.json current.json

    # Compare with custom threshold
    zeno compare --threshold=2.5 before.json after.json

    # View in TUI
    zeno view -f results.json

    # Generate HTML web report
    zeno view --web -f results.json

    # Generate HTML comparison report
    zeno view --web -f current.json --compare baseline.json -o compare.html

    # Pipe from go test to HTML
    go test -bench=. -benchmem | Zeno view --web

Use "zeno <command> --help" for more information about a command.
`, version)
}
