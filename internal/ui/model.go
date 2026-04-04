package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jerryagbesi/skipper/internal/sshconfig"
)

type item struct {
	host sshconfig.Host
}

func (i item) Title() string       { return i.host.Alias }
func (i item) Description() string { return i.host.Hostname }
func (i item) FilterValue() string { return i.host.Alias + " " + i.host.Hostname }

// func hostDescription(host sshconfig.Host) string {
// 	desc := host.Hostname

// 	if host.User != "" {
// 		desc = host.User + "@" + host.Hostname
// 	}

// 	if host.Port != 0 {
// 		desc += fmt.Sprintf(":%d", host.Port)
// 	}

// 	return desc
// }

type Model struct {
	list         list.Model
	selectedHost *sshconfig.Host
	quitting     bool
}

type Result struct {
	Host      *sshconfig.Host
	Cancelled bool
}

func NewModel(hosts []sshconfig.Host) *Model {
	items := make([]list.Item, len(hosts))
	for i, h := range hosts {
		items[i] = item{host: h}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select a Host"
	l.SetFilteringEnabled(true)
	l.SetShowStatusBar(true)
	return &Model{list: l}
}

func (m Model) Init() tea.Cmd { // Do nothing on start up. The list of host would have been loaded by the time the UI starts.
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if m.list.FilterState() == list.Filtering {
				break
			}

			if i, ok := m.list.SelectedItem().(item); ok {
				m.selectedHost = &i.host
			}
			m.quitting = true
			return m, tea.Quit
		}

	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg) // recursively call the update loop
	return m, cmd
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}
	return m.list.View()
}

func Run(hosts []sshconfig.Host) (Result, error) {
	model := NewModel(hosts)

	if len(hosts) == 1 {
		return Result{Host: &hosts[0]}, nil
	}

	program := tea.NewProgram(model, tea.WithAltScreen())

	final, err := program.Run()
	if err != nil {
		return Result{}, err
	}

	m := final.(Model)

	if m.selectedHost == nil {
		return Result{Cancelled: true}, nil
	}

	return Result{Host: m.selectedHost}, nil
}
