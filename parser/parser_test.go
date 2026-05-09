package parser

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDiscoverBooks(t *testing.T) {
	content := "\uFEFF" + `The Great Gatsby (F. Scott Fitzgerald)
- Your Highlight on page 1 | Added on Friday, October 10, 2025 10:00:00 AM

Some content that should not be a title
==========
Second Book (Author B)
- Your Highlight on page 5 | Added on Friday, October 10, 2025 10:05:00 AM

More content
==========
`
	tmpDir, err := os.MkdirTemp("", "discovery_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	clippingsPath := filepath.Join(tmpDir, "My Clippings.txt")
	os.WriteFile(clippingsPath, []byte(content), 0644)

	note := &Note{FileLocation: clippingsPath}
	books, err := note.DiscoverBooks()
	if err != nil {
		t.Fatalf("DiscoverBooks() error = %v", err)
	}

	expected := []BookInfo{
		{Title: "The Great Gatsby", Author: "F. Scott Fitzgerald", RawLine: "The Great Gatsby (F. Scott Fitzgerald)"},
		{Title: "Second Book", Author: "Author B", RawLine: "Second Book (Author B)"},
	}

	if len(books) != len(expected) {
		t.Errorf("Got %d books, want %d", len(books), len(expected))
	}

	for i, b := range books {
		if b.Title != expected[i].Title || b.Author != expected[i].Author {
			t.Errorf("Book %d: Got %v, want %v", i, b, expected[i])
		}
	}
}

func TestGetAuthorAndFormatTitle(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantTitle  string
		wantAuthor string
	}{
		{
			name:       "Simple title",
			input:      "Mastery (Greene, Robert)",
			wantTitle:  "Mastery",
			wantAuthor: "Greene, Robert",
		},
		{
			name:       "A-nother title",
			input:      "Uncommon Sense Teaching (Barbara Oakley etc.) (something something) (Barbara Oakley, PhD)",
			wantTitle:  "Uncommon Sense Teaching",
			wantAuthor: "Barbara Oakley, PhD",
		},
		{
			name:       "Title with no author",
			input:      "No Author Book",
			wantTitle:  "No Author Book",
			wantAuthor: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAuthor, gotTitle := getAuthorAndFormatTitle(tt.input)
			if gotTitle != tt.wantTitle {
				t.Errorf("gotTitle = %v, want %v", gotTitle, tt.wantTitle)
			}
			if gotAuthor != tt.wantAuthor {
				t.Errorf("gotAuthor = %v, want %v", gotAuthor, tt.wantAuthor)
			}
		})
	}
}

