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
		// Only update if this is the result of our most recent search
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

func (m *Model) updatePathSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		var path string
		if len(m.SearchResults) > 0 && m.SearchIndex < len(m.SearchResults) {
			path = m.SearchResults[m.SearchIndex]
		} else {
			path = m.TextInput.Value()
		}

		if path != "" {
			if m.State == StateCustomPathInput {
				m.Path = path
				return m.LoadBooks()
			} else {
				m.BasePath = path
				return m.handleDestSelection()
			}
		}
		return m, nil

	case "esc":
		if m.State == StateCustomPathInput {
			m.State = StateSelectingSource
		} else {
			m.State = StateSelectingDest
		}
		m.Searching = false
		m.SearchResults = nil
		m.TextInput.Blur()
		m.TextInput.Reset()
		return m, nil

	case "up":
		if m.SearchIndex > 0 {
			m.SearchIndex--
		}
		return m, nil
	case "down":
		if m.SearchIndex < len(m.SearchResults)-1 {
			m.SearchIndex++
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.TextInput, cmd = m.TextInput.Update(msg)

	query := m.TextInput.Value()
	
	// Allow searching with 1 char if it's a path starter, otherwise min 3 chars
	shouldSearch := len(query) >= 3 || (len(query) >= 1 && (strings.HasPrefix(query, "/") || strings.HasPrefix(query, "~")))
	
	if shouldSearch {
		m.Searching = true
		m.SearchID++
		isDir := m.State == StateCustomDestInput
		return m, tea.Batch(cmd, m.searchSystem(query, m.SearchID, isDir))
	}

	m.Searching = false
	m.SearchResults = nil
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
				return m.handleDestSelection()
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

func (m *Model) handleDestSelection() (tea.Model, tea.Cmd) {
	m.Done = true

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
	return m, tea.Quit
}

func (m *Model) getDestItems() []list.Item {
	notePath := os.Getenv("NOTE_PATH")
	return []list.Item{
		Item{TitleStr: "Default Path", DescStr: notePath},
		Item{TitleStr: "Custom Path", DescStr: "Manually enter a path for your notes"},
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

// func (m *Model) getSourceItems() []list.Item {
// 	clippingPath := os.Getenv("CLIPPING_PATH")
// 	return []list.Item{
// 		Item{TitleStr: "Default Path", DescStr: clippingPath},
// 		Item{TitleStr: "Custom Path", DescStr: "Manually enter a path to your clippings file"},
// 	}
// }

func (m *Model) searchSystem(query string, id int, isDir bool) tea.Cmd {
	return func() tea.Msg {
		// 1. Setup a context with a strict 2-second timeout
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		home, _ := os.UserHomeDir()

		// Smart Case: Use case-insensitive glob if query is all lowercase
		globFlag := "--glob"
		if query == strings.ToLower(query) {
			globFlag = "--iglob"
		}

		// 2. Search using rg with safety limits:
		cmd := exec.CommandContext(ctx, "rg",
			"--files",
			"--max-depth", "6",
			"-j1",
			globFlag, "*"+query+"*",
			"-g", "!Library",
			"-g", "!.Trash",
			"-g", "!node_modules",
			home)

		output, err := cmd.Output()

		// If timeout, error, or no matches
		if err != nil {
			return SearchResultMsg{ID: id, Results: []string{}}
		}

		lines := strings.Split(string(output), "\n")
		results := []string{}
		isQueryLower := query == strings.ToLower(query)

		if isDir {
			dirMap := make(map[string]bool)
			for _, line := range lines {
				if line == "" {
					continue
				}

				// Check each segment of the path to find mid-path matches
				parts := strings.Split(line, string(os.PathSeparator))
				currentPath := ""
				if strings.HasPrefix(line, string(os.PathSeparator)) {
					currentPath = string(os.PathSeparator)
				}

				for _, part := range parts {
					if part == "" {
						continue
					}
					currentPath = filepath.Join(currentPath, part)

					// Match segment using Smart Case
					match := false
					if isQueryLower {
						match = strings.Contains(strings.ToLower(part), query)
					} else {
						match = strings.Contains(part, query)
					}

					if match {
						if !dirMap[currentPath] {
							dirMap[currentPath] = true
							results = append(results, currentPath)
						}
						// Continue to next file to keep results diverse, 
						// or we could keep checking deeper segments. 
						// For now, we stop at the first matching segment in this path.
						break
					}
				}
				if len(results) >= 10 {
					break
				}
			}
		} else {
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
		}
		return SearchResultMsg{ID: id, Results: results}
	}
}
