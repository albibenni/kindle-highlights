package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SearchResultMsg:
		if msg.ID == m.SearchID {
			m.Searching = false
			m.SearchResults = msg.Results
			m.SearchIndex = 0
		}
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return m.handleKeyInput(msg)

	case tea.WindowSizeMsg:
		return m.handleResize(msg)
	}

	return m.updateActiveComponent(msg)
}

func (m *Model) handleKeyInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.State {
	case StateSelectingSource:
		return m.updateSelectingSource(msg)
	case StateCustomPathInput, StateCustomDestInput:
		return m.updatePathSearch(msg)
	case StateSelectingBook:
		return m.updateSelectingBook(msg)
	case StateSelectingDest:
		return m.updateSelectingDest(msg)
	case StateConfirmExport:
		return m.updateConfirmExport(msg)
	case StateConfirmSuccess:
		return m, tea.Quit
	default:
		return m, nil
	}
}

func (m *Model) handleResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.Width, m.Height = msg.Width, msg.Height
	h, v := DocStyle.GetFrameSize()

	m.SourceList.SetSize(msg.Width-h, msg.Height-v)
	if m.State == StateSelectingBook {
		m.BookList.SetSize(msg.Width-h, msg.Height-v)
	}
	if m.State == StateSelectingDest {
		m.DestList.SetSize(msg.Width-h, msg.Height-v)
	}
	return m, nil
}

func (m *Model) updateActiveComponent(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.State {
	case StateSelectingSource:
		m.SourceList, cmd = m.SourceList.Update(msg)
	case StateCustomPathInput, StateCustomDestInput:
		m.TextInput, cmd = m.TextInput.Update(msg)
	case StateSelectingBook:
		m.BookList, cmd = m.BookList.Update(msg)
	case StateSelectingDest:
		m.DestList, cmd = m.DestList.Update(msg)
	}
	return m, cmd
}
