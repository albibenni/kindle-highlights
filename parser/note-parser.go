package parser

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/albibenni/kindle-highlights/types"
)

type NoteInterface interface {
	GetAuthor() string
	GetTitle() string
	GetContent() []string
	GetFileLocation() string
	ParseNotes() ([]string, error)
	WriteFile() (string, error)
}

type Note struct {
	Author            string
	Title             string
	Content           []string
	FileLocation      string
	FileDestination   string
	IsLookingForTitle bool
}

type parserState int

const (
	stateLookingForTitle parserState = iota
	stateCollectingContent
	stateSkippingNote
)

func (note *Note) DiscoverBooks() ([]BookInfo, error) {
	file, err := os.Open(note.FileLocation)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	booksMap := make(map[string]BookInfo)
	var books []BookInfo
	isNextLineTitle := true

	for scanner.Scan() {
		line := note.prepareLine(scanner.Text())
		isNextLineTitle = note.handleDiscoveryLine(line, isNextLineTitle, booksMap, &books)
	}

	return books, nil
}

func (note *Note) handleDiscoveryLine(line string, isNextLineTitle bool, booksMap map[string]BookInfo, books *[]BookInfo) bool {
	if line == "==========" {
		return true
	}

	if isNextLineTitle && line != "" {
		note.addUniqueBook(line, booksMap, books)
		return false
	}

	return isNextLineTitle
}

func (note *Note) addUniqueBook(line string, booksMap map[string]BookInfo, books *[]BookInfo) {
	if _, exists := booksMap[line]; !exists {
		author, title := getAuthorAndFormatTitle(line)
		if title != "" {
			info := BookInfo{Title: title, Author: author, RawLine: line}
			booksMap[line] = info
			*books = append(*books, info)
		}
	}
}

type BookInfo struct {
	Title   string
	Author  string
	RawLine string
}

func (note *Note) ParseNotes() ([]string, error) {
	file, err := os.Open(note.FileLocation)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentHighlight strings.Builder
	state := stateLookingForTitle
	titleLookup := strings.TrimSpace(note.Title)

	if len(titleLookup) == 0 {
		return nil, errors.New("title not defined")
	}

	for scanner.Scan() {
		line := note.prepareLine(scanner.Text())

		if line == "==========" {
			state = note.finalizeNote(state, &currentHighlight)
			continue
		}

		state = note.processLineByState(state, line, &currentHighlight, titleLookup)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return note.Content, nil
}

func (note *Note) prepareLine(line string) string {
	line = strings.TrimPrefix(line, "\uFEFF")
	return strings.TrimRight(line, "\r\n")
}

func (note *Note) finalizeNote(state parserState, currentHighlight *strings.Builder) parserState {
	if state == stateCollectingContent && currentHighlight.Len() > 0 {
		note.Content = append(note.Content, strings.TrimSpace(currentHighlight.String()))
		currentHighlight.Reset()
	}
	return stateLookingForTitle
}

func (note *Note) processLineByState(state parserState, line string, currentHighlight *strings.Builder, titleLookup string) parserState {
	switch state {
	case stateLookingForTitle:
		return note.handleLookingForTitle(line, titleLookup)
	case stateCollectingContent:
		note.handleCollectingContent(line, currentHighlight)
		return stateCollectingContent
	case stateSkippingNote:
		return stateSkippingNote
	default:
		return stateLookingForTitle
	}
}

func (note *Note) handleLookingForTitle(line string, titleLookup string) parserState {
	if line == "" {
		return stateLookingForTitle
	}
	if strings.Contains(line, titleLookup) {
		if note.IsLookingForTitle {
			note.setTitleAndAuthor(line)
		}
		return stateCollectingContent
	}
	return stateSkippingNote
}

func (note *Note) handleCollectingContent(line string, currentHighlight *strings.Builder) {
	lowerLine := strings.ToLower(line)
	if strings.Contains(lowerLine, "your highlight") ||
		strings.Contains(lowerLine, "your note") ||
		line == "" {
		return
	}

	if currentHighlight.Len() > 0 {
		currentHighlight.WriteString("\n")
	}
	currentHighlight.WriteString(line)
}

func (note *Note) WriteFile() (string, error) {

	note.setFileDestination()

	if len(note.FileDestination) == 0 {
		return "", errors.New("File Destination not present")
	}
	unitedNotes, err := uniteNotes(note.Content, note.Title)
	if err != nil {
		return "", err
	}

	err = writeContentToFile(note.FileDestination, unitedNotes)
	if err != nil {
		return "", err
	}
	return note.FileDestination, nil
}

func (note Note) GetAuthor() (string, error) {
	if len(strings.TrimSpace(note.Author)) == 0 {
		err := errors.New("Author not defined")
		return "", err
	}
	return note.Author, nil
}
func (note Note) GetTitle() (string, error) {
	if len(strings.TrimSpace(note.Title)) == 0 {
		err := errors.New("Title not defined")
		return "", err
	}
	return note.Title, nil
}

func (note Note) GetFileLocation() (string, error) {
	if len(strings.TrimSpace(note.FileLocation)) == 0 {
		err := errors.New("FileLocation not defined")
		return "", err
	}
	return note.FileLocation, nil
}

func (note Note) GetContent() ([]string, error) {
	if len(note.Content) == 0 {
		err := errors.New("Content not defined")
		return nil, err
	}
	return note.Content, nil
}

func (note *Note) setFileDestination() {
	path := types.NotePath.Value()
	
	safeTitle := sanitizeFilename(note.Title)
	safeAuthor := sanitizeFilename(note.Author)

	var fileDestination string
	if safeAuthor != "" {
		fileDestination = filepath.Join(path, safeTitle, safeTitle+" - "+safeAuthor+".md")
	} else {
		fileDestination = filepath.Join(path, safeTitle, safeTitle+".md")
	}
	note.FileDestination = fileDestination
}

func sanitizeFilename(name string) string {
	// Illegal characters in Windows: \ / : * ? " < > |
	// We'll replace them with a hyphen or remove them
	replacer := strings.NewReplacer(
		":", "-",
		"/", "-",
		"\\", "-",
		"*", "",
		"?", "",
		"\"", "",
		"<", "",
		">", "",
		"|", "-",
	)
	return strings.TrimSpace(replacer.Replace(name))
}

func (note *Note) setTitleAndAuthor(buffLine string) {
	if strings.Contains(buffLine, note.Title) {
		author, formattedTitle := getAuthorAndFormatTitle(buffLine)
		fmt.Printf("Title: %s,\nAuthor: %s\n", formattedTitle, author)
		note.Author = author
		note.Title = formattedTitle
		note.IsLookingForTitle = false
	}
}

func getAuthorAndFormatTitle(str string) (author string, formattedTitle string) {
	// 1. Remove common noise markers
	str = strings.ReplaceAll(str, "(Z-Library)", "")
	str = strings.TrimSpace(str)

	// 2. Loop to find and extract the last parenthesis group as the author
	// and keep everything else as title.
	// We handle titles like "Title (Info) (Author)" by iteratively stripping from the end.
	tempStr := str
	for {
		lastOpen := strings.LastIndex(tempStr, "(")
		lastClose := strings.LastIndex(tempStr, ")")

		if lastOpen != -1 && lastClose > lastOpen && lastClose == len(tempStr)-1 {
			// Found a potential author at the very end
			if author == "" {
				author = tempStr[lastOpen+1 : lastClose]
			}
			tempStr = strings.TrimSpace(tempStr[:lastOpen])
			continue
		}
		break
	}

	if author != "" {
		return author, tempStr
	}

	return "", str
}
