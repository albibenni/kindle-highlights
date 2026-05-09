package tui

import (
	"fmt"
)

func (m Model) View() string {
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

func (m Model) renderError() string {
	return fmt.Sprintf("\nError: %v\n", m.Err)
}

func (m Model) renderDone() string {
	return fmt.Sprintf("\n✓ Successfully exported highlights for '%s'\nDest: %s\n", m.Choice, m.Dest)
}

func (m Model) renderSourceSelection() string {
	return DocStyle.Render(m.SourceList.View())
}

func (m Model) renderCustomPathInput() string {
	return DocStyle.Render(
		fmt.Sprintf(
			"Enter path to your clippings file:\n\n%s\n\n(esc to go back)",
			m.TextInput.View(),
		),
	)
}

func (m Model) renderBookSelection() string {
	return DocStyle.Render(m.BookList.View())
}
