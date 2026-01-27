package web

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/mateusfdl/zeno/bench"
)

type Generator struct {
	runs         []bench.Run
	comparison   []bench.ComparisonResult
	threshold    float64
	isComparison bool
	title        string
}

func NewGenerator(runs []bench.Run, threshold float64) *Generator {
	title := "Benchmark Results"
	if len(runs) > 0 && runs[0].Version != "" {
		title = fmt.Sprintf("Benchmarks - %s", runs[0].Version)
	}
	return &Generator{
		runs:      runs,
		threshold: threshold,
		title:     title,
	}
}

func NewComparisonGenerator(comparison []bench.ComparisonResult, threshold float64) *Generator {
	return &Generator{
		comparison:   comparison,
		threshold:    threshold,
		isComparison: true,
		title:        "Benchmark Comparison",
	}
}

func (g *Generator) GenerateToFile(outputPath string) error {
	html := g.Generate()

	dir := filepath.Dir(outputPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating directory: %w", err)
		}
	}

	if err := os.WriteFile(outputPath, []byte(html), 0644); err != nil {
		return fmt.Errorf("error writing HTML file: %w", err)
	}

	return nil
}

func (g *Generator) GenerateToFileAndOpen(outputPath string) error {
	if err := g.GenerateToFile(outputPath); err != nil {
		return err
	}

	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("error getting absolute path: %w", err)
	}

	if err := openBrowser(absPath); err != nil {
		return fmt.Errorf("error opening browser: %w", err)
	}

	return nil
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}
	args = append(args, url)

	return exec.Command(cmd, args...).Start()
}

