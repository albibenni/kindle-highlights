package tui

import (
	"fmt"
	"os"
	"path/filepath"
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

	m := &Model{
		State:      StateSelectingSource,
		SourceList: list.New(items, list.NewDefaultDelegate(), 0, 0),
		TextInput:  ti,
	}

	// 1. Test moving from Source Selection to Custom Path Input
	// Select "Custom Path" (it's the second item, index 1)
	m.SourceList.Select(1)

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updatedModel := newModel.(*Model)

	if updatedModel.State != StateCustomPathInput {
		t.Errorf("Expected state to be StateCustomPathInput, got %v", updatedModel.State)
	}
	if !updatedModel.TextInput.Focused() {
		t.Error("Expected text input to be focused")
	}

	// 2. Test escaping back to Source Selection
	newModel, _ = updatedModel.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updatedModel = newModel.(*Model)

	if updatedModel.State != StateSelectingSource {
		t.Errorf("Expected state to return to StateSelectingSource, got %v", updatedModel.State)
	}

	// 3. Test Window Resize
	newModel, _ = updatedModel.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	updatedModel = newModel.(*Model)

	if updatedModel.Width != 100 || updatedModel.Height != 50 {
		t.Errorf("Expected size 100x50, got %dx%d", updatedModel.Width, updatedModel.Height)
	}
}

func TestTUILoadBooks(t *testing.T) {
	// Create a temporary clippings file
	content := "Book 1 (Author 1)\n- Metadata\n\nContent\n==========\n"
	tmpFile, err := os.CreateTemp("", "clippings_test.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(content)
	tmpFile.Close()

	m := &Model{
		Path: tmpFile.Name(),
	}

	newModel, cmd := m.LoadBooks()
	updatedModel := newModel.(*Model)
	if updatedModel.Err != nil {
		t.Fatalf("LoadBooks() error = %v", updatedModel.Err)
	}
	if cmd != nil {
		t.Error("Expected cmd to be nil")
	}
	if len(updatedModel.BookList.Items()) != 1 {
		t.Errorf("Expected 1 book, got %d", len(updatedModel.BookList.Items()))
	}
	if updatedModel.State != StateSelectingBook {
		t.Errorf("Expected state StateSelectingBook, got %v", updatedModel.State)
	}
}

func TestTUIView(t *testing.T) {
	m := &Model{
		State:      StateSelectingSource,
		SourceList: list.New([]list.Item{Item{TitleStr: "Test"}}, list.NewDefaultDelegate(), 0, 0),
		TextInput:  textinput.New(),
	}

	// Test Source Selection View
	view := m.View()
	if view == "" {
		t.Error("View returned empty string for Source Selection")
	}

	// Test Custom Path Input View
	m.State = StateCustomPathInput
	view = m.View()
	if !strings.Contains(view, "Enter path") {
		t.Error("Expected view to contain 'Enter path'")
	}

	// Test Error View
	m.Err = fmt.Errorf("test error")
	view = m.View()
	if !strings.Contains(view, "Error: test error") {
		t.Errorf("Expected view to contain error message, got: %s", view)
	}

	// Test Done View
	m.Err = nil
	m.Done = true
	m.Choice = "My Book"
	m.Dest = "/path/to/dest"
	view = m.View()
	if !strings.Contains(view, "Successfully exported") || !strings.Contains(view, "My Book") {
		t.Error("Expected view to contain success message and book title")
	}
}

func TestTUIFullFlow(t *testing.T) {
	// 1. Setup temporary clippings file
	rawLine := "Integration Test Book (Test Author)"
	content := rawLine + "\n- Your Highlight on page 1\n\nSome text\n==========\n"
	tmpFile, err := os.CreateTemp("", "integration_test.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(content)
	tmpFile.Close()

	// 2. Setup destination environment
	tmpDest, _ := os.MkdirTemp("", "dest_test")
	defer os.RemoveAll(tmpDest)
	os.Setenv("NOTE_PATH", tmpDest+"/")

	// 3. Initialize Model in SelectingSource state
	sourceItems := []list.Item{
		Item{TitleStr: "Default Path", DescStr: tmpFile.Name()},
	}
	m := &Model{
		State:      StateSelectingSource,
		SourceList: list.New(sourceItems, list.NewDefaultDelegate(), 100, 100),
	}

	// 4. Simulate selecting source (Enter)
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(*Model)
	if m.State != StateSelectingBook {
		t.Fatalf("Expected state StateSelectingBook, got %v", m.State)
	}

	// 5. Test Navigation in Book List (j/k)
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})

	// 6. Simulate selecting book (Enter)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(*Model)

	if !m.Done {
		t.Errorf("Expected model to be done, error: %v", m.Err)
	}
	if m.Choice != "Integration Test Book" {
		t.Errorf("Expected choice 'Integration Test Book', got %s", m.Choice)
	}

	// 8. Verify the file was actually written
	expectedFile := filepath.Join(tmpDest, "Integration Test Book", "Integration Test Book - Test Author.md")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Expected export file to exist at %s", expectedFile)
	}
}
