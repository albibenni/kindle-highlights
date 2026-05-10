package tui

import (
	"fmt"
	"path/filepath"
	"github.com/charmbracelet/lipgloss"
)

var (
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true)
	normalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
)

func (m *Model) View() string {
	if m.Err != nil {
		return m.renderError()
	}
	if m.Done {
		return m.renderDone()
	}

	switch m.State {
	case StateSelectingSource:
		return m.renderSourceSelection()
	case StateCustomPathInput:
		return m.renderCustomPathInput()
	case StateSelectingBook:
		return m.renderBookSelection()
	case StateSelectingDest:
		return m.renderDestSelection()
	case StateCustomDestInput:
		return m.renderCustomDestInput()
	default:
		return ""
	}
}

func (m *Model) renderError() string {
	return fmt.Sprintf("\nError: %v\n", m.Err)
}

func (m *Model) renderDone() string {
	return fmt.Sprintf("\n✓ Successfully exported highlights for '%s'\nDest: %s\n", m.Choice, m.Dest)
}

func (m *Model) renderSourceSelection() string {
	return DocStyle.Render(m.SourceList.View())
}

func (m *Model) renderCustomPathInput() string {
	status := ""
	if m.Searching {
		status = " (Searching...)"
	}

	header := fmt.Sprintf("Enter path to your clippings file%s:", status)
	input := m.TextInput.View()
	footer := "(esc to go back)"
	
	results := ""
	if len(m.SearchResults) > 0 {
		results = "\n\nSearch Results:\n"
		for i, res := range m.SearchResults {
			cursor := "  "
			style := normalStyle
			if i == m.SearchIndex {
				cursor = "> "
				style = selectedStyle
			}
			
			filename := filepath.Base(res)
			results += fmt.Sprintf("%s%s (%s)\n", cursor, style.Render(filename), res)
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"\n"+input,
		"\n"+footer,
		results,
	)

	return DocStyle.Render(content)
}

func (m *Model) renderBookSelection() string {
	return DocStyle.Render(m.BookList.View())
}

func (m *Model) renderDestSelection() string {
	return DocStyle.Render(m.DestList.View())
}

func (m *Model) renderCustomDestInput() string {
	status := ""
	if m.Searching {
		status = " (Searching...)"
	}

	header := fmt.Sprintf("Enter destination path for your notes%s:", status)
	input := m.TextInput.View()
	footer := "(esc to go back)"
	
	results := ""
	if len(m.SearchResults) > 0 {
		results = "\n\nSearch Results:\n"
		for i, res := range m.SearchResults {
			cursor := "  "
			style := normalStyle
			if i == m.SearchIndex {
				cursor = "> "
				style = selectedStyle
			}
			
			filename := filepath.Base(res)
			results += fmt.Sprintf("%s%s (%s)\n", cursor, style.Render(filename), res)
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"\n"+input,
		"\n"+footer,
		results,
	)

	return DocStyle.Render(content)
}