func (g *Generator) Generate() string {
	var content string

	if g.isComparison {
		content = g.generateComparisonContent()
	} else {
		content = g.generateRunsContent()
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
%s
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>%s</h1>
            <p class="timestamp">Generated: %s</p>
        </header>
%s
    </div>
    <script>
%s
    </script>
</body>
</html>`, g.title, g.getCSS(), g.title, time.Now().Format("2006-01-02 15:04:05"), content, g.getJS())
}

func (g *Generator) generateRunsContent() string {
	var sections []string

	if len(g.runs) == 0 {
		return g.generateEmptyState("No benchmark data available")
	}

	sections = append(sections, g.generateSummary())

	sections = append(sections, g.generateTabs())

	return joinSections(sections...)
}

func (g *Generator) generateComparisonContent() string {
	var sections []string

	if len(g.comparison) == 0 {
		return g.generateEmptyState("No comparison data available")
	}

	sections = append(sections, g.generateComparisonSummary())

	sections = append(sections, g.generateComparisonCharts())

	sections = append(sections, g.generateComparisonTable())

	return joinSections(sections...)
}

func (g *Generator) generateSummary() string {
	run := g.runs[0]

	var metadata []string
	if run.Version != "" {
		metadata = append(metadata, fmt.Sprintf(`<div class="metadata-item">
            <span class="metadata-label">Version:</span>
            <span class="metadata-value">%s</span>
        </div>`, run.Version))
	}
	if run.Date > 0 {
		metadata = append(metadata, fmt.Sprintf(`<div class="metadata-item">
            <span class="metadata-label">Date:</span>
            <span class="metadata-value">%s</span>
        </div>`, time.Unix(run.Date, 0).Format("2006-01-02 15:04:05")))
	}
	if len(run.Tags) > 0 {
		tags := joinStrings(run.Tags, ", ")
		metadata = append(metadata, fmt.Sprintf(`<div class="metadata-item">
            <span class="metadata-label">Tags:</span>
            <span class="metadata-value">%s</span>
        </div>`, tags))
	}

	suiteCount := 0
	benchCount := 0
	for _, s := range run.Suites {
		suiteCount++
		benchCount += len(s.Benchmarks)
	}

	metadata = append(metadata, fmt.Sprintf(`<div class="metadata-item">
        <span class="metadata-label">Suites:</span>
        <span class="metadata-value">%d</span>
    </div>`, suiteCount))
	metadata = append(metadata, fmt.Sprintf(`<div class="metadata-item">
        <span class="metadata-label">Benchmarks:</span>
        <span class="metadata-value">%d</span>
    </div>`, benchCount))

	return fmt.Sprintf(`<section class="summary">
    <h2>Summary</h2>
    <div class="metadata">
%s
    </div>
</section>`, joinStrings(metadata, "\n"))
}

func (g *Generator) generateComparisonSummary() string {
	regressions := 0
	improvements := 0

	for _, r := range g.comparison {
		if r.IsRegression(g.threshold) {
			regressions++
		} else if r.NsPerOpPct < -g.threshold {
			improvements++
		}
	}

	return fmt.Sprintf(`<section class="summary">
    <h2>Comparison Summary</h2>
    <div class="stats">
        <div class="stat-card">
            <div class="stat-value">%d</div>
            <div class="stat-label">Total Benchmarks</div>
        </div>
        <div class="stat-card stat-danger">
            <div class="stat-value">%d</div>
            <div class="stat-label">Regressions</div>
        </div>
        <div class="stat-card stat-success">
            <div class="stat-value">%d</div>
            <div class="stat-label">Improvements</div>
        </div>
    </div>
</section>`, len(g.comparison), regressions, improvements)
}

func (g *Generator) generateTabs() string {
	var tabs []string
	var content []string

	for i, run := range g.runs {
		tabID := fmt.Sprintf("run-%d", i)
		tabs = append(tabs, fmt.Sprintf(`<button class="tab-btn %s" data-tab="%s">%s</button>`,
			map[bool]string{true: "active", false: ""}[i == 0],
			tabID,
			g.getRunTitle(run, i)))

		content = append(content, g.generateRunTab(run, i, i == 0))
	}

	return fmt.Sprintf(`<section class="tabs-section">
    <div class="tabs">%s</div>
    <div class="tab-content">%s</div>
</section>`, joinStrings(tabs, "\n"), joinStrings(content, "\n"))
}

func (g *Generator) getRunTitle(run bench.Run, index int) string {
	if run.Version != "" {
		return run.Version
	}
	if run.Date > 0 {
		return time.Unix(run.Date, 0).Format("2006-01-02 15:04")
	}
	return fmt.Sprintf("Run %d", index+1)
}

func (g *Generator) generateRunTab(run bench.Run, index int, isActive bool) string {
	var sections []string

	for _, suite := range run.Suites {
		sections = append(sections, g.generateSuiteCard(suite))
	}

	activeClass := map[bool]string{true: "active", false: ""}[isActive]

	return fmt.Sprintf(`<div id="run-%d" class="tab-pane %s">
%s
</div>`, index, activeClass, joinStrings(sections, "\n"))
}

func (g *Generator) generateSuiteCard(suite bench.Suite) string {
	timeChart := g.generateTimeBarChart(suite)
	memChart := g.generateMemBarChart(suite)

	return fmt.Sprintf(`<div class="card">
    <div class="card-header">
        <h3>%s</h3>
        <div class="suite-info">
            <span class="badge">%s</span>
            <span class="badge">%s/%s</span>
        </div>
    </div>
    <div class="card-body">
        <div class="chart-section">
            <h4>Execution Time (ns/op)</h4>
            <div class="bar-chart">
%s
            </div>
        </div>
        <div class="chart-section">
            <h4>Memory Usage (B/op)</h4>
            <div class="bar-chart">
%s
            </div>
        </div>
    </div>
</div>`, suite.Pkg, suite.Go, suite.Goos, suite.Goarch, timeChart, memChart)
}

func (g *Generator) generateTimeBarChart(suite bench.Suite) string {
	maxVal := 0.0
	for _, b := range suite.Benchmarks {
		if b.NsPerOp > maxVal {
			maxVal = b.NsPerOp
		}
	}
	if maxVal == 0 {
		return "<p class='no-data'>No timing data available</p>"
	}

	greenCount, yellowCount, redCount := 0, 0, 0

	var bars []string
	for _, b := range suite.Benchmarks {
		if b.NsPerOp <= 0 {
			continue
		}

		pct := (b.NsPerOp / maxVal) * 100
		if pct < 5 {
			pct = 5
		}
		color, shade := g.getPerformanceColor(b.NsPerOp, maxVal, &greenCount, &yellowCount, &redCount)

		bars = append(bars, fmt.Sprintf(`<div class="bar-row">
            <div class="bar-name">%s</div>
            <div class="bar-track">
                <div class="bar %s" style="width: %.1f%%; %s"></div>
            </div>
            <div class="bar-value">%s</div>
        </div>`, escapeHTML(b.Name), color, pct, shade, formatValue(b.NsPerOp)))
	}

	return joinStrings(bars, "\n")
}

func (g *Generator) generateMemBarChart(suite bench.Suite) string {
	maxVal := 0.0
	for _, b := range suite.Benchmarks {
		if b.Mem != nil && b.Mem.BytesPerOp > maxVal {
			maxVal = b.Mem.BytesPerOp
		}
	}
	if maxVal == 0 {
		return "<p class='no-data'>No memory data available</p>"
	}

	greenCount, yellowCount, redCount := 0, 0, 0

	var bars []string
	for _, b := range suite.Benchmarks {
		if b.Mem == nil || b.Mem.BytesPerOp <= 0 {
			continue
		}

		pct := (b.Mem.BytesPerOp / maxVal) * 100
		if pct < 5 {
			pct = 5
		}
		color, shade := g.getPerformanceColor(b.Mem.BytesPerOp, maxVal, &greenCount, &yellowCount, &redCount)

		bars = append(bars, fmt.Sprintf(`<div class="bar-row">
            <div class="bar-name">%s</div>
            <div class="bar-track">
                <div class="bar %s" style="width: %.1f%%; %s"></div>
            </div>
            <div class="bar-value">%s</div>
        </div>`, escapeHTML(b.Name), color, pct, shade, formatBytes(b.Mem.BytesPerOp)))
	}

	return joinStrings(bars, "\n")
}

func (g *Generator) getPerformanceColor(value, maxVal float64, greenCount, yellowCount, redCount *int) (string, string) {
	ratio := value / maxVal

	var color string
	var shadeIdx int

	if ratio <= 0.5 {
		color = "bar-fast"
		shadeIdx = *greenCount
		*greenCount++
	} else if ratio <= 0.8 {
		color = "bar-medium"
		shadeIdx = *yellowCount
		*yellowCount++
	} else {
		color = "bar-slow"
		shadeIdx = *redCount
		*redCount++
	}

	lightenPct := shadeIdx * 12
	if lightenPct > 48 {
		lightenPct = 48
	}

	shade := ""
	if lightenPct > 0 {
		shade = fmt.Sprintf("filter: brightness(%d%%);", 100+lightenPct)
	}

	return color, shade
}

func formatValue(v float64) string {
	if v >= 1000000 {
		return fmt.Sprintf("%.1fM", v/1000000)
	} else if v >= 1000 {
		return fmt.Sprintf("%.1fK", v/1000)
	} else if v >= 100 {
		return fmt.Sprintf("%.0f", v)
	} else if v >= 10 {
		return fmt.Sprintf("%.1f", v)
	}
	return fmt.Sprintf("%.2f", v)
}

func formatBytes(v float64) string {
	if v >= 1073741824 {
		return fmt.Sprintf("%.1f GB", v/1073741824)
	} else if v >= 1048576 {
		return fmt.Sprintf("%.1f MB", v/1048576)
	} else if v >= 1024 {
		return fmt.Sprintf("%.1f KB", v/1024)
	}
	return fmt.Sprintf("%.0f B", v)
}

func (g *Generator) generateBenchmarkRow(b bench.Benchmark) string {
	var metrics []string

	metrics = append(metrics, fmt.Sprintf(`<div class="metric">
        <span class="metric-label">Runs:</span>
        <span class="metric-value">%s</span>
    </div>`, formatNumber(b.Runs)))

	if b.NsPerOp > 0 {
		metrics = append(metrics, fmt.Sprintf(`<div class="metric">
            <span class="metric-label">Time:</span>
            <span class="metric-value">%.2f ns/op</span>
        </div>`, b.NsPerOp))
	}

	if b.Mem != nil {
		if b.Mem.BytesPerOp > 0 {
			metrics = append(metrics, fmt.Sprintf(`<div class="metric">
                <span class="metric-label">Memory:</span>
                <span class="metric-value">%.0f B/op</span>
            </div>`, b.Mem.BytesPerOp))
		}
		if b.Mem.AllocsPerOp > 0 {
			metrics = append(metrics, fmt.Sprintf(`<div class="metric">
                <span class="metric-label">Allocs:</span>
                <span class="metric-value">%.0f allocs/op</span>
            </div>`, b.Mem.AllocsPerOp))
		}
	}

	return fmt.Sprintf(`<div class="benchmark-item">
    <div class="benchmark-name">%s</div>
    <div class="benchmark-metrics">%s</div>
</div>`, b.Name, joinStrings(metrics, "\n"))
}

func (g *Generator) generateComparisonCharts() string {
	timeData := g.prepareTimeChartData()
	memData := g.prepareMemChartData()

	return fmt.Sprintf(`<section class="charts">
    <h2>Performance Changes</h2>
    <div class="chart-grid">
        <div class="chart-card">
            <h3>Execution Time Changes</h3>
            <canvas id="timeChart"></canvas>
        </div>
        <div class="chart-card">
            <h3>Memory Usage Changes</h3>
            <canvas id="memChart"></canvas>
        </div>
    </div>
</section>
<script>
window.timeData = %s;
window.memData = %s;
</script>`, timeData, memData)
}

func (g *Generator) prepareTimeChartData() string {
	var labels []string
	var data []string
	var colors []string

	for _, r := range g.comparison {
		labels = append(labels, fmt.Sprintf("'%s'", escapeJS(r.Name)))
		data = append(data, fmt.Sprintf("%.2f", r.NsPerOpPct))

		color := "#64748b"
		if r.NsPerOpPct > g.threshold {
			color = "#ef4444"
		} else if r.NsPerOpPct < -g.threshold {
			color = "#22c55e"
		}
		colors = append(colors, color)
	}

	return fmt.Sprintf(`{
    labels: [%s],
    datasets: [{
        label: 'Time Change %%',
        data: [%s],
        backgroundColor: [%s]
    }]
}`, joinStrings(labels, ", "), joinStrings(data, ", "), joinStrings(colors, ", "))
}

func (g *Generator) prepareMemChartData() string {
	var labels []string
	var data []string
	var colors []string

	for _, r := range g.comparison {
		if r.OldBytes > 0 || r.NewBytes > 0 {
			labels = append(labels, fmt.Sprintf("'%s'", escapeJS(r.Name)))
			data = append(data, fmt.Sprintf("%.2f", r.BytesPct))

			color := "#64748b"
			if r.BytesPct > g.threshold {
				color = "#ef4444"
			} else if r.BytesPct < -g.threshold {
				color = "#22c55e"
			}
			colors = append(colors, color)
		}
	}

	return fmt.Sprintf(`{
    labels: [%s],
    datasets: [{
        label: 'Memory Change %%',
        data: [%s],
        backgroundColor: [%s]
    }]
}`, joinStrings(labels, ", "), joinStrings(data, ", "), joinStrings(colors, ", "))
}

func (g *Generator) generateComparisonTable() string {
	var rows []string

	for _, r := range g.comparison {
		timeClass := getClassForChange(r.NsPerOpPct, g.threshold)
		memClass := getClassForChange(r.BytesPct, g.threshold)

		rows = append(rows, fmt.Sprintf(`<tr>
            <td class="bench-name">%s</td>
            <td class="text-right">%.0f</td>
            <td class="text-right">%.0f</td>
            <td class="text-right %s">%+.1f%%</td>
            <td class="text-right">%.0f</td>
            <td class="text-right">%.0f</td>
            <td class="text-right %s">%+.1f%%</td>
        </tr>`, escapeHTML(r.Name), r.OldNsPerOp, r.NewNsPerOp, timeClass, r.NsPerOpPct,
			r.OldBytes, r.NewBytes, memClass, r.BytesPct))
	}

	return fmt.Sprintf(`<section class="comparison-table">
    <h2>Detailed Comparison</h2>
    <div class="table-wrapper">
        <table>
            <thead>
                <tr>
                    <th>Benchmark</th>
                    <th class="text-right">Old (ns/op)</th>
                    <th class="text-right">New (ns/op)</th>
                    <th class="text-right">Time Δ%</th>
                    <th class="text-right">Old (B/op)</th>
                    <th class="text-right">New (B/op)</th>
                    <th class="text-right">Mem Δ%</th>
                </tr>
            </thead>
            <tbody>
%s
            </tbody>
        </table>
    </div>
</section>`, joinStrings(rows, "\n"))
}

func (g *Generator) generateEmptyState(message string) string {
	return fmt.Sprintf(`<div class="empty-state">
    <div class="empty-icon">📊</div>
    <h2>%s</h2>
    <p>Run benchmarks with 'go test -bench=.' to generate data</p>
</div>`, message)
}

func (g *Generator) getCSS() string {
	return `* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

:root {
    --bg-primary: #1a1a2e;
    --bg-secondary: #16213e;
    --bg-card: #1f2940;
    --bg-bar-track: #0d1321;
    --text-primary: #e8e8e8;
    --text-secondary: #a0aec0;
    --text-muted: #718096;
    --border: #2d3748;
    --gopher-cyan: #00ADD8;
    --gopher-blue: #5DC9E2;
    --gopher-dark: #007d9c;
    --fast: #00ADD8;
    --medium: #f6ad55;
    --slow: #fc8181;
}

body {
    font-family: 'JetBrains Mono', 'Fira Code', 'SF Mono', Consolas, monospace;
    background: var(--bg-primary);
    color: var(--text-primary);
    line-height: 1.5;
    min-height: 100vh;
}

.container {
    max-width: 1600px;
    margin: 0 auto;
    padding: 1.5rem 2rem;
}

header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;
    padding-bottom: 1rem;
    border-bottom: 2px solid var(--gopher-cyan);
}

header h1 {
    font-size: 1.5rem;
    font-weight: 600;
    color: var(--gopher-cyan);
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

header h1::before {
    content: ">";
    color: var(--gopher-blue);
}

.timestamp {
    color: var(--text-muted);
    font-size: 0.75rem;
}

section {
    margin-bottom: 1.5rem;
}

h2 {
    font-size: 1rem;
    margin-bottom: 0.75rem;
    color: var(--gopher-cyan);
    text-transform: uppercase;
    letter-spacing: 0.1em;
    font-weight: 600;
}

h3 {
    font-size: 0.95rem;
    color: var(--text-primary);
    margin-bottom: 0.5rem;
    font-weight: 500;
}

/* Summary */
.summary {
    background: var(--bg-secondary);
    border-radius: 6px;
    padding: 1rem;
    border: 1px solid var(--border);
}

.metadata {
    display: flex;
    flex-wrap: wrap;
    gap: 2rem;
}

.metadata-item {
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.metadata-label {
    color: var(--text-muted);
    font-size: 0.75rem;
}

.metadata-value {
    color: var(--gopher-cyan);
    font-weight: 600;
    font-size: 0.85rem;
}

/* Stats Cards */
.stats {
    display: flex;
    gap: 1rem;
    flex-wrap: wrap;
}

.stat-card {
    background: var(--bg-card);
    border-radius: 6px;
    padding: 1rem 1.5rem;
    text-align: center;
    border: 1px solid var(--border);
    min-width: 120px;
}

.stat-card.stat-danger {
    border-color: var(--slow);
}

.stat-card.stat-success {
    border-color: var(--gopher-cyan);
}

.stat-value {
    font-size: 1.5rem;
    font-weight: 700;
    margin-bottom: 0.25rem;
    color: var(--text-primary);
}

.stat-label {
    color: var(--text-muted);
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
}

/* Tabs */
.tabs-section {
    background: var(--bg-secondary);
    border-radius: 6px;
    border: 1px solid var(--border);
}

.tabs {
    display: flex;
    gap: 0;
    border-bottom: 1px solid var(--border);
    background: var(--bg-primary);
}

.tab-btn {
    background: transparent;
    border: none;
    color: var(--text-muted);
    padding: 0.6rem 1rem;
    cursor: pointer;
    border-bottom: 2px solid transparent;
    transition: all 0.2s;
    font-family: inherit;
    font-size: 0.8rem;
}

.tab-btn:hover {
    color: var(--text-primary);
}

.tab-btn.active {
    color: var(--gopher-cyan);
    border-bottom-color: var(--gopher-cyan);
    background: var(--bg-secondary);
}

.tab-content {
    padding: 1rem;
}

.tab-pane {
    display: none;
}

.tab-pane.active {
    display: block;
}

/* Cards */
.card {
    background: var(--bg-card);
    border-radius: 6px;
    margin-bottom: 1.5rem;
    border: 1px solid var(--border);
}

.card-header {
    padding: 0.75rem 1rem;
    background: var(--bg-secondary);
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 1px solid var(--border);
}

.card-header h3 {
    margin: 0;
    color: var(--gopher-cyan);
    font-size: 0.9rem;
    font-weight: 500;
}

.suite-info {
    display: flex;
    gap: 0.5rem;
}

.badge {
    background: var(--gopher-dark);
    color: white;
    padding: 0.2rem 0.6rem;
    border-radius: 3px;
    font-size: 0.7rem;
    font-weight: 500;
}

.card-body {
    padding: 1rem;
}

/* Benchmark List */
.benchmark-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.benchmark-item {
    background: var(--bg-secondary);
    border-radius: 6px;
    padding: 1rem;
}

.benchmark-name {
    font-weight: 600;
    color: var(--primary);
    margin-bottom: 0.75rem;
}

.benchmark-metrics {
    display: flex;
    flex-wrap: wrap;
    gap: 1rem;
}

.metric {
    display: flex;
    gap: 0.5rem;
}

.metric-label {
    color: var(--text-muted);
}

.metric-value {
    color: var(--text-primary);
    font-weight: 600;
}

/* Charts */
.charts {
    background: var(--bg-secondary);
    border-radius: 8px;
    padding: 1.5rem;
}

.chart-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
    gap: 1.5rem;
    margin-top: 1rem;
}

.chart-card {
    background: var(--bg-card);
    border-radius: 8px;
    padding: 1.5rem;
}

.chart-card h3 {
    margin-bottom: 1rem;
}

.chart-card canvas {
    max-height: 300px;
}

/* Comparison Table */
.comparison-table {
    background: var(--bg-secondary);
    border-radius: 8px;
    padding: 1.5rem;
}

.table-wrapper {
    overflow-x: auto;
}

table {
    width: 100%;
    border-collapse: collapse;
}

thead tr {
    border-bottom: 2px solid var(--border);
}

th {
    padding: 1rem;
    text-align: left;
    color: var(--text-secondary);
    font-weight: 600;
}

th.text-right {
    text-align: right;
}

td {
    padding: 1rem;
    border-bottom: 1px solid var(--border);
}

td.text-right {
    text-align: right;
}

.bench-name {
    color: var(--primary);
    font-weight: 500;
}

.text-right {
    text-align: right;
}

.change-positive {
    color: var(--success);
}

.change-negative {
    color: var(--danger);
}

.change-neutral {
    color: var(--neutral);
}

/* Empty State */
.empty-state {
    text-align: center;
    padding: 4rem 2rem;
    background: var(--bg-secondary);
    border-radius: 8px;
}

.empty-icon {
    font-size: 4rem;
    margin-bottom: 1rem;
}

.empty-state h2 {
    margin-bottom: 0.5rem;
}

.empty-state p {
    color: var(--text-muted);
}

/* Bar Chart */
.chart-section {
    margin-bottom: 1.5rem;
}

.chart-section:last-child {
    margin-bottom: 0;
}

.chart-section h4 {
    color: var(--text-muted);
    font-size: 0.7rem;
    margin-bottom: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    font-weight: 600;
}

.bar-chart {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
}

.bar-row {
    display: grid;
    grid-template-columns: minmax(200px, 300px) 1fr 80px;
    align-items: center;
    gap: 0.75rem;
    padding: 0.25rem 0;
}

.bar-row:hover {
    background: rgba(0, 173, 216, 0.05);
}

.bar-name {
    font-size: 0.8rem;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    font-weight: 400;
}

.bar-track {
    height: 22px;
    background: var(--bg-bar-track);
    border-radius: 3px;
    overflow: hidden;
}

.bar {
    height: 100%;
    border-radius: 3px;
    transition: width 0.3s ease;
    min-width: 4px;
}

.bar-fast {
    background: linear-gradient(90deg, var(--gopher-cyan) 0%, var(--gopher-blue) 100%);
}

.bar-medium {
    background: linear-gradient(90deg, #f6ad55 0%, #ed8936 100%);
}

.bar-slow {
    background: linear-gradient(90deg, #fc8181 0%, #f56565 100%);
}

.bar-value {
    font-size: 0.75rem;
    color: var(--text-secondary);
    font-weight: 500;
    text-align: right;
    font-variant-numeric: tabular-nums;
}

.no-data {
    color: var(--text-muted);
    font-size: 0.8rem;
    padding: 0.5rem 0;
}

/* Responsive */
@media (max-width: 768px) {
    .container {
        padding: 1rem;
    }

    header h1 {
        font-size: 1.5rem;
    }

    .metadata,
    .stats {
        grid-template-columns: 1fr;
    }

    .chart-grid {
        grid-template-columns: 1fr;
    }

    .tabs {
        flex-direction: column;
    }

    .tab-btn {
        border-bottom: 1px solid var(--border);
        border-left: 3px solid transparent;
    }

    .tab-btn.active {
        border-bottom-color: var(--border);
        border-left-color: var(--primary);
    }
}
`
}

func (g *Generator) getJS() string {
	return `
document.addEventListener('DOMContentLoaded', function() {
    const tabBtns = document.querySelectorAll('.tab-btn');
    const tabPanes = document.querySelectorAll('.tab-pane');

    tabBtns.forEach(btn => {
        btn.addEventListener('click', function() {
            const tabId = this.getAttribute('data-tab');

            
            tabBtns.forEach(b => b.classList.remove('active'));
            
            this.classList.add('active');

            
            tabPanes.forEach(pane => pane.classList.remove('active'));
            
            document.getElementById(tabId).classList.add('active');
        });
    });

    
    if (typeof window.timeData !== 'undefined') {
        createTimeChart(window.timeData);
    }
    if (typeof window.memData !== 'undefined') {
        createMemChart(window.memData);
    }
});

function createTimeChart(data) {
    const ctx = document.getElementById('timeChart').getContext('2d');
    new Chart(ctx, {
        type: 'bar',
        data: data,
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)'
                    },
                    ticks: {
                        color: '#94a3b8'
                    }
                },
                x: {
                    grid: {
                        display: false
                    },
                    ticks: {
                        color: '#94a3b8',
                        maxRotation: 45,
                        minRotation: 45
                    }
                }
            }
        }
    });
}

