package parser

import (
	"os"
	"path/filepath"
	"reflect"
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
