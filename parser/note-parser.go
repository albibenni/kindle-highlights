package parser

import (
	"bufio"
	"errors"
	"fmt"
	"log"
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

func (note *Note) ParseNotes() ([]string, error) {
	file, err := os.Open(note.FileLocation)
	scanner := bufio.NewScanner(file)
	if err != nil {
		log.Fatal("File not found:", err)
		return nil, err
	}
	isNextNote := true
	titleLookup := note.Title
	for scanner.Scan() {
		line := scanner.Text()
		if note.IsLookingForTitle {
			note.setTitleAndAuthor(line)
			fmt.Println("Looking for title: ", note.Title)
		}
		res, isNextNotee, err := checkNotesByTitle(titleLookup, line, isNextNote)
		if err != nil {
			return nil, err
		}
		isNextNote = isNextNotee
		if len(res) > 0 {
			note.Content = append(note.Content, res) // remember to trim in case \r\n to mac/linux format
		}
	}
	return note.Content, nil
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
		log.Fatal("File not found:", err)
		return "", err
	}
	return note.Author, nil
}
func (note Note) GetTitle() (string, error) {
	if len(strings.TrimSpace(note.Title)) == 0 {
		err := errors.New("Title not defined")
		log.Fatal("File not found:", err)
		return "", err
	}
	return note.Title, nil
}

func (note Note) GetFileLocation() (string, error) {
	if len(strings.TrimSpace(note.FileLocation)) == 0 {
		err := errors.New("FileLocation not defined")
		log.Fatal("File not found:", err)
		return "", err
	}
	return note.FileLocation, nil
}

func (note Note) GetContent() ([]string, error) {
	if len(note.Content) == 0 {
		err := errors.New("Content not defined")
		log.Fatal("File not found:", err)
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
		author, formattedTitle := getAuthorAndFormatTitle(note.Title)
		fmt.Println(formattedTitle, author)
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