func TestParseNotes(t *testing.T) {
	// Create a temporary clippings file with UTF-8 BOM
	content := "\uFEFF" + `The Great Gatsby (F. Scott Fitzgerald)
- Your Highlight on page 1 | Added on Friday, October 10, 2025 10:00:00 AM

In my younger and more vulnerable years
my father gave me some advice
==========
Random Book (Author)
- Your Highlight on page 5 | Added on Friday, October 10, 2025 10:05:00 AM

Some other content
==========
The Great Gatsby (F. Scott Fitzgerald)
- Your Highlight on page 2 | Added on Friday, October 10, 2025 10:10:00 AM

Second highlight
for the same book
==========
`
	tmpDir, err := os.MkdirTemp("", "parser_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	clippingsPath := filepath.Join(tmpDir, "My Clippings.txt")
	if err := os.WriteFile(clippingsPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	tests := []struct {
		name     string
		title    string
		expected []string
	}{
		{
			name:  "Find Great Gatsby highlights",
			title: "The Great Gatsby",
			expected: []string{
				"In my younger and more vulnerable years\nmy father gave me some advice",
				"Second highlight\nfor the same book",
			},
		},
		{
			name:     "Book not found",
			title:    "Non-existent Book",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			note := &Note{
				Title:             tt.title,
				FileLocation:      clippingsPath,
				IsLookingForTitle: true,
			}

			results, err := note.ParseNotes()
			if err != nil {
				t.Errorf("ParseNotes() error = %v", err)
				return
			}

			if !reflect.DeepEqual(results, tt.expected) {
				t.Errorf("ParseNotes() results = %v, want %v", results, tt.expected)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Normal Title", "Normal Title"},
		{"Title: Subtitle", "Title- Subtitle"},
		{"Path/With/Slash", "Path-With-Slash"},
		{"Illegal*?\"<>|", "Illegal-"},
		{"  Trim Me  ", "Trim Me"},
	}

	for _, tt := range tests {
		got := sanitizeFilename(tt.input)
		if got != tt.expected {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestNoteGettersAndSetters(t *testing.T) {
	// Set environment for setFileDestination
	tmpDir, _ := os.MkdirTemp("", "getter_test")
	defer os.RemoveAll(tmpDir)
	os.Setenv("NOTE_PATH", tmpDir+string(os.PathSeparator))

	note := &Note{
		Author:       "Greene, Robert",
		Title:        "Mastery",
		Content:      []string{"Highlight 1"},
		FileLocation: "/path/to/clippings.txt",
	}

	// Test Getters
	if a, _ := note.GetAuthor(); a != "Greene, Robert" {
		t.Errorf("GetAuthor() = %q, want Greene, Robert", a)
	}
	if t_, _ := note.GetTitle(); t_ != "Mastery" {
		t.Errorf("GetTitle() = %q, want Mastery", t_)
	}
	if f, _ := note.GetFileLocation(); f != "/path/to/clippings.txt" {
		t.Errorf("GetFileLocation() = %q, want /path/to/clippings.txt", f)
	}
	if c, _ := note.GetContent(); len(c) != 1 || c[0] != "Highlight 1" {
		t.Errorf("GetContent() = %v, want [Highlight 1]", c)
	}

	// Test Setter & File Destination
	note.setFileDestination()
	expectedDest := filepath.Join(tmpDir, "Mastery", "Mastery - Greene, Robert.md")
	if note.FileDestination != expectedDest {
		t.Errorf("setFileDestination() = %q, want %q", note.FileDestination, expectedDest)
	}

	// Test Empty Getters
	emptyNote := &Note{}
	if _, err := emptyNote.GetAuthor(); err == nil {
		t.Error("Expected error for empty author")
	}
	if _, err := emptyNote.GetTitle(); err == nil {
		t.Error("Expected error for empty title")
	}
	if _, err := emptyNote.GetFileLocation(); err == nil {
		t.Error("Expected error for empty file location")
	}
	if _, err := emptyNote.GetContent(); err == nil {
		t.Error("Expected error for empty content")
	}
}

func TestWriteFile(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "write_test")
	defer os.RemoveAll(tmpDir)
	os.Setenv("NOTE_PATH", tmpDir+string(os.PathSeparator))

	note := &Note{
		Title:   "Test Book",
		Author:  "Test Author",
		Content: []string{"Note 1", "Note 2"},
	}

	dest, err := note.WriteFile()
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Verify file content
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("Failed to read exported file: %v", err)
	}

	expectedContent := "# Test Book\n\nNote 1\n\n---\n\nNote 2"
	if string(data) != expectedContent {
		t.Errorf("Exported content = %q, want %q", string(data), expectedContent)
	}

	// Test WriteFile error (missing content)
	emptyNote := &Note{Title: "Empty Book"}
	if _, err := emptyNote.WriteFile(); err == nil {
		t.Error("Expected error for empty content in WriteFile")
	}

	// Test WriteFile error (invalid path)
	os.Setenv("NOTE_PATH", "/non/existent/path/that/should/fail")
	badPathNote := &Note{Title: "Bad Path", Content: []string{"Note"}}
	if _, err := badPathNote.WriteFile(); err == nil {
		t.Error("Expected error for invalid path in WriteFile")
	}
}

func TestUniteNotesError(t *testing.T) {
	_, err := uniteNotes([]string{}, "Empty Title")
	if err == nil {
		t.Error("Expected error from uniteNotes with empty lines")
	}
}

func TestParseNotesErrors(t *testing.T) {
	// Test missing title
	note := &Note{Title: ""}
	if _, err := note.ParseNotes(); err == nil {
		t.Error("Expected error for empty title")
	}

	// Test file not found
	note = &Note{Title: "Valid Title", FileLocation: "non-existent-file.txt"}
	if _, err := note.ParseNotes(); err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestParserInternalStates(t *testing.T) {
	note := &Note{Title: "Target"}
	var sb strings.Builder

	// Test processLineByState routing for skipping
	state := note.processLineByState(stateSkippingNote, "some line", &sb, "Target")
	if state != stateSkippingNote {
		t.Errorf("Expected state to remain stateSkippingNote, got %v", state)
	}

	// Test handleLookingForTitle with empty line
	state = note.handleLookingForTitle("", "Target")
	if state != stateLookingForTitle {
		t.Error("Expected empty line to remain in looking state")
	}

	// Test handleLookingForTitle when IsLookingForTitle is false
	note.IsLookingForTitle = false
	state = note.handleLookingForTitle("Target Book", "Target")
	if state != stateCollectingContent {
		t.Error("Expected to transition to collecting even if discovery is off")
	}
}

func TestWriteFileFailures(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "write_fail_test")
	defer os.RemoveAll(tmpDir)

	// Create a file where a directory should be to trigger MkdirAll failure
	conflictFile := filepath.Join(tmpDir, "Conflict")
	os.WriteFile(conflictFile, []byte("I am a file"), 0644)

	note := &Note{
		Title:   "Conflict", // This will try to create a directory named 'Conflict'
		Content: []string{"test"},
	}
	os.Setenv("NOTE_PATH", tmpDir+string(os.PathSeparator))
	
	_, err := note.WriteFile()
	if err == nil {
		t.Error("Expected error when directory creation conflicts with a file")
	}
}

func TestCheckWritePermissionFailure(t *testing.T) {
	// This is platform specific, but on Unix we can test a non-writable dir
	tmpDir, _ := os.MkdirTemp("", "perm_test")
	defer os.RemoveAll(tmpDir)
	
	readonlyDir := filepath.Join(tmpDir, "readonly")
	os.Mkdir(readonlyDir, 0555) // Read and execute only
	
	err := checkWritePermission(readonlyDir)
	if err == nil {
		t.Error("Expected error for checkWritePermission on readonly directory")
	}
}
