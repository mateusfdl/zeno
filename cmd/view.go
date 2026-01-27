package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mateusfdl/zeno/bench"
	tui "github.com/mateusfdl/zeno/views/terminal"
	"github.com/mateusfdl/zeno/views/web"
	flag "github.com/spf13/pflag"
)

type ViewCommand struct {
	fs        *flag.FlagSet
	filePath  string
	compare   string
	threshold float64
	web       bool
	webOutput string
}

func NewViewCommand() *ViewCommand {
	vc := &ViewCommand{
		fs: flag.NewFlagSet("view", flag.ExitOnError),
	}

	vc.fs.StringVarP(&vc.filePath, "file", "f", "", "JSON file to view (default: stdin)")
	vc.fs.StringVarP(&vc.compare, "compare", "c", "", "Compare with this file (enables comparison mode)")
	vc.fs.Float64VarP(&vc.threshold, "threshold", "t", 5.0, "Regression threshold percentage")
	vc.fs.BoolVarP(&vc.web, "web", "w", false, "Generate HTML report instead of TUI")
	vc.fs.StringVarP(&vc.webOutput, "output", "o", "bench-report.html", "Output file for HTML report")

	return vc
}

func (vc *ViewCommand) Run(args []string) error {
	if err := vc.fs.Parse(args); err != nil {
		return err
	}

	if vc.web {
		if vc.compare != "" {
			return vc.runWebComparison()
		} else if vc.filePath != "" {
			return vc.runWebSingleFile()
		} else {
			return vc.runWebStdin()
		}
	}

	if vc.compare != "" {
		return vc.runComparison()
	} else if vc.filePath != "" {
		return vc.runSingleFile()
	} else {
		return vc.runStdin()
	}
}

func (vc *ViewCommand) runComparison() error {
	results, err := bench.CompareTwoFiles(vc.filePath, vc.compare)
	if err != nil {
		return fmt.Errorf("error comparing files: %w", err)
	}

	model := tui.NewComparisonModel(results, vc.threshold)
	p := runTea(model)

	_, err = p.Run()
	return err
}

func (vc *ViewCommand) runSingleFile() error {
	runs, err := bench.ReadRuns(vc.filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	if len(runs) == 0 {
		return fmt.Errorf("no benchmark runs found in file")
	}

	model := tui.NewModel(runs, vc.threshold)
	p := runTea(model)

	_, err = p.Run()
	return err
}

func (vc *ViewCommand) runStdin() error {
	stat, _ := os.Stdin.Stat()

	if stat.Mode()&os.ModeCharDevice == 0 {
		return vc.runStreamingStdin()
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("error reading stdin: %w", err)
	}

	runs, err := bench.DecodeRuns(strings.NewReader(string(data)))
	if err == nil && len(runs) > 0 {
		model := tui.NewModel(runs, vc.threshold)
		p := runTea(model)
		_, err = p.Run()
		return err
	}

	parser := bench.NewParser()
	suites, err := parser.ParseBytes(data)
	if err != nil {
		return fmt.Errorf("error parsing benchmark output: %w", err)
	}

	run := bench.CreateRun(suites, "", 0, nil)
	model := tui.NewModel([]bench.Run{run}, vc.threshold)
	p := runTea(model)

	_, err = p.Run()
	return err
}

func (vc *ViewCommand) runStreamingStdin() error {
	model := tui.NewStreamingModel(vc.threshold)
	p := runTea(model)

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			p.Send(tui.BenchmarkLineMsg{Line: scanner.Text()})
		}
		p.Send(tui.StreamDoneMsg{Err: scanner.Err()})
	}()

	_, err := p.Run()
	return err
}

func runTea(model tui.Model) *tea.Program {
	return tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
}

func (vc *ViewCommand) runWebComparison() error {
	results, err := bench.CompareTwoFiles(vc.filePath, vc.compare)
	if err != nil {
		return fmt.Errorf("error comparing files: %w", err)
	}

	generator := web.NewComparisonGenerator(results, vc.threshold)
	fmt.Printf("Generating HTML report: %s\n", vc.webOutput)
	return generator.GenerateToFileAndOpen(vc.webOutput)
}

func (vc *ViewCommand) runWebSingleFile() error {
	runs, err := bench.ReadRuns(vc.filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	if len(runs) == 0 {
		return fmt.Errorf("no benchmark runs found in file")
	}

	generator := web.NewGenerator(runs, vc.threshold)
	fmt.Printf("Generating HTML report: %s\n", vc.webOutput)
	return generator.GenerateToFileAndOpen(vc.webOutput)
}

func (vc *ViewCommand) runWebStdin() error {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("error reading stdin: %w", err)
	}

	runs, err := bench.DecodeRuns(strings.NewReader(string(data)))
	if err == nil && len(runs) > 0 {
		generator := web.NewGenerator(runs, vc.threshold)
		fmt.Printf("Generating HTML report: %s\n", vc.webOutput)
		return generator.GenerateToFileAndOpen(vc.webOutput)
	}

	parser := bench.NewParser()
	suites, err := parser.ParseBytes(data)
	if err != nil {
		return fmt.Errorf("error parsing benchmark output: %w", err)
	}

	run := bench.CreateRun(suites, "", 0, nil)
	generator := web.NewGenerator([]bench.Run{run}, vc.threshold)
	fmt.Printf("Generating HTML report: %s\n", vc.webOutput)
	return generator.GenerateToFileAndOpen(vc.webOutput)
}

func (vc *ViewCommand) Usage() string {
	return `Usage: zeno view [options]

View benchmark results in a TUI or generate HTML report.

Displays benchmark results in an interactive terminal UI or as a static HTML report.
Can view single files, compare two files, or pipe from go test.

Examples:
  # View a JSON file in TUI
  zeno view -f results.json

  # Compare two files in TUI
  zeno view -f current.json --compare baseline.json

  # Generate HTML report
  eno view --web -f results.json

  # Generate HTML comparison
  zeno view --web -f current.json --compare baseline.json -o compare.html

  # Pipe from go test to HTML
  go test -bench=. -benchmem | zeno view --web

  # Save and view
  go test -bench=. | zeno parse | zeno view --web

Options:`
}
