package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	primaryColor   = lipgloss.Color("86")
	successColor   = lipgloss.Color("42")
	warningColor   = lipgloss.Color("226")
	dangerColor    = lipgloss.Color("196")
	secondaryColor = lipgloss.Color("245")
	mutedColor     = lipgloss.Color("241")
	borderColor    = lipgloss.Color("238")

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2)

	cardTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	metricValueStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("255"))

	benchNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Width(40)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Width(12).
			Align(lipgloss.Right)

	improvementStyle = lipgloss.NewStyle().
				Foreground(successColor).
				Bold(true)

	regressionStyle = lipgloss.NewStyle().
			Foreground(dangerColor).
			Bold(true)

	neutralStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	barPositive = lipgloss.Color("40")
	barNegative = lipgloss.Color("196")
	barNeutral  = lipgloss.Color("245")

	footerStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginTop(1)
)

func GetChangeStyle(pct float64, threshold float64) lipgloss.Style {
	if pct > threshold {
		return regressionStyle
	} else if pct < -threshold {
		return improvementStyle
	}
	return neutralStyle
}

func GetBarColor(isNegative bool) lipgloss.Color {
	if isNegative {
		return barNegative
	}
	return barPositive
}
