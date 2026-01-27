package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/mateusfdl/zeno/bench"
	flag "github.com/spf13/pflag"
)

type ParseCommand struct {
	fs      *flag.FlagSet
	output  string
	version string
	tags    []string
	append  bool
	date    int64
}

func NewParseCommand() *ParseCommand {
	pc := &ParseCommand{
		fs: flag.NewFlagSet("parse", flag.ExitOnError),
	}

	pc.fs.StringVarP(&pc.output, "output", "o", "", "Output file path (default: stdout)")
	pc.fs.StringVar(&pc.version, "version", "", "Version tag for this run")
	pc.fs.StringSliceVar(&pc.tags, "tags", []string{}, "Tags to add to this run")
	pc.fs.BoolVar(&pc.append, "append", false, "Append to existing output file")
	pc.fs.Int64Var(&pc.date, "date", time.Now().Unix(), "Timestamp for this run (Unix timestamp)")

	return pc
}

func (pc *ParseCommand) Run(args []string) error {
	if err := pc.fs.Parse(args); err != nil {
		return err
	}

	p := bench.NewParser()
	suites, err := p.ParseStdin()
	if err != nil {
		return fmt.Errorf("error parsing benchmark: %w", err)
	}

	if len(suites) == 0 {
		return fmt.Errorf("no benchmark suites found")
	}

	run := bench.CreateRun(suites, pc.version, pc.date, pc.tags)

	if pc.output != "" {
		if err := bench.WriteRunToFile(pc.output, &run, pc.append); err != nil {
			return fmt.Errorf("error writing output: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Parsed %d benchmark suites to %s\n", len(suites), pc.output)
	} else {
		if err := bench.EncodeRuns(os.Stdout, []bench.Run{run}); err != nil {
			return fmt.Errorf("error encoding output: %w", err)
		}
	}

	return nil
}

func (pc *ParseCommand) Usage() string {
	return `Usage: zeno parse [options]

Parse benchmark output from stdin and output structured JSON.

Reads Go benchmark output from stdin and produces JSON format.
Can write to a file or stdout.

Examples:
  go test -bench=. -benchmem | ueno parse -o results.json
  go test -bench=. | zeno parse --version=v1.0.0 --tags=ci
  zeno parse --append -o history.json`
}
