package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mateusfdl/zeno/bench"
)

type SortMode int

const (
	SortNone SortMode = iota
	SortByNameAsc
	SortByNameDesc
	SortByValueAsc
	SortByValueDesc
)

type Model struct {
	width        int
	height       int
	runs         []bench.Run
	comparison   []bench.ComparisonResult
	threshold    float64
	quitting     bool
	showHelp     bool
	selectedRuns []int
	currentTab   int
	sortMode     SortMode
}

func NewModel(runs []bench.Run, threshold float64) Model {
	return Model{
		runs:       runs,
		threshold:  threshold,
		width:      80,
		height:     24,
		currentTab: 0,
	}
}

func NewComparisonModel(comparison []bench.ComparisonResult, threshold float64) Model {
	return Model{
		comparison: comparison,
		threshold:  threshold,
		width:      80,
		height:     24,
		currentTab: 1,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Close help modal on any key if it's open
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "?":
			m.showHelp = true
			return m, nil
		case "1", "2", "3":
			tabNum := int(msg.String()[0] - '1')
			maxTab := m.getMaxTab()
			if tabNum <= maxTab {
				m.currentTab = tabNum
			}
			return m, nil
		case "s":
			if m.sortMode == SortByNameAsc {
				m.sortMode = SortByNameDesc
			} else {
				m.sortMode = SortByNameAsc
			}
			return m, nil
		case "S":
			if m.sortMode == SortByValueAsc {
				m.sortMode = SortByValueDesc
			} else {
				m.sortMode = SortByValueAsc
			}
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}
	return m, nil
}

func (m Model) getMaxTab() int {
	if len(m.comparison) > 0 {
		return 2 // Compare mode: Summary, Time Changes, Details
	}
	if len(m.runs) > 0 {
		return 2 // Run mode: Execution Time, Memory Usage, Benchmark Output
	}
	return 0
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	if m.showHelp {
		return m.renderHelpModal()
	}

	var content string

	if len(m.comparison) > 0 {
		// Comparison mode
		content = m.renderComparisonView()
	} else if len(m.runs) > 0 {
		// Run mode with tabs
		switch m.currentTab {
		case 0:
			content = m.renderExecutionTimeView()
		case 1:
			content = m.renderMemoryUsageView()
		case 2:
			content = m.renderBenchmarkOutputView()
		}
	} else {
		content = renderNoData("No benchmark data available")
	}

	return renderContainer(content, m.renderTabs(), m.renderHelp())
}

func (m Model) renderRunView(run bench.Run) string {
	var sections []string

	header := m.renderHeader(run)
	sections = append(sections, header)

	for _, suite := range run.Suites {
		suiteSection := m.renderSuite(suite)
		sections = append(sections, suiteSection)

		if len(suite.Benchmarks) > 0 {
			timeChart := m.renderBenchmarkTimeChart(suite)
			if timeChart != "" {
				sections = append(sections, timeChart)
			}

			memChart := m.renderBenchmarkMemoryChart(suite)
			if memChart != "" {
				sections = append(sections, memChart)
			}
		}
	}

	return strings.Join(sections, "\n\n")
}

func (m Model) renderExecutionTimeView() string {
	if len(m.runs) == 0 {
		return renderNoData("No benchmark data available")
	}

	var sections []string
	run := m.runs[0]

	for _, suite := range run.Suites {
		if len(suite.Benchmarks) > 0 {
			// Add suite header
			header := cardTitleStyle.Render(fmt.Sprintf("%s (%s/%s)", suite.Pkg, suite.Goos, suite.Goarch))
			sections = append(sections, header)

			timeChart := m.renderBenchmarkTimeChart(suite)
			if timeChart != "" {
				sections = append(sections, timeChart)
			}
		}
	}

	if len(sections) == 0 {
		return renderNoData("No execution time data available")
	}

	return strings.Join(sections, "\n\n")
}

func (m Model) renderMemoryUsageView() string {
	if len(m.runs) == 0 {
		return renderNoData("No benchmark data available")
	}

	var sections []string
	run := m.runs[0]

	for _, suite := range run.Suites {
		if len(suite.Benchmarks) > 0 {
			// Add suite header
			header := cardTitleStyle.Render(fmt.Sprintf("%s (%s/%s)", suite.Pkg, suite.Goos, suite.Goarch))
			sections = append(sections, header)

			memChart := m.renderBenchmarkMemoryChart(suite)
			if memChart != "" {
				sections = append(sections, memChart)
			}
		}
	}

	if len(sections) == 0 {
		return renderNoData("No memory usage data available")
	}

	return strings.Join(sections, "\n\n")
}

