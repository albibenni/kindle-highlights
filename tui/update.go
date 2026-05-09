package tui

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/albibenni/kindle-highlights/parser"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SearchResultMsg:
		m.Searching = false
		items := []list.Item{}
		for _, path := range msg {
			items = append(items, Item{TitleStr: filepath.Base(path), DescStr: path, Raw: path})
		}
		m.SourceList.SetItems(items)
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
	case StateCustomPathInput:
		return m.updateCustomPathInput(msg)
	case StateSelectingBook:
		return m.updateSelectingBook(msg)
	default:
		return m, nil
	}
}

func (m *Model) updateSelectingSource(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "enter" {
		if i, ok := m.SourceList.SelectedItem().(Item); ok {
			if i.TitleStr == "Default Path" {
				m.Path = i.DescStr
				return m.LoadBooks()
			}
			if i.TitleStr == "Custom Path" {
				m.State = StateCustomPathInput
				m.TextInput.Focus()
				// Clear the search list so it doesn't show previous selections
				m.SourceList.SetItems([]list.Item{})
				return m, nil
			}
		}
	}
	var cmd tea.Cmd
	m.SourceList, cmd = m.SourceList.Update(msg)
	return m, cmd
}

func (m *Model) updateCustomPathInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if i, ok := m.SourceList.SelectedItem().(Item); ok && len(m.SourceList.Items()) > 0 && m.TextInput.Value() != "" {
			m.Path = i.Raw
			return m.LoadBooks()
		}
		m.Path = m.TextInput.Value()
		if m.Path != "" {
			return m.LoadBooks()
		}
		return m, nil

	case "esc":
		m.State = StateSelectingSource
		m.Searching = false
		m.TextInput.Blur()
		m.TextInput.Reset()
		m.SourceList.SetItems(m.getSourceItems())
		return m, nil

	case "up", "down":
		var listCmd tea.Cmd
		m.SourceList, listCmd = m.SourceList.Update(msg)
		return m, listCmd
	}

	var cmd tea.Cmd
	m.TextInput, cmd = m.TextInput.Update(msg)

	query := m.TextInput.Value()
	if len(query) >= 3 {
		m.Searching = true
		return m, tea.Batch(cmd, m.searchSystem(query))
	}

	// If query is too short, reset results and stop searching indicator
	m.Searching = false
	m.SourceList.SetItems([]list.Item{})
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

func (m *Model) handleResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.Width, m.Height = msg.Width, msg.Height
	h, v := DocStyle.GetFrameSize()

	m.SourceList.SetSize(msg.Width-h, msg.Height-v)
	if m.State == StateSelectingBook {
		m.BookList.SetSize(msg.Width-h, msg.Height-v)
	}
	return m, nil
}

func (m *Model) updateActiveComponent(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m *Model) getSourceItems() []list.Item {
	clippingPath := os.Getenv("CLIPPING_PATH")
	return []list.Item{
		Item{TitleStr: "Default Path", DescStr: clippingPath},
		Item{TitleStr: "Custom Path", DescStr: "Manually enter a path to your clippings file"},
	}
}

func (m *Model) searchSystem(query string) tea.Cmd {
	return func() tea.Msg {
		// 1. Setup a context with a strict 2-second timeout
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		home, _ := os.UserHomeDir()

		// 2. Search using rg with safety limits:
		// - max-depth 6 (deep enough for most user files)
		// - exclude large system dirs
		// - j1 (single thread)
		// - glob for the query
		cmd := exec.CommandContext(ctx, "rg",
			"--files",
			"--max-depth", "6",
			"-j1",
			"--glob", "*"+query+"*",
			"-g", "!Library",
			"-g", "!.Trash",
			"-g", "!node_modules",
			home)

		output, err := cmd.Output()

		// If timeout, error, or no matches
		if err != nil {
			return SearchResultMsg{}
		}

		lines := strings.Split(string(output), "\n")
		results := []string{}
		count := 0
		for _, line := range lines {
			if line != "" {
				results = append(results, line)
				count++
				if count >= 10 {
					break
				}
			}
		}
		return SearchResultMsg(results)
	}
}
