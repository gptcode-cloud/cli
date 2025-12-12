package output

import (
	"fmt"
	"strings"

	"os"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var (
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7aa2f7")). // Tokyo Night Blue
			MarginTop(1).
			MarginBottom(1)

	SeparatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3b4261")) // Tokyo Night Gutter (Dark Grey)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9ece6a")). // Tokyo Night Green
			Bold(true)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7dcfff")). // Tokyo Night Cyan
			Padding(0, 1)
)

func GetTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return 80
	}
	return width
}

func Separator() string {
	width := GetTerminalWidth()
	return SeparatorStyle.Render(strings.Repeat("━", width))
}

func Header(text string) string {
	return HeaderStyle.Render(text)
}

func Success(text string) string {
	return SuccessStyle.Render(fmt.Sprintf("✓ %s", text))
}

func CodeBlockBox(title, code string) string {
	styledCode := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#c0caf5")). // Tokyo Night FG
		Render(code)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#e0af68")) // Tokyo Night Yellow

	content := fmt.Sprintf("%s\n\n%s", titleStyle.Render(title), styledCode)
	return BoxStyle.Render(content)
}
