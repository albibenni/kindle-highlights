package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/albibenni/kindle-highlights/parser"
	"github.com/albibenni/kindle-highlights/types"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joho/godotenv"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	title, desc, raw string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title + " " + i.desc }

type sessionState int

const (
	stateSelectingSource sessionState = iota
	stateCustomPathInput
	stateSelectingBook
)

type model struct {
	state      sessionState
	sourceList list.Model
	bookList   list.Model
	textInput  textinput.Model
	choice     string
	author     string
	rawTitle   string
	err        error
	done       bool
	path       string
	dest       string
	width      int
	height     int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

// --- Update Helpers ---

func (m model) handleKeyInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateSelectingSource:
		return m.updateSelectingSource(msg)
	case stateCustomPathInput:
		return m.updateCustomPathInput(msg)
	case stateSelectingBook:
		return m.updateSelectingBook(msg)
	default:
		return m, nil
	}
}

func (m model) updateSelectingSource(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "enter" {
		if i, ok := m.sourceList.SelectedItem().(item); ok {
			if i.title == "Default Path" {
				m.path = i.desc
				return m.loadBooks()
			}
			if i.title == "Custom Path" {
				m.state = stateCustomPathInput
				m.textInput.Focus()
				return m, nil
			}
		}
	}
	var cmd tea.Cmd
	m.sourceList, cmd = m.sourceList.Update(msg)
	return m, cmd
}

func (m model) updateCustomPathInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "enter" {
		m.path = m.textInput.Value()
		if m.path != "" {
			return m.loadBooks()
		}
	}
	if msg.String() == "esc" {
		m.state = stateSelectingSource
		return m, nil
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) updateSelectingBook(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.bookList.FilterState() != list.Filtering {
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "j":
			m.bookList.CursorDown()
			return m, nil
		case "k":
			m.bookList.CursorUp()
			return m, nil
		}
	}

	if msg.String() == "enter" {
		return m.handleBookSelection()
	}

	var cmd tea.Cmd
	m.bookList, cmd = m.bookList.Update(msg)
	return m, cmd
}

func (m model) handleResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width, m.height = msg.Width, msg.Height
	h, v := docStyle.GetFrameSize()

	m.sourceList.SetSize(msg.Width-h, msg.Height-v)
	if m.state == stateSelectingBook {
		m.bookList.SetSize(msg.Width-h, msg.Height-v)
	}
	return m, nil
}

func (m model) updateActiveComponent(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.state {
	case stateSelectingSource:
		m.sourceList, cmd = m.sourceList.Update(msg)
	case stateCustomPathInput:
		m.textInput, cmd = m.textInput.Update(msg)
	case stateSelectingBook:
		m.bookList, cmd = m.bookList.Update(msg)
	}
	return m, cmd
}

// --- Logic Helpers ---

func (m model) loadBooks() (model, tea.Cmd) {
	note := parser.Note{FileLocation: m.path}
	books, err := note.DiscoverBooks()
	if err != nil {
		m.err = err
		return m, tea.Quit
	}

	items := []list.Item{}
	for _, b := range books {
		items = append(items, item{title: b.Title, desc: b.Author, raw: b.RawLine})
	}

	m.bookList = list.New(items, list.NewDefaultDelegate(), 0, 0)
	m.bookList.Title = "Select a Book"

	if m.width > 0 && m.height > 0 {
		h, v := docStyle.GetFrameSize()
		m.bookList.SetSize(m.width-h, m.height-v)
	}

	m.state = stateSelectingBook
	return m, nil
}

func (m model) handleBookSelection() (model, tea.Cmd) {
	i, ok := m.bookList.SelectedItem().(item)
	if !ok {
		return m, nil
	}

	m.choice, m.author, m.rawTitle = i.title, i.desc, i.raw
	m.done = true

	note := parser.Note{
		Title:             m.rawTitle,
		FileLocation:      m.path,
		IsLookingForTitle: true,
	}

	if _, err := note.ParseNotes(); err != nil {
		m.err = err
		return m, tea.Quit
	}
	dest, err := note.WriteFile()
	if err != nil {
		m.err = err
		return m, tea.Quit
	}
	m.dest = dest
	return m, tea.Quit
}

// --- View Helpers ---

func (m model) View() string {
	if m.err != nil {
		return m.renderError()
	}
	if m.done {
		return m.renderDone()
	}

	switch m.state {
	case stateSelectingSource:
		return m.renderSourceSelection()
	case stateCustomPathInput:
		return m.renderCustomPathInput()
	case stateSelectingBook:
		return m.renderBookSelection()
	default:
		return ""
	}
}

func (m model) renderError() string {
	return fmt.Sprintf("\nError: %v\n", m.err)
}

func (m model) renderDone() string {
	return fmt.Sprintf("\n✓ Successfully exported highlights for '%s'\nDest: %s\n", m.choice, m.dest)
}

func (m model) renderSourceSelection() string {
	return docStyle.Render(m.sourceList.View())
}

func (m model) renderCustomPathInput() string {
	return docStyle.Render(
		fmt.Sprintf(
			"Enter path to your clippings file:\n\n%s\n\n(esc to go back)",
			m.textInput.View(),
		),
	)
}

func (m model) renderBookSelection() string {
	return docStyle.Render(m.bookList.View())
}

// --- Setup ---

func main() {
	loadEnvironment()

	clippingPath := types.ClippingPath.Value()
	sourceItems := []list.Item{
		item{title: "Default Path", desc: clippingPath},
		item{title: "Custom Path", desc: "Manually enter a path to your clippings file"},
	}

	ti := textinput.New()
	ti.Placeholder = "/path/to/My Clippings.txt"
	ti.CharLimit = 255
	ti.Width = 50

	m := model{
		state:      stateSelectingSource,
		sourceList: list.New(sourceItems, list.NewDefaultDelegate(), 0, 0),
		textInput:  ti,
	}
	m.sourceList.Title = "Kindle Highlights - Select Source"
	m.sourceList.SetShowStatusBar(false)
	m.sourceList.SetFilteringEnabled(false)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func loadEnvironment() {
	envFile := types.GetEnvFile()
	if envFile == "wrong pc" {
		log.Fatal("Windows is not supported yet.")
	}

	if err := godotenv.Load(envFile); err != nil {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			configPath := filepath.Join(homeDir, ".config", "kindle-highlights", ".env")
			_ = godotenv.Load(configPath)
		}
	}
}