func (m Model) renderBenchmarkOutputView() string {
	if len(m.runs) == 0 {
		return renderNoData("No benchmark data available")
	}

	var sections []string
	run := m.runs[0]

	// Add run header
	header := m.renderHeader(run)
	sections = append(sections, header)

	for _, suite := range run.Suites {
		suiteSection := m.renderSuite(suite)
		sections = append(sections, suiteSection)
	}

	return strings.Join(sections, "\n\n")
}

func (m Model) renderComparisonView() string {
	switch m.currentTab {
	case 0:
		return m.renderComparisonSummary()
	case 1:
		return m.renderBarChart()
	case 2:
		return m.renderComparisonTable()
	}
	return renderNoData("No comparison data available")
}

func (m Model) renderAllRunsView() string {
	var sections []string

	sections = append(sections, cardTitleStyle.Render("All Benchmark Runs"))

	for i, run := range m.runs {
		runHeader := fmt.Sprintf("%s. %s", numberToString(i+1), renderRunLine(run))
		sections = append(sections, lipgloss.NewStyle().Foreground(secondaryColor).Render(runHeader))
		suites := renderSuiteList(run.Suites)
		sections = append(sections, "  "+suites)
	}

	return strings.Join(sections, "\n")
}

func (m Model) renderHeader(run bench.Run) string {
	lines := []string{
		fmt.Sprintf("Version: %s", renderValue(run.Version)),
		fmt.Sprintf("Date: %s", renderDate(run.Date)),
		fmt.Sprintf("Tags: %s", renderTags(run.Tags)),
	}

	return cardStyle.Width(m.width - 4).Render(
		cardTitleStyle.Render("Run Information") + "\n" +
			strings.Join(lines, "\n"),
	)
}

func (m Model) renderSuite(suite bench.Suite) string {
	var lines []string

	lines = append(lines, cardTitleStyle.Render(
		fmt.Sprintf("%s (%s/%s)", suite.Pkg, suite.Goos, suite.Goarch),
	))

	for _, b := range suite.Benchmarks {
		lines = append(lines, renderBenchmark(b))
	}

	return cardStyle.Width(m.width - 4).Render(strings.Join(lines, "\n"))
}

func (m Model) renderComparisonSummary() string {
	regressions := 0
	improvements := 0

	for _, r := range m.comparison {
		if r.IsRegression(m.threshold) {
			regressions++
		} else if r.NsPerOpPct < -m.threshold {
			improvements++
		}
	}

	lines := []string{
		fmt.Sprintf("Total Benchmarks: %s", metricValueStyle.Render(fmt.Sprintf("%d", len(m.comparison)))),
		fmt.Sprintf("Regressions: %s", regressionStyle.Render(fmt.Sprintf("%d", regressions))),
		fmt.Sprintf("Improvements: %s", improvementStyle.Render(fmt.Sprintf("%d", improvements))),
	}

	return cardStyle.Width(m.width - 4).Render(
		cardTitleStyle.Render("Summary") + "\n" +
			strings.Join(lines, "\n"),
	)
}

func (m Model) renderBarChart() string {
	if len(m.comparison) == 0 {
		return ""
	}

	var bars []BarValue
	for _, r := range m.comparison {
		color := barNeutral
		if r.NsPerOpPct > 0 {
			color = barNegative
		} else if r.NsPerOpPct < 0 {
			color = barPositive
		}

		bars = append(bars, BarValue{
			Label: truncateName(r.Name, 25),
			Value: r.NsPerOpPct,
			Color: color,
		})
	}

	chart := BarChart{
		Width:       m.width - 10,
		Values:      bars,
		ShowPercent: true,
	}

	return cardStyle.Width(m.width - 4).Render(
		cardTitleStyle.Render("Time Changes") + "\n" +
			chart.Render(),
	)
}

