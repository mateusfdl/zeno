package cmd

import (
	"fmt"
	"os"

	"github.com/mateusfdl/zeno/bench"
	flag "github.com/spf13/pflag"
)

type MergeCommand struct {
	fs       *flag.FlagSet
	output   string
	sortDesc bool
	unique   bool
}

func NewMergeCommand() *MergeCommand {
	mc := &MergeCommand{
		fs: flag.NewFlagSet("merge", flag.ExitOnError),
	}

	mc.fs.StringVarP(&mc.output, "output", "o", "", "Output file path (default: stdout)")
	mc.fs.BoolVar(&mc.sortDesc, "sort-desc", false, "Sort by date descending (newest first)")
	mc.fs.BoolVar(&mc.sortDesc, "d", false, "Shorthand for --sort-desc")
	mc.fs.BoolVar(&mc.unique, "unique", false, "Remove duplicate runs")

	return mc
}

func (mc *MergeCommand) Run(args []string) error {
	if err := mc.fs.Parse(args); err != nil {
		return err
	}

	remaining := mc.fs.Args()
	if len(remaining) < 1 {
		return fmt.Errorf("merge requires at least one input file")
	}

	runs, err := bench.MergeRunsFromFiles(remaining...)
	if err != nil {
		return fmt.Errorf("error merging runs: %w", err)
	}

	if len(runs) == 0 {
		return fmt.Errorf("no runs found in input files")
	}

	if mc.unique {
		runs = bench.DeduplicateRuns(runs)
	}

	if mc.sortDesc {
		bench.SortByDateDescending(runs)
	} else {
		bench.SortByDate(runs)
	}

	if mc.output != "" {
		if err := bench.WriteRuns(mc.output, runs); err != nil {
			return fmt.Errorf("error writing output: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Merged %d runs to %s\n", len(runs), mc.output)
	} else {
		if err := bench.EncodeRuns(os.Stdout, runs); err != nil {
			return fmt.Errorf("error encoding output: %w", err)
		}
	}

	return nil
}

func (mc *MergeCommand) Usage() string {
	return `Usage: mygobenchtool merge [options] <file1.json> [file2.json ...]

Merge multiple benchmark JSON files into one.

Combines benchmark runs from multiple files into a single JSON file.
Can sort and deduplicate runs.

Examples:
  mygobenchtool merge -o combined.json file1.json file2.json file3.json
  mygobenchtool merge --unique --sort-desc -o all.json *.json
  mygobenchtool merge before.json after.json

Options:`
}
