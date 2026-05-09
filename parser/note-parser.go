package parser

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
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
	var fileDestination string
	if note.Author != "" {
		fileDestination = path + note.Title + "/" + note.Title + " - " + note.Author + ".md"
	} else {
		fileDestination = path + note.Title + "/" + note.Title + ".md"
	}
	note.FileDestination = fileDestination
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
	// remove (Z-Library) if exists
	formattedTitle = strings.ReplaceAll(str, "(Z-Library)", "")

	// get the author
	re := regexp.MustCompile(`\(([^)]+)\)`)

	matches := re.FindStringSubmatch(formattedTitle)
	if len(matches) == 0 {
		return "", formattedTitle
	}
	formattedTitle = strings.ReplaceAll(formattedTitle, "("+matches[1]+")", "")
	formattedTitle = strings.TrimSpace(formattedTitle)
	return matches[1], formattedTitle
}
