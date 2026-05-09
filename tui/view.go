package tui

import (
	"fmt"
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
		status = "\n  Searching..."
	}

	results := ""
	if len(m.SourceList.Items()) > 0 && m.TextInput.Value() != "" {
		results = "\n\nSearch Results (arrows to navigate):\n" + m.SourceList.View()
	}

	return DocStyle.Render(
		fmt.Sprintf(
			"Enter path to your clippings file:%s\n\n%s\n\n(esc to go back)%s",
			status,
			m.TextInput.View(),
			results,
		),
	)
}

func (m *Model) renderBookSelection() string {
	return DocStyle.Render(m.BookList.View())
}
