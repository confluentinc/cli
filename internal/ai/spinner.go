package ai

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	spinner    spinner.Model
	out        string
	isQuitting bool
}

func (m *model) Init() tea.Cmd {
	m.spinner = spinner.New(spinner.WithSpinner(spinner.Ellipsis))
	return m.spinner.Tick
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.out != "" {
		m.isQuitting = true
		return m, tea.Quit
	}

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			m.isQuitting = true
			cmd = tea.Quit
		}
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
	}

	return m, cmd
}

func (m *model) View() string {
	if m.isQuitting {
		return ""
	}

	return "\n" + fmt.Sprintf("Waiting%s", m.spinner.View())
}
