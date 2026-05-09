package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var DocStyle = lipgloss.NewStyle().Margin(1, 2)

type Item struct {
	TitleStr string
	DescStr  string
	Raw      string
}

func (i Item) Title() string       { return i.TitleStr }
func (i Item) Description() string { return i.DescStr }
func (i Item) FilterValue() string { return i.TitleStr + " " + i.DescStr }

type SessionState int

const (
	StateSelectingSource SessionState = iota
	StateCustomPathInput
	StateSelectingBook
)

type SearchResultMsg struct {
	ID      int
	Results []string
}

type Model struct {
	State      SessionState
	SourceList list.Model
	BookList   list.Model
	TextInput  textinput.Model
	Choice     string
	Author     string
	RawTitle   string
	Err        error
	Done       bool
	Path       string
	Dest       string
	Width      int
	Height     int
	Searching  bool
	SearchID   int
}

func (m *Model) Init() tea.Cmd {
	return nil
}
