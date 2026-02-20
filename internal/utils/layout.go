package utils

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func Center(s string, width int) string {
	padding := width - lipgloss.Width(s)
	if padding <= 0 {
		return s
	}
	left := padding / 2
	right := padding - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}
