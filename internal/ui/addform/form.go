// Package addform renders the interactive prompt used by `skipper add`
// when invoked without positional arguments.
package addform

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Input struct {
	Alias    string
	User     string
	Hostname string
	Port     string
}

type Result struct {
	Alias     string
	User      string
	Hostname  string
	Port      int
	Cancelled bool
}

const (
	fieldAlias = iota
	fieldUser
	fieldHostname
	fieldPort
	fieldCount
)

var fieldLabels = [fieldCount]string{"Alias", "User", "HostName", "Port"}
var fieldPlaceholders = [fieldCount]string{
	"devone (optional, defaults to host name)",
	"root",
	"10.0.0.8",
	"22 (leave blank to omit)",
}

type model struct {
	inputs    [fieldCount]textinput.Model
	focused   int
	err       string
	submitted bool
	cancelled bool
}

func newModel(in Input) model {
	var inputs [fieldCount]textinput.Model
	initVals := [fieldCount]string{in.Alias, in.User, in.Hostname, in.Port}

	for i := range inputs {
		t := textinput.New()
		t.Placeholder = fieldPlaceholders[i]
		t.SetValue(initVals[i])
		inputs[i] = t
	}
	inputs[fieldAlias].Focus()
	return model{inputs: inputs}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit

		case "enter":
			if m.focused < fieldCount-1 {
				m.inputs[m.focused].Blur()
				m.focused++
				m.inputs[m.focused].Focus()
				m.err = ""
				return m, textinput.Blink
			}
			// last field — validate all
			if err := m.validate(); err != "" {
				m.err = err
				return m, nil
			}
			m.submitted = true
			return m, tea.Quit

		case "shift+tab", "up":
			if m.focused > 0 {
				m.inputs[m.focused].Blur()
				m.focused--
				m.inputs[m.focused].Focus()
				m.err = ""
			}
			return m, textinput.Blink

		case "tab", "down":
			if m.focused < fieldCount-1 {
				m.inputs[m.focused].Blur()
				m.focused++
				m.inputs[m.focused].Focus()
				m.err = ""
			}
			return m, textinput.Blink
		}
	}

	var cmd tea.Cmd
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	return m, cmd
}

func (m model) validate() string {
	if err := validateRequiredField("user", m.inputs[fieldUser].Value()); err != nil {
		return err.Error()
	}
	if err := validateRequiredField("host name", m.inputs[fieldHostname].Value()); err != nil {
		return err.Error()
	}
	if err := validateOptionalField("alias", m.inputs[fieldAlias].Value()); err != nil {
		return err.Error()
	}
	if err := validatePort(m.inputs[fieldPort].Value()); err != nil {
		return err.Error()
	}
	return ""
}

func (m model) View() string {
	var b strings.Builder
	b.WriteString("\n  Add SSH host\n\n")
	for i, inp := range m.inputs {
		cursor := "  "
		if i == m.focused {
			cursor = "> "
		}
		required := ""
		if i == fieldUser || i == fieldHostname {
			required = " *"
		}
		fmt.Fprintf(&b, "%s%s%s\n    %s\n\n", cursor, fieldLabels[i], required, inp.View())
	}
	if m.err != "" {
		fmt.Fprintf(&b, "  error: %s\n\n", m.err)
	}
	b.WriteString("  enter: next/submit  tab/shift+tab: navigate  esc: cancel\n")
	return b.String()
}

// Run displays the interactive add-host form and returns the collected values.
// If the user aborts (Ctrl+C / Esc) Result.Cancelled is true and err is nil.
func Run(in Input) (Result, error) {
	m := newModel(in)
	program := tea.NewProgram(m)
	final, err := program.Run()
	if err != nil {
		return Result{}, err
	}

	fm := final.(model)
	if fm.cancelled {
		return Result{Cancelled: true}, nil
	}
	if !fm.submitted {
		return Result{Cancelled: true}, nil
	}

	parsedPort, err := parsePort(fm.inputs[fieldPort].Value())
	if err != nil {
		return Result{}, err
	}

	return Result{
		Alias:    strings.TrimSpace(fm.inputs[fieldAlias].Value()),
		User:     strings.TrimSpace(fm.inputs[fieldUser].Value()),
		Hostname: strings.TrimSpace(fm.inputs[fieldHostname].Value()),
		Port:     parsedPort,
	}, nil
}

func validateAlias(value string) error    { return validateOptionalField("alias", value) }
func validateUser(value string) error     { return validateRequiredField("user", value) }
func validateHostname(value string) error { return validateRequiredField("host name", value) }

func validatePort(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	port, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("port must be a number")
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	return nil
}

func parsePort(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	return strconv.Atoi(value)
}

func validateRequiredField(name, value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fmt.Errorf("%s is required", name)
	}
	if strings.ContainsFunc(trimmed, unicode.IsSpace) {
		return fmt.Errorf("%s cannot contain whitespace", name)
	}
	return nil
}

func validateOptionalField(name, value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	if strings.ContainsFunc(trimmed, unicode.IsSpace) {
		return fmt.Errorf("%s cannot contain whitespace", name)
	}
	return nil
}

// keep these exported for tests that call them directly via the old huh-era names
var _ = validateAlias
var _ = validateUser
var _ = validateHostname
