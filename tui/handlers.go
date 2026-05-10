package tui

import (
	"github.com/albibenni/kindle-highlights/parser"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) updateSelectingSource(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "enter" {
		if i, ok := m.SourceList.SelectedItem().(Item); ok {
			if i.TitleStr == "Default Path" {
				m.Path = i.DescStr
				return m.LoadBooks()
			}
			if i.TitleStr == "Custom Path" {
				m.State = StateCustomPathInput
				m.TextInput.Reset()
				m.SearchResults = nil
				m.TextInput.Focus()
				return m, nil
			}
		}
	}
	var cmd tea.Cmd
	m.SourceList, cmd = m.SourceList.Update(msg)
	return m, cmd
}

func (m *Model) updateSelectingBook(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.BookList.FilterState() != list.Filtering {
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "j":
			m.BookList.CursorDown()
			return m, nil
		case "k":
			m.BookList.CursorUp()
			return m, nil
		}
	}

	if msg.String() == "enter" {
		return m.handleBookSelection()
	}

	var cmd tea.Cmd
	m.BookList, cmd = m.BookList.Update(msg)
	return m, cmd
}

func (m *Model) updateSelectingDest(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "enter" {
		if i, ok := m.DestList.SelectedItem().(Item); ok {
			if i.TitleStr == "Default Path" {
				m.BasePath = i.DescStr
				return m.prepareExport()
			}
			if i.TitleStr == "Custom Path" {
				m.State = StateCustomDestInput
				m.TextInput.Reset()
				m.SearchResults = nil
				m.TextInput.Focus()
				return m, nil
			}
		}
	}
	if msg.String() == "esc" {
		m.State = StateSelectingBook
		return m, nil
	}

	var cmd tea.Cmd
	m.DestList, cmd = m.DestList.Update(msg)
	return m, cmd
}

func (m *Model) LoadBooks() (tea.Model, tea.Cmd) {
	note := parser.Note{FileLocation: m.Path}
	books, err := note.DiscoverBooks()
	if err != nil {
		m.Err = err
		return m, tea.Quit
	}

	items := []list.Item{}
	for _, b := range books {
		items = append(items, Item{TitleStr: b.Title, DescStr: b.Author, Raw: b.RawLine})
	}

	m.BookList = list.New(items, list.NewDefaultDelegate(), 0, 0)
	m.BookList.Title = "Select a Book"

	var cmd tea.Cmd
	if m.Width > 0 && m.Height > 0 {
		h, v := DocStyle.GetFrameSize()
		m.BookList.SetSize(m.Width-h, m.Height-v)
		m.BookList, cmd = m.BookList.Update(tea.WindowSizeMsg{Width: m.Width - h, Height: m.Height - v})
	}

	m.State = StateSelectingBook
	return m, cmd
}

func (m *Model) handleBookSelection() (tea.Model, tea.Cmd) {
	i, ok := m.BookList.SelectedItem().(Item)
	if !ok {
		return m, nil
	}

	m.Choice, m.Author, m.RawTitle = i.TitleStr, i.DescStr, i.Raw

	// Initialize Dest List
	m.DestList = list.New(m.getDestItems(), list.NewDefaultDelegate(), 0, 0)
	m.DestList.Title = "Select Destination Path"
	m.DestList.SetShowStatusBar(false)
	m.DestList.SetFilteringEnabled(false)

	if m.Width > 0 && m.Height > 0 {
		h, v := DocStyle.GetFrameSize()
		m.DestList.SetSize(m.Width-h, m.Height-v)
	}

	m.State = StateSelectingDest
	return m, nil
}
