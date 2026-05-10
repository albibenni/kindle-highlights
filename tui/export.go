package tui

import (
	"os"

	"github.com/albibenni/kindle-highlights/parser"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) prepareExport() (tea.Model, tea.Cmd) {
	m.State = StateConfirmExport
	return m, nil
}

func (m *Model) updateConfirmExport(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "y":
		return m.runExport()
	case "esc", "n":
		m.State = StateSelectingDest
		return m, nil
	}
	return m, nil
}

func (m *Model) runExport() (tea.Model, tea.Cmd) {
	note := parser.Note{
		Title:             m.RawTitle,
		FileLocation:      m.Path,
		BasePath:          m.BasePath,
		IsLookingForTitle: true,
	}

	if _, err := note.ParseNotes(); err != nil {
		m.Err = err
		return m, tea.Quit
	}
	dest, err := note.WriteFile()
	if err != nil {
		m.Err = err
		return m, tea.Quit
	}
	m.Dest = dest
	m.Done = true
	m.State = StateConfirmSuccess
	return m, nil
}

func (m *Model) getDestItems() []list.Item {
	notePath := os.Getenv("NOTE_PATH")
	return []list.Item{
		Item{TitleStr: "Default Path", DescStr: notePath},
		Item{TitleStr: "Custom Path", DescStr: "Manually enter a path for your notes"},
	}
}
