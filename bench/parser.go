package bench

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unsafe"
)

var (
	prefixGoos      = []byte("goos:")
	prefixGoarch    = []byte("goarch:")
	prefixPkg       = []byte("pkg:")
	prefixBenchmark = []byte("Benchmark")
	prefixPASS      = []byte("PASS")
	prefixFAIL      = []byte("FAIL")
	prefixOk        = []byte("ok")
)

type Parser struct {
	goVersion string
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(r io.Reader) ([]Suite, error) {
	br := bufio.NewReader(r)

	suites := make([]Suite, 0, 4)

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

		if len(line) > 0 && line[0] == 'g' && bytes.HasPrefix(line, prefixGoos) {
			suite, err := p.readBenchmarkSuite(br, line)
			if err != nil {
				return nil, err
			}
			suites = append(suites, *suite)
		}
	}

	return suites, nil
}

func (p *Parser) ParseBytes(data []byte) ([]Suite, error) {
	return p.Parse(bytes.NewReader(data))
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

func (p *Parser) readBenchmarkSuite(br *bufio.Reader, firstLine []byte) (*Suite, error) {
	lineStr := bytesToString(firstLine)
	_, value, found := strings.Cut(lineStr, ": ")
	if !found {
		return nil, fmt.Errorf("invalid goos line: %s", lineStr)
	}

	suite := Suite{
		Goos:       strings.TrimSpace(value),
		Benchmarks: make([]Benchmark, 0, 32),
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

		if len(line) == 0 {
			continue
		}

		switch line[0] {
		case 'P':
			if bytes.HasPrefix(line, prefixPASS) {
				return &suite, nil
			}
		case 'F':
			if bytes.HasPrefix(line, prefixFAIL) {
				return &suite, nil
			}
		case 'o':
			if bytes.HasPrefix(line, prefixOk) {
				return &suite, nil
			}
		case 'g':
			if bytes.HasPrefix(line, prefixGoarch) {
				lineStr := bytesToString(line)
				if _, value, found := strings.Cut(lineStr, ": "); found {
					suite.Goarch = strings.TrimSpace(value)
				}
			}
		case 'p':
			if bytes.HasPrefix(line, prefixPkg) {
				lineStr := bytesToString(line)
				if _, value, found := strings.Cut(lineStr, ": "); found {
					suite.Pkg = strings.TrimSpace(value)
				}
			}
		case 'B':
			if bytes.HasPrefix(line, prefixBenchmark) {
				lineStr := bytesToString(line)
				bench, err := p.parseBenchmark(lineStr)
				if err != nil {
					return nil, fmt.Errorf("%w: %q", err, lineStr)
				}
				suite.Benchmarks = append(suite.Benchmarks, *bench)
			}
		}
	}
}

func (p *Parser) parseBenchmark(line string) (*Benchmark, error) {
	parts := strings.Split(line, "\t")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid benchmark format: expected at least 3 fields, got %d", len(parts))
	}

	bench := &Benchmark{
		Name: strings.TrimSpace(parts[0]),
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
	valueStr, unit, found := strings.Cut(metric, " ")
	if !found {
		return fmt.Errorf("%s: invalid metric format: %s", bench.Name, metric)
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return fmt.Errorf("%s: could not parse value: %w", bench.Name, err)
	}

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
		if bench.Custom == nil {
			bench.Custom = make(map[string]float64, 4)
		}
		bench.Custom[unit] = value
	}

	return nil
}

func (p *Parser) SetGoVersion(version string) {
	p.goVersion = version
}

func bytesToString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(&b[0], len(b))
}