func (m Model) renderComparisonTable() string {
	if len(m.comparison) == 0 {
		return ""
	}

	var lines []string

	benchWidth := 35
	oldWidth := 12
	newWidth := 12
	deltaWidth := 10

	header := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(benchWidth).Render("Benchmark"),
		lipgloss.NewStyle().Width(oldWidth).Align(lipgloss.Right).Render("Old"),
		lipgloss.NewStyle().Width(newWidth).Align(lipgloss.Right).Render("New"),
		lipgloss.NewStyle().Width(deltaWidth).Align(lipgloss.Right).Render("Delta"),
	)
	lines = append(lines, lipgloss.NewStyle().
		Foreground(mutedColor).
		Render(strings.Repeat("─", m.width-8)))
	lines = append(lines, header)
	lines = append(lines, lipgloss.NewStyle().
		Foreground(mutedColor).
		Render(strings.Repeat("─", m.width-8)))

	for _, r := range m.comparison {
		changeStyle := GetChangeStyle(r.NsPerOpPct, m.threshold)
		changeStr := fmt.Sprintf("%+.1f%%", r.NsPerOpPct)

		row := lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.NewStyle().Width(benchWidth).Render(truncateName(r.Name, benchWidth)),
			lipgloss.NewStyle().Width(oldWidth).Align(lipgloss.Right).Render(fmt.Sprintf("%.0f", r.OldNsPerOp)),
			lipgloss.NewStyle().Width(newWidth).Align(lipgloss.Right).Render(fmt.Sprintf("%.0f", r.NewNsPerOp)),
			changeStyle.Width(deltaWidth).Align(lipgloss.Right).Render(changeStr),
		)
		lines = append(lines, row)
	}

	return cardStyle.Width(m.width - 4).Render(strings.Join(lines, "\n"))
}

func (m Model) renderBenchmarkTimeChart(suite bench.Suite) string {

	var bars []BarValue
	maxTime := 0.0

	for _, b := range suite.Benchmarks {
		if b.NsPerOp > 0 {
			if b.NsPerOp > maxTime {
				maxTime = b.NsPerOp
			}
		}
	}

	if maxTime == 0 {
		return ""
	}

	for _, b := range suite.Benchmarks {
		if b.NsPerOp > 0 {
			color := successColor
			if b.NsPerOp > maxTime*0.7 {
				color = warningColor
			}
			if b.NsPerOp > maxTime*0.9 {
				color = dangerColor
			}

			bars = append(bars, BarValue{
				Label: truncateName(b.Name, 30),
				Value: b.NsPerOp,
				Color: color,
			})
		}
	}

	if len(bars) == 0 {
		return ""
	}

	m.sortBars(bars)

	chart := BarChart{
		Width:       m.width - 10,
		Values:      bars,
		ShowPercent: false,
	}

	return cardStyle.Width(m.width - 4).Render(
		cardTitleStyle.Render("Execution Time (ns/op)") + "\n" +
			chart.Render(),
	)
}

func (m Model) renderBenchmarkMemoryChart(suite bench.Suite) string {

	var bars []BarValue
	maxMem := 0.0

	for _, b := range suite.Benchmarks {
		if b.Mem != nil && b.Mem.BytesPerOp > 0 {
			if b.Mem.BytesPerOp > maxMem {
				maxMem = b.Mem.BytesPerOp
			}
		}
	}

	if maxMem == 0 {
		return ""
	}

	for _, b := range suite.Benchmarks {
		if b.Mem != nil && b.Mem.BytesPerOp > 0 {
			color := successColor
			if b.Mem.BytesPerOp > maxMem*0.7 {
				color = warningColor
			}
			if b.Mem.BytesPerOp > maxMem*0.9 {
				color = dangerColor
			}

			bars = append(bars, BarValue{
				Label: truncateName(b.Name, 30),
				Value: b.Mem.BytesPerOp,
				Color: color,
			})
		}
	}

	if len(bars) == 0 {
		return ""
	}

	m.sortBars(bars)

	chart := BarChart{
		Width:       m.width - 10,
		Values:      bars,
		ShowPercent: false,
	}

	return cardStyle.Width(m.width - 4).Render(
		cardTitleStyle.Render("Memory Usage (B/op)") + "\n" +
			chart.Render(),
	)
}

func (m Model) sortBars(bars []BarValue) {
	switch m.sortMode {
	case SortByNameAsc:
		sort.Slice(bars, func(i, j int) bool {
			return bars[i].Label < bars[j].Label
		})
	case SortByNameDesc:
		sort.Slice(bars, func(i, j int) bool {
			return bars[i].Label > bars[j].Label
		})
	case SortByValueAsc:
		sort.Slice(bars, func(i, j int) bool {
			return bars[i].Value < bars[j].Value
		})
	case SortByValueDesc:
		sort.Slice(bars, func(i, j int) bool {
			return bars[i].Value > bars[j].Value
		})
	}
}

