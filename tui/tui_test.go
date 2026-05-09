package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func TestTUIStateTransitions(t *testing.T) {
	// Initialize a basic model
	items := []list.Item{
		Item{TitleStr: "Default Path", DescStr: "/test/path"},
		Item{TitleStr: "Custom Path", DescStr: "manual"},
	}
	ti := textinput.New()
	
	m := Model{
		State:      StateSelectingSource,
		SourceList: list.New(items, list.NewDefaultDelegate(), 0, 0),
		TextInput:  ti,
	}

	// 1. Test moving from Source Selection to Custom Path Input
	// Select "Custom Path" (it's the second item, index 1)
	m.SourceList.Select(1)
	
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updatedModel := newModel.(Model)

	if updatedModel.State != StateCustomPathInput {
		t.Errorf("Expected state to be StateCustomPathInput, got %v", updatedModel.State)
	}
	if !updatedModel.TextInput.Focused() {
		t.Error("Expected text input to be focused")
	}

	// 2. Test escaping back to Source Selection
	newModel, _ = updatedModel.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updatedModel = newModel.(Model)

	if updatedModel.State != StateSelectingSource {
		t.Errorf("Expected state to return to StateSelectingSource, got %v", updatedModel.State)
	}

	// 3. Test Window Resize
	newModel, _ = updatedModel.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	updatedModel = newModel.(Model)

	if updatedModel.Width != 100 || updatedModel.Height != 50 {
		t.Errorf("Expected size 100x50, got %dx%d", updatedModel.Width, updatedModel.Height)
	}
}

func TestTUIView(t *testing.T) {
	m := Model{
		State: StateSelectingSource,
		SourceList: list.New([]list.Item{Item{TitleStr: "Test"}}, list.NewDefaultDelegate(), 0, 0),
	}

	view := m.View()
	if view == "" {
		t.Error("View returned empty string")
	}

	m.Err = fmt.Errorf("test error")
	view = m.View()
	if !strings.Contains(view, "Error: test error") {
		t.Errorf("Expected view to contain error message, got: %s", view)
	}
}
