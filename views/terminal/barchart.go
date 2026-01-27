package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type BarChart struct {
	Width       int
	Label       string
	Values      []BarValue
	ShowPercent bool
}

type BarValue struct {
	Label string
	Value float64
	Color lipgloss.Color
}

func (bc *BarChart) Render() string {
	if len(bc.Values) == 0 {
		return ""
	}

	var sb strings.Builder

	maxValue := 0.0
	for _, v := range bc.Values {
		absValue := v.Value
		if absValue < 0 {
			absValue = -absValue
		}
		if absValue > maxValue {
			maxValue = absValue
		}
	}
	if maxValue == 0 {
		maxValue = 1
	}

	labelWidth := 30
	valueWidth := 12
	barMaxWidth := bc.Width - labelWidth - valueWidth - 4
	if barMaxWidth < 10 {
		barMaxWidth = 10
	}

	for _, v := range bc.Values {

		absValue := v.Value
		if absValue < 0 {
			absValue = -absValue
		}
		barWidth := int((absValue / maxValue) * float64(barMaxWidth))
		if barWidth > barMaxWidth {
			barWidth = barMaxWidth
		}
		if barWidth < 1 && v.Value != 0 {
			barWidth = 1
		}

		bar := strings.Repeat("█", barWidth)

		valueStr := formatValueWithMode(v.Value, bc.ShowPercent)

		row := lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.NewStyle().Width(labelWidth).Render(v.Label),
			lipgloss.NewStyle().Foreground(v.Color).Render(bar),
			lipgloss.NewStyle().Width(valueWidth).Align(lipgloss.Right).Render(valueStr),
		)

		sb.WriteString(row + "\n\n")
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

func formatValueWithMode(v float64, showPercent bool) string {
	if showPercent {
		if v > 0 {
			return "+" + formatFloat(v) + "%"
		}
		return formatFloat(v) + "%"
	}

	return formatRawValue(v)
}

func formatRawValue(v float64) string {
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

func formatFloat(v float64) string {
	if math.Abs(v) >= 100 {
		return fmtString("%.0f", v)
	} else if math.Abs(v) >= 10 {
		return fmtString("%.1f", v)
	}
	return fmtString("%.2f", v)
}

func fmtString(format string, v float64) string {
	return fmt.Sprintf(format, v)
}

func truncateName(name string, maxLen int) string {
	if len(name) <= maxLen {
		return name
	}
	return name[:maxLen-3] + "..."
}
