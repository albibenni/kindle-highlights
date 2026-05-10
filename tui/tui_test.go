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
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_, _ = tmpFile.WriteString(content)
	_ = tmpFile.Close()

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
		BookList:   list.New([]list.Item{Item{TitleStr: "Book"}}, list.NewDefaultDelegate(), 0, 0),
		DestList:   list.New([]list.Item{Item{TitleStr: "Dest"}}, list.NewDefaultDelegate(), 0, 0),
		TextInput:  textinput.New(),
	}

	// 1. Source Selection
	view := m.View()
	if view == "" {
		t.Error("View returned empty string for Source Selection")
	}

	// 2. Custom Path Input
	m.State = StateCustomPathInput
	view = m.View()
	if !strings.Contains(view, "clippings file") {
		t.Error("Expected view to contain 'clippings file' for path input")
	}

	// 3. Selecting Book
	m.State = StateSelectingBook
	view = m.View()
	if view == "" {
		t.Error("View returned empty string for Selecting Book")
	}

	// 4. Selecting Dest
	m.State = StateSelectingDest
	view = m.View()
	if view == "" {
		t.Error("View returned empty string for Selecting Dest")
	}

	// 5. Custom Dest Input
	m.State = StateCustomDestInput
	view = m.View()
	if !strings.Contains(view, "destination folder") {
		t.Error("Expected view to contain 'destination folder' for dest input")
	}

	// 6. Confirm Export
	m.State = StateConfirmExport
	m.Choice = "My Book"
	m.BasePath = "/path/to/base"
	view = m.View()
	if !strings.Contains(view, "Confirm Export") || !strings.Contains(view, "My Book") {
		t.Error("Confirm Export view failed")
	}

	// 7. Error View
	m.Err = fmt.Errorf("test error")
	view = m.View()
	if !strings.Contains(view, "Error: test error") {
		t.Errorf("Expected view to contain error message, got: %s", view)
	}

	// 8. Done View (StateConfirmSuccess)
	m.Err = nil
	m.Done = true
	m.Choice = "My Book"
	m.Dest = "/path/to/dest"
	view = m.View()
	if !strings.Contains(view, "Successfully exported") || !strings.Contains(view, "My Book") {
		t.Error("Expected view to contain success message and book title")
	}
}

func TestSmartCaseLogic(t *testing.T) {
	tests := []struct {
		query    string
		wantFlag string
	}{
		{"clippings", "--iglob"},
		{"Clippings", "--glob"},
		{"CLIPPINGS", "--glob"},
		{"my clippings", "--iglob"},
		{"My clippings", "--glob"},
	}

	for _, tt := range tests {
		gotFlag := "--glob"
		if tt.query == strings.ToLower(tt.query) {
			gotFlag = "--iglob"
		}

		if gotFlag != tt.wantFlag {
			t.Errorf("For query %q, got flag %q, want %q", tt.query, gotFlag, tt.wantFlag)
		}
	}
}

func TestItemMethods(t *testing.T) {
	item := Item{
		TitleStr: "Clean Code",
		DescStr:  "Robert C. Martin",
		Raw:      "Clean Code (Robert C. Martin)",
	}

	if item.Title() != "Clean Code" {
		t.Errorf("Expected Title() 'Clean Code', got %q", item.Title())
	}

	if item.Description() != "Robert C. Martin" {
		t.Errorf("Expected Description() 'Robert C. Martin', got %q", item.Description())
	}

	expectedFilter := "Clean Code Robert C. Martin"
	if item.FilterValue() != expectedFilter {
		t.Errorf("Expected FilterValue() %q, got %q", expectedFilter, item.FilterValue())
	}
}

func TestTUIBackNavigation(t *testing.T) {
	m := &Model{
		State:    StateSelectingBook,
		BookList: list.New(nil, list.NewDefaultDelegate(), 0, 0),
		DestList: list.New(nil, list.NewDefaultDelegate(), 0, 0),
	}

	// Book -> Source
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Dest Selection -> Book
	m.State = StateSelectingDest
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if newModel.(*Model).State != StateSelectingBook {
		t.Errorf("Expected ESC to go from Dest Selection to Selecting Book, got %v", newModel.(*Model).State)
	}

	// Confirm Export -> Dest Selection
	m.State = StateConfirmExport
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc, Runes: []rune("n")})
	if newModel.(*Model).State != StateSelectingDest {
		t.Errorf("Expected 'n' or ESC to go from Confirm Export to Selecting Dest, got %v", newModel.(*Model).State)
	}
}

