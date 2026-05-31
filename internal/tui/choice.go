package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

// Choice represents a selectable option in a menu
type Choice struct {
	Label       string
	Description string
	Value       string
}

type choiceModel struct {
	title   string
	choices []Choice
	cursor  int
	done    bool
	aborted bool
}

func newChoiceModel(title string, choices []Choice) choiceModel {
	return choiceModel{
		title:   title,
		choices: choices,
		cursor:  0,
	}
}

func (m choiceModel) Init() tea.Cmd {
	return nil
}

func (m choiceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.aborted = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
			return m, nil

		case "enter":
			m.done = true
			return m, tea.Quit
		}

		// Number keys for quick selection
		if len(msg.String()) == 1 && msg.String()[0] >= '1' && msg.String()[0] <= '9' {
			idx := int(msg.String()[0] - '1')
			if idx < len(m.choices) {
				m.cursor = idx
				m.done = true
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m choiceModel) View() string {
	if m.done {
		return fmt.Sprintf("%s %s\n",
			SuccessStyle.Render("✓"),
			lipgloss.NewStyle().Foreground(White).Render(m.choices[m.cursor].Label))
	}

	if m.aborted {
		return MutedStyle.Render("  Cancelled.") + "\n"
	}

	var b strings.Builder

	b.WriteString(InputPromptStyle.Render("  "+m.title) + "\n\n")

	for i, choice := range m.choices {
		cursor := "  "
		style := lipgloss.NewStyle().Foreground(Muted)
		numStyle := MutedStyle

		if i == m.cursor {
			cursor = InfoStyle.Render("› ")
			style = lipgloss.NewStyle().Foreground(White)
			numStyle = InfoStyle
		}

		line := fmt.Sprintf("%s%s %s",
			cursor,
			numStyle.Render(fmt.Sprintf("%d.", i+1)),
			style.Render(choice.Label))

		if choice.Description != "" {
			line += "  " + MutedStyle.Render(choice.Description)
		}

		b.WriteString(line + "\n")
	}

	b.WriteString("\n" + MutedStyle.Render("  ↑/↓ to move • enter to select • 1-9 for quick pick") + "\n")

	return b.String()
}

// SelectChoice shows an interactive choice menu and returns the selected Choice.
func SelectChoice(title string, choices []Choice) (*Choice, error) {
	if len(choices) == 0 {
		return nil, fmt.Errorf("no choices provided")
	}

	if !isatty.IsTerminal(os.Stdout.Fd()) {
		// Fallback for non-TTY
		fmt.Fprintf(os.Stderr, "%s\n", title)
		for i, c := range choices {
			fmt.Fprintf(os.Stderr, "  %d. %s\n", i+1, c.Label)
		}
		fmt.Fprint(os.Stderr, "Choice: ")
		var input string
		_, _ = fmt.Scanln(&input)
		idx := 0
		_, _ = fmt.Sscanf(input, "%d", &idx)
		if idx >= 1 && idx <= len(choices) {
			return &choices[idx-1], nil
		}
		return nil, fmt.Errorf("invalid choice")
	}

	model := newChoiceModel(title, choices)
	program := tea.NewProgram(model)

	finalModel, err := program.Run()
	if err != nil {
		return nil, err
	}

	if m, ok := finalModel.(choiceModel); ok {
		if m.aborted {
			return nil, fmt.Errorf("selection cancelled")
		}
		return &m.choices[m.cursor], nil
	}

	return nil, fmt.Errorf("unexpected model type")
}
