package tui

import (
	"github.com/albibenni/kindle-highlights/parser"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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

func (m Model) handleKeyInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.State {
	case StateSelectingSource:
		return m.updateSelectingSource(msg)
	case StateCustomPathInput:
		return m.updateCustomPathInput(msg)
	case StateSelectingBook:
		return m.updateSelectingBook(msg)
	default:
		return m, nil
	}
}

func (m Model) updateSelectingSource(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "enter" {
		if i, ok := m.SourceList.SelectedItem().(Item); ok {
			if i.TitleStr == "Default Path" {
				m.Path = i.DescStr
				return m.LoadBooks()
			}
			if i.TitleStr == "Custom Path" {
				m.State = StateCustomPathInput
				m.TextInput.Focus()
				return m, nil
			}
		}
	}
	var cmd tea.Cmd
	m.SourceList, cmd = m.SourceList.Update(msg)
	return m, cmd
}

func (m Model) updateCustomPathInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "enter" {
		m.Path = m.TextInput.Value()
		if m.Path != "" {
			return m.LoadBooks()
		}
	}
	if msg.String() == "esc" {
		m.State = StateSelectingSource
		return m, nil
	}
	var cmd tea.Cmd
	m.TextInput, cmd = m.TextInput.Update(msg)
	return m, cmd
}

func (m Model) updateSelectingBook(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

func (m Model) handleResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.Width, m.Height = msg.Width, msg.Height
	h, v := DocStyle.GetFrameSize()

	m.SourceList.SetSize(msg.Width-h, msg.Height-v)
	if m.State == StateSelectingBook {
		m.BookList.SetSize(msg.Width-h, msg.Height-v)
	}
	return m, nil
}

func (m Model) updateActiveComponent(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.State {
	case StateSelectingSource:
		m.SourceList, cmd = m.SourceList.Update(msg)
	case StateCustomPathInput:
		m.TextInput, cmd = m.TextInput.Update(msg)
	case StateSelectingBook:
		m.BookList, cmd = m.BookList.Update(msg)
	}
	return m, cmd
}

func (m Model) LoadBooks() (Model, tea.Cmd) {
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
		// Trigger an update to ensure internal state like pagination is initialized
		m.BookList, cmd = m.BookList.Update(tea.WindowSizeMsg{Width: m.Width - h, Height: m.Height - v})
	}

	m.State = StateSelectingBook
	return m, cmd
}

func (m Model) handleBookSelection() (Model, tea.Cmd) {
	i, ok := m.BookList.SelectedItem().(Item)
	if !ok {
		return m, nil
	}

	m.Choice, m.Author, m.RawTitle = i.TitleStr, i.DescStr, i.Raw
	m.Done = true

	note := parser.Note{
		Title:             m.RawTitle,
		FileLocation:      m.Path,
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
	return m, tea.Quit
}
