package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

// InputOption configures a text input prompt
type InputOption func(*inputModel)

// WithPlaceholder sets placeholder text
func WithPlaceholder(placeholder string) InputOption {
	return func(m *inputModel) {
		m.input.Placeholder = placeholder
	}
}

// WithDefault sets a default value
func WithDefault(value string) InputOption {
	return func(m *inputModel) {
		m.input.SetValue(value)
	}
}

// WithValidation sets a validation function
func WithValidation(fn func(string) error) InputOption {
	return func(m *inputModel) {
		m.validate = fn
	}
}

// WithRequired marks the input as required
func WithRequired() InputOption {
	return func(m *inputModel) {
		m.required = true
	}
}

// WithEchoNone hides input (for passwords/secrets)
func WithEchoNone() InputOption {
	return func(m *inputModel) {
		m.input.EchoMode = textinput.EchoNone
	}
}

// WithEchoPassword shows asterisks
func WithEchoPassword() InputOption {
	return func(m *inputModel) {
		m.input.EchoMode = textinput.EchoPassword
	}
}

// WithCharLimit sets character limit
func WithCharLimit(limit int) InputOption {
	return func(m *inputModel) {
		m.input.CharLimit = limit
	}
}

type inputModel struct {
	input    textinput.Model
	label    string
	validate func(string) error
	required bool
	err      string
	done     bool
	aborted  bool
}

func newInputModel(label string, opts ...InputOption) inputModel {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50
	ti.PromptStyle = InputPromptStyle
	ti.TextStyle = InputTextStyle
	ti.PlaceholderStyle = InputPlaceholderStyle
	ti.Cursor.Style = InputCursorStyle
	ti.Prompt = "› "

	m := inputModel{
		input: ti,
		label: label,
	}

	for _, opt := range opts {
		opt(&m)
	}

	return m
}

func (m inputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.aborted = true
			return m, tea.Quit

		case "enter":
			value := strings.TrimSpace(m.input.Value())

			if m.required && value == "" {
				m.err = "This field is required"
				return m, nil
			}

			if m.validate != nil {
				if err := m.validate(value); err != nil {
					m.err = err.Error()
					return m, nil
				}
			}

			m.done = true
			m.err = ""
			return m, tea.Quit
		}
	}

	m.err = ""
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m inputModel) View() string {
	if m.done {
		value := m.input.Value()
		if m.input.EchoMode != textinput.EchoNormal {
			value = strings.Repeat("•", len(value))
		}
		return fmt.Sprintf("%s %s\n",
			SuccessStyle.Render("✓"),
			lipgloss.NewStyle().Foreground(White).Render(m.label+": ")+
				InfoStyle.Render(value))
	}

	if m.aborted {
		return MutedStyle.Render("  Cancelled.") + "\n"
	}

	var b strings.Builder
	b.WriteString(InputPromptStyle.Render("  "+m.label) + "\n")
	b.WriteString("  " + m.input.View() + "\n")

	if m.err != "" {
		b.WriteString("  " + ErrorStyle.Render("⚠ "+m.err) + "\n")
	}

	return b.String()
}

// PromptInput shows an interactive text input prompt and returns the entered value.
// Falls back to simple stdin reading if not a TTY.
func PromptInput(label string, opts ...InputOption) (string, error) {
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		// Fallback for non-TTY: simple prompt
		fmt.Fprintf(os.Stderr, "%s: ", label)
		var input string
		_, err := fmt.Scanln(&input)
		return strings.TrimSpace(input), err
	}

	model := newInputModel(label, opts...)
	program := tea.NewProgram(model)

	finalModel, err := program.Run()
	if err != nil {
		return "", err
	}

	if m, ok := finalModel.(inputModel); ok {
		if m.aborted {
			return "", fmt.Errorf("input cancelled")
		}
		return strings.TrimSpace(m.input.Value()), nil
	}

	return "", fmt.Errorf("unexpected model type")
}