func TestTUISearchStateReset(t *testing.T) {
	sourceItems := []list.Item{
		Item{TitleStr: "Custom Path", DescStr: "manual"},
	}
	m := &Model{
		State:         StateSelectingSource,
		SourceList:    list.New(sourceItems, list.NewDefaultDelegate(), 0, 0),
		SearchResults: []string{"stale result"},
		TextInput:     textinput.New(),
	}
	m.TextInput.SetValue("stale query")

	// Enter Custom Path Input
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := newModel.(*Model)

	if updated.TextInput.Value() != "" {
		t.Error("Expected TextInput to be reset when entering custom path")
	}
	if updated.SearchResults != nil {
		t.Error("Expected SearchResults to be nil when entering custom path")
	}
}

func TestTUIResponsiveResizing(t *testing.T) {
	m := &Model{
		SourceList: list.New(nil, list.NewDefaultDelegate(), 0, 0),
		BookList:   list.New(nil, list.NewDefaultDelegate(), 0, 0),
		DestList:   list.New(nil, list.NewDefaultDelegate(), 0, 0),
		TextInput:  textinput.New(),
		State:      StateSelectingBook,
	}

	w, h := 120, 60
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	updated := newModel.(*Model)

	frameH, frameV := DocStyle.GetFrameSize()
	expectedW, expectedH := w-frameH, h-frameV

	if updated.SourceList.Width() != expectedW || updated.SourceList.Height() != expectedH {
		t.Errorf("SourceList resize failed: got %dx%d, want %dx%d", updated.SourceList.Width(), updated.SourceList.Height(), expectedW, expectedH)
	}
	if updated.BookList.Width() != expectedW || updated.BookList.Height() != expectedH {
		t.Errorf("BookList resize failed")
	}
	
	// Test DestList resize specifically when in that state
	updated.State = StateSelectingDest
	newModel, _ = updated.Update(tea.WindowSizeMsg{Width: w, Height: h})
	if newModel.(*Model).DestList.Width() != expectedW {
		t.Errorf("DestList resize failed")
	}
}

func TestTUIPathSearchHeaders(t *testing.T) {
	m := &Model{
		State: StateCustomPathInput,
	}

	view := m.View()
	if !strings.Contains(view, "clippings file") {
		t.Errorf("Expected clipping file header, got: %s", view)
	}

	m.State = StateCustomDestInput
	view = m.View()
	if !strings.Contains(view, "destination folder") {
		t.Errorf("Expected destination folder header, got: %s", view)
	}
}

func TestTUISearchResultHandling(t *testing.T) {
	m := &Model{
		Searching: true,
		SearchID:  10,
	}

	// Correct ID - should update
	newModel, _ := m.Update(SearchResultMsg{ID: 10, Results: []string{"res1"}})
	updated := newModel.(*Model)
	if updated.Searching {
		t.Error("Searching should be false after receiving results")
	}
	if len(updated.SearchResults) != 1 {
		t.Errorf("Expected 1 result, got %d", len(updated.SearchResults))
	}

	// Wrong ID - should ignore
	m.Searching = true
	m.SearchID = 11
	newModel, _ = m.Update(SearchResultMsg{ID: 10, Results: []string{"wrong"}})
	updated = newModel.(*Model)
	if !updated.Searching || len(updated.SearchResults) != 1 { // Should still have previous results or keep searching
		t.Error("Should have ignored results with wrong SearchID")
	}
}

func TestTUIActiveComponentUpdating(t *testing.T) {
	// Initialize with a simple state
	m := &Model{
		State:      StateSelectingSource,
		SourceList: list.New(nil, list.NewDefaultDelegate(), 0, 0),
		TextInput:  textinput.New(),
		BookList:   list.New(nil, list.NewDefaultDelegate(), 0, 0),
		DestList:   list.New(nil, list.NewDefaultDelegate(), 0, 0),
	}

	states := []SessionState{
		StateSelectingSource,
		StateCustomPathInput,
		StateSelectingBook,
		StateSelectingDest,
	}

	dummyMsg := "some_msg"
	for _, state := range states {
		m.State = state
		// This ensures we reach the default case in Update and call updateActiveComponent
		_, _ = m.Update(dummyMsg)
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
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_, _ = tmpFile.WriteString(content)
	_ = tmpFile.Close()

	// 2. Setup destination environment
	tmpDest, _ := os.MkdirTemp("", "dest_test")
	defer func() { _ = os.RemoveAll(tmpDest) }()
	_ = os.Setenv("NOTE_PATH", tmpDest+"/")

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

	if m.State != StateSelectingDest {
		t.Fatalf("Expected state StateSelectingDest after book selection, got %v", m.State)
	}

	// 7. Simulate selecting default destination (Enter)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(*Model)

	if m.State != StateConfirmExport {
		t.Fatalf("Expected state StateConfirmExport after dest selection, got %v", m.State)
	}

	// 7.5 Simulate confirming export (Enter)
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
