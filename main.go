package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/albibenni/kindle-highlights/parser"
	"github.com/albibenni/kindle-highlights/types"
	"github.com/charmbracelet/bubbles/list"
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

type model struct {
	list     list.Model
	choice   string
	author   string
	rawTitle string
	err      error
	done     bool
	path     string
	dest     string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle global quit keys
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Only handle j/k/q when NOT filtering
		if m.list.FilterState() != list.Filtering {
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "j":
				m.list.CursorDown()
				return m, nil
			case "k":
				m.list.CursorUp()
				return m, nil
			}
		}

		if msg.String() == "enter" {
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = i.title
				m.author = i.desc
				m.rawTitle = i.raw
				m.done = true

				// Execute parsing
				note := parser.Note{
					Title:             m.rawTitle, // Use raw title for parsing match
					FileLocation:      m.path,
					IsLookingForTitle: true,
				}

				_, err := note.ParseNotes()
				if err != nil {
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
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nError: %v\n", m.err)
	}
	if m.done {
		return fmt.Sprintf("\n✓ Successfully exported highlights for '%s'\nDest: %s\n", m.choice, m.dest)
	}
	return docStyle.Render(m.list.View())
}

func main() {
	loadEnvironment()

	clippingPath := types.ClippingPath.Value()
	if clippingPath == "" {
		log.Fatal("CLIPPING_PATH not set in environment")
	}

	note := parser.Note{FileLocation: clippingPath}
	books, err := note.DiscoverBooks()
	if err != nil {
		log.Fatalf("Error discovering books: %v", err)
	}

	items := []list.Item{}
	for _, b := range books {
		items = append(items, item{title: b.Title, desc: b.Author, raw: b.RawLine})
	}

	m := model{
		list: list.New(items, list.NewDefaultDelegate(), 0, 0),
		path: clippingPath,
	}
	m.list.Title = "Kindle Highlights - Select a Book"

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
