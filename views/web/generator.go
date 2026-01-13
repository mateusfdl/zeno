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
	var benchmarks []string

	for _, b := range suite.Benchmarks {
		benchmarks = append(benchmarks, g.generateBenchmarkRow(b))
	}

	return fmt.Sprintf(`<div class="card">
    <div class="card-header">
        <h3>%s</h3>
        <div class="suite-info">
            <span class="badge">%s</span>
            <span class="badge">%s/%s</span>
        </div>
    </div>
    <div class="card-body">
        <div class="benchmark-list">
%s
        </div>
    </div>
</div>`, suite.Pkg, suite.Go, suite.Goos, suite.Goarch, joinStrings(benchmarks, "\n"))
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
    --bg-primary: #0f172a;
    --bg-secondary: #1e293b;
    --bg-card: #334155;
    --text-primary: #f1f5f9;
    --text-secondary: #94a3b8;
    --text-muted: #64748b;
    --border: #475569;
    --primary: #3b82f6;
    --success: #22c55e;
    --danger: #ef4444;
    --warning: #f59e0b;
    --neutral: #64748b;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
    background: var(--bg-primary);
    color: var(--text-primary);
    line-height: 1.6;
    min-height: 100vh;
}

.container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 2rem;
}

header {
    text-align: center;
    margin-bottom: 2rem;
    padding-bottom: 2rem;
    border-bottom: 1px solid var(--border);
}

header h1 {
    font-size: 2rem;
    margin-bottom: 0.5rem;
    color: var(--primary);
}

.timestamp {
    color: var(--text-muted);
    font-size: 0.875rem;
}

section {
    margin-bottom: 2rem;
}

h2 {
    font-size: 1.5rem;
    margin-bottom: 1rem;
    color: var(--text-primary);
}

h3 {
    font-size: 1.125rem;
    color: var(--text-secondary);
    margin-bottom: 0.5rem;
}

/* Summary */
.summary {
    background: var(--bg-secondary);
    border-radius: 8px;
    padding: 1.5rem;
}

.metadata {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1rem;
}

.metadata-item {
    display: flex;
    flex-direction: column;
}

.metadata-label {
    color: var(--text-muted);
    font-size: 0.875rem;
}

.metadata-value {
    color: var(--text-primary);
    font-weight: 600;
}

/* Stats Cards */
.stats {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: 1rem;
}

.stat-card {
    background: var(--bg-card);
    border-radius: 8px;
    padding: 1.5rem;
    text-align: center;
    border: 2px solid transparent;
}

.stat-card.stat-danger {
    border-color: var(--danger);
}

.stat-card.stat-success {
    border-color: var(--success);
}

.stat-value {
    font-size: 2rem;
    font-weight: 700;
    margin-bottom: 0.5rem;
}

.stat-label {
    color: var(--text-secondary);
    font-size: 0.875rem;
}

/* Tabs */
.tabs-section {
    background: var(--bg-secondary);
    border-radius: 8px;
    overflow: hidden;
}

.tabs {
    display: flex;
    gap: 0;
    border-bottom: 1px solid var(--border);
}

.tab-btn {
    background: transparent;
    border: none;
    color: var(--text-secondary);
    padding: 1rem 1.5rem;
    cursor: pointer;
    border-bottom: 2px solid transparent;
    transition: all 0.2s;
}

.tab-btn:hover {
    color: var(--text-primary);
    background: var(--bg-card);
}

.tab-btn.active {
    color: var(--primary);
    border-bottom-color: var(--primary);
}

.tab-content {
    padding: 1.5rem;
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
    border-radius: 8px;
    margin-bottom: 1rem;
    overflow: hidden;
}

.card-header {
    padding: 1rem 1.5rem;
    background: var(--bg-secondary);
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.card-header h3 {
    margin: 0;
    color: var(--text-primary);
}

.suite-info {
    display: flex;
    gap: 0.5rem;
}

.badge {
    background: var(--primary);
    color: white;
    padding: 0.25rem 0.75rem;
    border-radius: 4px;
    font-size: 0.75rem;
    font-weight: 600;
}

.card-body {
    padding: 1.5rem;
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
