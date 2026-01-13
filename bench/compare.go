package bench

import (
	"fmt"
	"math"
	"strings"
)

func CompareTwoRuns(before, after Run) ([]ComparisonResult, error) {
	if len(before.Suites) != len(after.Suites) {
		return nil, fmt.Errorf("number of suites mismatch: %d vs %d", len(before.Suites), len(after.Suites))
	}

	var results []ComparisonResult

	for i := 0; i < len(before.Suites); i++ {
		beforeSuite := before.Suites[i]
		afterSuite := after.Suites[i]

		if beforeSuite.Pkg != afterSuite.Pkg {
			return nil, fmt.Errorf("suite package mismatch at index %d: %s vs %s", i, beforeSuite.Pkg, afterSuite.Pkg)
		}

		suiteResults := compareSuites(beforeSuite, afterSuite)
		results = append(results, suiteResults...)
	}

	return results, nil
}

func CompareTwoFiles(beforePath, afterPath string) ([]ComparisonResult, error) {
	beforeRuns, err := ReadRuns(beforePath)
	if err != nil {
		return nil, fmt.Errorf("error reading before file: %w", err)
	}

	afterRuns, err := ReadRuns(afterPath)
	if err != nil {
		return nil, fmt.Errorf("error reading after file: %w", err)
	}

	if len(beforeRuns) == 0 {
		return nil, fmt.Errorf("no runs in before file")
	}

	if len(afterRuns) == 0 {
		return nil, fmt.Errorf("no runs in after file")
	}

	return CompareTwoRuns(beforeRuns[0], afterRuns[0])
}

func compareSuites(before, after Suite) []ComparisonResult {
	var results []ComparisonResult

	afterMap := make(map[string]Benchmark)
	for _, b := range after.Benchmarks {
		afterMap[b.Name] = b
	}

	for _, beforeBench := range before.Benchmarks {
		afterBench, ok := afterMap[beforeBench.Name]
		if !ok {

			continue
		}

		result := ComparisonResult{
			Name:       fmt.Sprintf("%s/%s", before.Pkg, beforeBench.Name),
			OldRuns:    beforeBench.Runs,
			NewRuns:    afterBench.Runs,
			OldNsPerOp: beforeBench.NsPerOp,
			NewNsPerOp: afterBench.NsPerOp,
		}

		if beforeBench.NsPerOp > 0 {
			result.NsPerOpDiff = afterBench.NsPerOp - beforeBench.NsPerOp
			result.NsPerOpPct = (result.NsPerOpDiff / beforeBench.NsPerOp) * 100
		}

		if beforeBench.Mem != nil && afterBench.Mem != nil {
			result.OldBytes = beforeBench.Mem.BytesPerOp
			result.NewBytes = afterBench.Mem.BytesPerOp

			if beforeBench.Mem.BytesPerOp > 0 {
				result.BytesDiff = afterBench.Mem.BytesPerOp - beforeBench.Mem.BytesPerOp
				result.BytesPct = (result.BytesDiff / beforeBench.Mem.BytesPerOp) * 100
			}

			result.OldAllocs = beforeBench.Mem.AllocsPerOp
			result.NewAllocs = afterBench.Mem.AllocsPerOp

			if beforeBench.Mem.AllocsPerOp > 0 {
				result.AllocsDiff = afterBench.Mem.AllocsPerOp - beforeBench.Mem.AllocsPerOp
				result.AllocsPct = (result.AllocsDiff / beforeBench.Mem.AllocsPerOp) * 100
			}
		}

		results = append(results, result)
	}

	return results
}

func FormatComparisonResults(results []ComparisonResult, threshold float64) string {
	var sb strings.Builder

	sb.WriteString("Benchmark Comparison Results:\n")
	sb.WriteString(strings.Repeat("=", 120) + "\n\n")

	if len(results) == 0 {
		sb.WriteString("No benchmarks to compare.\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("%-50s %12s %12s %10s | %10s %10s %10s\n",
		"Benchmark", "Time Old", "Time New", "Time Δ%", "Mem Old", "Mem New", "Mem Δ%"))
	sb.WriteString(strings.Repeat("-", 120) + "\n")

	regressions := 0
	improvements := 0

	for _, r := range results {

		timeDelta := formatDelta(r.NsPerOpDiff, r.NsPerOpPct)
		sb.WriteString(fmt.Sprintf("%-50s %12.0f %12.0f %10s | ",
			truncateString(r.Name, 50),
			r.OldNsPerOp,
			r.NewNsPerOp,
			timeDelta))

		if r.OldBytes > 0 || r.NewBytes > 0 {
			memDelta := formatDelta(r.BytesDiff, r.BytesPct)
			sb.WriteString(fmt.Sprintf("%10.0f %10.0f %10s\n",
				r.OldBytes,
				r.NewBytes,
				memDelta))
		} else {
			sb.WriteString(fmt.Sprintf("%10s %10s %10s\n", "-", "-", "-"))
		}

		if r.IsRegression(threshold) {
			regressions++
		} else if r.NsPerOpPct < -threshold || r.BytesPct < -threshold {
			improvements++
		}
	}

	sb.WriteString(strings.Repeat("-", 120) + "\n")
	sb.WriteString(fmt.Sprintf("\nSummary: %d benchmarks compared", len(results)))

	if regressions > 0 {
		sb.WriteString(fmt.Sprintf(", %d REGRESSIONS detected (threshold: %.1f%%)", regressions, threshold))
	}
	if improvements > 0 {
		sb.WriteString(fmt.Sprintf(", %d improvements", improvements))
	}
	sb.WriteString("\n")

	return sb.String()
}

func formatDelta(diff, pct float64) string {
	if math.Abs(pct) < 0.01 {
		return "~0%"
	}

	sign := "+"
	if pct < 0 {
		sign = ""
	}
	return fmt.Sprintf("%s%.1f%%", sign, pct)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func FormatComparisonAsJSON(results []ComparisonResult) string {

	var sb strings.Builder
	sb.WriteString("[\n")

	for i, r := range results {
		if i > 0 {
			sb.WriteString(",\n")
		}
		sb.WriteString(fmt.Sprintf("  {\n"))
		sb.WriteString(fmt.Sprintf("    \"name\": \"%s\",\n", r.Name))
		sb.WriteString(fmt.Sprintf("    \"oldNsPerOp\": %.2f,\n", r.OldNsPerOp))
		sb.WriteString(fmt.Sprintf("    \"newNsPerOp\": %.2f,\n", r.NewNsPerOp))
		sb.WriteString(fmt.Sprintf("    \"nsPerOpChange\": %.2f,\n", r.NsPerOpPct))
		sb.WriteString(fmt.Sprintf("    \"oldBytesPerOp\": %.2f,\n", r.OldBytes))
		sb.WriteString(fmt.Sprintf("    \"newBytesPerOp\": %.2f,\n", r.NewBytes))
		sb.WriteString(fmt.Sprintf("    \"bytesPerOpChange\": %.2f\n", r.BytesPct))
		sb.WriteString(fmt.Sprintf("  }"))
	}

	sb.WriteString("\n]\n")
	return sb.String()
}