func (m Model) renderTabs() string {
	var tabs []string

	if len(m.comparison) > 0 {
		tabs = []string{"Summary", "Time Changes", "Details"}
	} else if len(m.runs) > 0 {
		tabs = []string{"Execution Time", "Memory Usage", "Benchmark Output"}
	} else {
		return ""
	}

	var parts []string
	for i, tab := range tabs {
		style := lipgloss.NewStyle().
			Foreground(mutedColor).
			Padding(0, 2)

		if i == m.currentTab {
			style = style.Foreground(primaryColor).Bold(true)
		}

		parts = append(parts, style.Render(fmt.Sprintf("%d %s", i+1, tab)))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func (m Model) renderHelp() string {
	help := "?: help | q: quit | 1/2/3: tabs"
	return footerStyle.Render(help)
}

func (m Model) renderHelpModal() string {
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(1, 3).
		Width(50)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		MarginBottom(1)

	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("255")).
		Width(12)

	descStyle := lipgloss.NewStyle().
		Foreground(secondaryColor)

	keybinds := []struct {
		key  string
		desc string
	}{
		{"?", "Toggle this help"},
		{"q", "Quit"},
		{"1/2/3", "Switch tabs"},
		{"s", "Sort by name"},
		{"S", "Sort by value"},
	}

	var lines []string
	lines = append(lines, titleStyle.Render("Keyboard Shortcuts"))
	lines = append(lines, "")

	for _, kb := range keybinds {
		line := lipgloss.JoinHorizontal(
			lipgloss.Top,
			keyStyle.Render(kb.key),
			descStyle.Render(kb.desc),
		)
		lines = append(lines, line)
	}

	lines = append(lines, "")
	lines = append(lines, footerStyle.Render("Press any key to close"))

	modal := modalStyle.Render(strings.Join(lines, "\n"))

	// Center the modal
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		modal,
	)
}

func renderContainer(content, tabs, footer string) string {
	return "\n" + tabs + "\n\n" + content + "\n\n" + footer + "\n"
}

func renderNoData(message string) string {
	return cardStyle.Render(
		lipgloss.NewStyle().
			Foreground(mutedColor).
			Align(lipgloss.Center).
			Render(message),
	)
}

func renderBenchmark(b bench.Benchmark) string {
	var parts []string

	parts = append(parts, benchNameStyle.Render(b.Name))

	if b.NsPerOp > 0 {
		parts = append(parts, valueStyle.Render(fmt.Sprintf("%.2f ns/op", b.NsPerOp)))
	}

	if b.Mem != nil {
		if b.Mem.BytesPerOp > 0 {
			parts = append(parts, valueStyle.Render(fmt.Sprintf("%.0f B/op", b.Mem.BytesPerOp)))
		}
		if b.Mem.AllocsPerOp > 0 {
			parts = append(parts, valueStyle.Render(fmt.Sprintf("%.0f allocs/op", b.Mem.AllocsPerOp)))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func renderSuiteList(suites []bench.Suite) string {
	var parts []string
	for _, suite := range suites {
		parts = append(parts, fmt.Sprintf("%s (%d benches)", suite.Pkg, len(suite.Benchmarks)))
	}
	return strings.Join(parts, ", ")
}

func renderRunLine(run bench.Run) string {
	parts := []string{}
	if run.Version != "" {
		parts = append(parts, run.Version)
	}
	if run.Date > 0 {
		parts = append(parts, renderDate(run.Date))
	}
	if len(run.Tags) > 0 {
		parts = append(parts, renderTags(run.Tags))
	}
	return strings.Join(parts, " · ")
}

func renderValue(v string) string {
	if v == "" {
		return "—"
	}
	return v
}

func renderTags(tags []string) string {
	if len(tags) == 0 {
		return "—"
	}
	return strings.Join(tags, ", ")
}

func renderDate(ts int64) string {
	if ts == 0 {
		return "—"
	}

	return fmt.Sprintf("@%d", ts)
}

func numberToString(n int) string {
	return fmt.Sprintf("%d", n)
}
