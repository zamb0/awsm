package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

type confirmModel struct {
	label     string
	confirmed bool
	done      bool
	aborted   bool
	cursor    int // 0 = yes, 1 = no
	danger    bool
}

func newConfirmModel(label string, danger bool) confirmModel {
	return confirmModel{
		label:  label,
		cursor: 1, // Default to "No"
		danger: danger,
	}
}

func (m confirmModel) Init() tea.Cmd {
	return nil
}

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.aborted = true
			return m, tea.Quit

		case "left", "h":
			m.cursor = 0
			return m, nil

		case "right", "l":
			m.cursor = 1
			return m, nil

		case "tab":
			m.cursor = (m.cursor + 1) % 2
			return m, nil

		case "y", "Y":
			m.confirmed = true
			m.done = true
			return m, tea.Quit

		case "n", "N":
			m.confirmed = false
			m.done = true
			return m, tea.Quit

		case "enter":
			m.confirmed = m.cursor == 0
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m confirmModel) View() string {
	if m.done {
		if m.confirmed {
			return fmt.Sprintf("%s %s %s\n",
				SuccessStyle.Render("✓"),
				lipgloss.NewStyle().Foreground(White).Render(m.label),
				SuccessStyle.Render("Yes"))
		}
		return fmt.Sprintf("%s %s %s\n",
			MutedStyle.Render("✗"),
			lipgloss.NewStyle().Foreground(White).Render(m.label),
			MutedStyle.Render("No"))
	}

	if m.aborted {
		return MutedStyle.Render("  Cancelled.") + "\n"
	}

	var b strings.Builder

	// Question
	promptColor := InputPromptStyle
	if m.danger {
		promptColor = lipgloss.NewStyle().Foreground(Rose).Bold(true)
	}
	b.WriteString(promptColor.Render("  "+m.label) + "\n")

	// Yes/No buttons
	yesStyle := lipgloss.NewStyle().Padding(0, 2)
	noStyle := lipgloss.NewStyle().Padding(0, 2)

	if m.cursor == 0 {
		if m.danger {
			yesStyle = yesStyle.Background(Rose).Foreground(White).Bold(true)
		} else {
			yesStyle = yesStyle.Background(Primary).Foreground(lipgloss.Color("#000000")).Bold(true)
		}
		noStyle = noStyle.Foreground(Muted)
	} else {
		yesStyle = yesStyle.Foreground(Muted)
		noStyle = noStyle.Background(lipgloss.Color("#374151")).Foreground(White).Bold(true)
	}

	b.WriteString("  " + yesStyle.Render("Yes") + "  " + noStyle.Render("No") + "\n")
	b.WriteString(MutedStyle.Render("  ←/→ to select • enter to confirm") + "\n")

	return b.String()
}

// Confirm shows an interactive yes/no confirmation prompt.
// Returns true if the user confirmed.
func Confirm(label string) (bool, error) {
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		fmt.Fprintf(os.Stderr, "%s (y/N): ", label)
		var input string
		_, _ = fmt.Scanln(&input)
		return strings.ToLower(strings.TrimSpace(input)) == "y", nil
	}

	model := newConfirmModel(label, false)
	program := tea.NewProgram(model)

	finalModel, err := program.Run()
	if err != nil {
		return false, err
	}

	if m, ok := finalModel.(confirmModel); ok {
		if m.aborted {
			return false, nil
		}
		return m.confirmed, nil
	}

	return false, fmt.Errorf("unexpected model type")
}

// ConfirmDanger shows a destructive action confirmation with red styling.
func ConfirmDanger(label string) (bool, error) {
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		fmt.Fprintf(os.Stderr, "⚠ %s (y/N): ", label)
		var input string
		_, _ = fmt.Scanln(&input)
		return strings.ToLower(strings.TrimSpace(input)) == "y", nil
	}

	model := newConfirmModel(label, true)
	program := tea.NewProgram(model)

	finalModel, err := program.Run()
	if err != nil {
		return false, err
	}

	if m, ok := finalModel.(confirmModel); ok {
		if m.aborted {
			return false, nil
		}
		return m.confirmed, nil
	}

	return false, fmt.Errorf("unexpected model type")
}