function createMemChart(data) {
    const ctx = document.getElementById('memChart').getContext('2d');
    new Chart(ctx, {
        type: 'bar',
        data: data,
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)'
                    },
                    ticks: {
                        color: '#94a3b8'
                    }
                },
                x: {
                    grid: {
                        display: false
                    },
                    ticks: {
                        color: '#94a3b8',
                        maxRotation: 45,
                        minRotation: 45
                    }
                }
            }
        }
    });
}
`
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

func joinSections(sections ...string) string {
	return joinStrings(sections, "\n")
}

func formatNumber(n int64) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	} else if n >= 1000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

func escapeHTML(s string) string {
	s = replaceAll(s, "&", "&amp;")
	s = replaceAll(s, "<", "&lt;")
	s = replaceAll(s, ">", "&gt;")
	s = replaceAll(s, "\"", "&quot;")
	s = replaceAll(s, "'", "&#39;")
	return s
}

func escapeJS(s string) string {
	s = replaceAll(s, "\\", "\\\\")
	s = replaceAll(s, "'", "\\'")
	s = replaceAll(s, "\"", "\\\"")
	s = replaceAll(s, "\n", "\\n")
	s = replaceAll(s, "\r", "\\r")
	s = replaceAll(s, "\t", "\\t")
	return s
}

func replaceAll(s, old, new string) string {
	result := ""
	for {
		idx := indexOf(s, old)
		if idx == -1 {
			result += s
			break
		}
		result += s[:idx] + new
		s = s[idx+len(old):]
	}
	return result
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func getClassForChange(pct, threshold float64) string {
	if pct > threshold {
		return "change-negative"
	} else if pct < -threshold {
		return "change-positive"
	}
	return "change-neutral"
}
