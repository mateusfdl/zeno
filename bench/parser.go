package bench

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Parser struct {
	goVersion string
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(r io.Reader) ([]Suite, error) {
	br := bufio.NewReader(r)

	var suites []Suite

	for {
		line, isPrefix, err := br.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if isPrefix {
			return nil, fmt.Errorf("line too long")
		}

		lineStr := string(line)

		if strings.HasPrefix(lineStr, "goos:") {
			suite, err := p.readBenchmarkSuite(br, lineStr)
			if err != nil {
				return nil, err
			}
			suites = append(suites, *suite)
		}
	}

	return suites, nil
}

func (p *Parser) ParseBytes(data []byte) ([]Suite, error) {
	return p.Parse(strings.NewReader(string(data)))
}

func (p *Parser) ParseFile(path string) ([]Suite, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer f.Close()

	return p.Parse(f)
}

func (p *Parser) ParseStdin() ([]Suite, error) {
	return p.Parse(os.Stdin)
}

func (p *Parser) readBenchmarkSuite(br *bufio.Reader, firstLine string) (*Suite, error) {
	split := strings.SplitN(firstLine, ": ", 2)
	if len(split) != 2 {
		return nil, fmt.Errorf("invalid goos line: %s", firstLine)
	}

	suite := Suite{
		Goos:       strings.TrimSpace(split[1]),
		Benchmarks: make([]Benchmark, 0),
	}

	if p.goVersion != "" {
		suite.Go = p.goVersion
	}

	for {
		line, isPrefix, err := br.ReadLine()
		if err == io.EOF {
			return &suite, nil
		}
		if err != nil {
			return nil, err
		}
		if isPrefix {
			return nil, fmt.Errorf("line too long")
		}

		lineStr := string(line)

		if strings.HasPrefix(lineStr, "PASS") || strings.HasPrefix(lineStr, "FAIL") || strings.HasPrefix(lineStr, "ok") {
			break
		}

		if strings.HasPrefix(lineStr, "goarch:") {
			split := strings.SplitN(lineStr, ": ", 2)
			if len(split) == 2 {
				suite.Goarch = strings.TrimSpace(split[1])
			}
			continue
		}

		if strings.HasPrefix(lineStr, "pkg:") {
			split := strings.SplitN(lineStr, ": ", 2)
			if len(split) == 2 {
				suite.Pkg = strings.TrimSpace(split[1])
			}
			continue
		}

		if strings.HasPrefix(lineStr, "Benchmark") {
			bench, err := p.parseBenchmark(lineStr)
			if err != nil {
				return nil, fmt.Errorf("%w: %q", err, lineStr)
			}
			suite.Benchmarks = append(suite.Benchmarks, *bench)
		}
	}

	return &suite, nil
}

func (p *Parser) parseBenchmark(line string) (*Benchmark, error) {

	parts := strings.Split(line, "\t")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid benchmark format: expected at least 3 fields, got %d", len(parts))
	}

	bench := &Benchmark{
		Name:   strings.TrimSpace(parts[0]),
		Custom: make(map[string]float64),
	}

	runs, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%s: could not parse runs: %w", bench.Name, err)
	}
	bench.Runs = runs

	for i := 2; i < len(parts); i++ {
		if err := p.parseMetric(bench, strings.TrimSpace(parts[i])); err != nil {
			return nil, err
		}
	}

	return bench, nil
}

func (p *Parser) parseMetric(bench *Benchmark, metric string) error {

	parts := strings.Fields(metric)
	if len(parts) < 2 {
		return fmt.Errorf("%s: invalid metric format: %s", bench.Name, metric)
	}

	value, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return fmt.Errorf("%s: could not parse value: %w", bench.Name, err)
	}

	unit := parts[1]

	switch unit {
	case "ns/op":
		bench.NsPerOp = value
	case "B/op":
		if bench.Mem == nil {
			bench.Mem = &Mem{}
		}
		bench.Mem.BytesPerOp = value
	case "allocs/op":
		if bench.Mem == nil {
			bench.Mem = &Mem{}
		}
		bench.Mem.AllocsPerOp = value
	case "MB/s":
		if bench.Mem == nil {
			bench.Mem = &Mem{}
		}
		bench.Mem.MBPerSec = value
	default:

		bench.Custom[unit] = value
	}

	return nil
}

func (p *Parser) SetGoVersion(version string) {
	p.goVersion = version
}
