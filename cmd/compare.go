package cmd

import (
	"fmt"

	"github.com/mateusfdl/zeno/bench"
	flag "github.com/spf13/pflag"
)

type CompareCommand struct {
	fs        *flag.FlagSet
	threshold float64
	format    string
}

func NewCompareCommand() *CompareCommand {
	cc := &CompareCommand{
		fs: flag.NewFlagSet("compare", flag.ExitOnError),
	}

	cc.fs.Float64VarP(&cc.threshold, "threshold", "t", 5.0, "Regression threshold percentage")
	cc.fs.StringVarP(&cc.format, "format", "f", "table", "Output format: table or json")

	return cc
}

func (cc *CompareCommand) Run(args []string) error {
	if err := cc.fs.Parse(args); err != nil {
		return err
	}

	remaining := cc.fs.Args()
	if len(remaining) != 2 {
		return fmt.Errorf("compare requires exactly 2 file arguments (before and after)")
	}

	beforePath := remaining[0]
	afterPath := remaining[1]

	results, err := bench.CompareTwoFiles(beforePath, afterPath)
	if err != nil {
		return fmt.Errorf("error comparing benchmarks: %w", err)
	}

	switch cc.format {
	case "table":
		output := bench.FormatComparisonResults(results, cc.threshold)
		fmt.Println(output)
	case "json":
		output := bench.FormatComparisonAsJSON(results)
		fmt.Println(output)
	default:
		return fmt.Errorf("unknown format: %s (use 'table' or 'json')", cc.format)
	}

	return nil
}

func (cc *CompareCommand) Usage() string {
	return `Usage: zeno compare [options] <before.json> <after.json>

Compare two benchmark runs and detect performance regressions.

Compares benchmark metrics between two runs and calculates percentage changes.
Reports regressions exceeding the threshold.

Examples:
  zeno compare baseline.json current.json
  zeno compare --threshold=2.5 before.json after.json
  zeno compare --format=json old.json new.json

Options:`
}
