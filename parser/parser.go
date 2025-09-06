package parser

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

type NoteInterface interface {
	GetAuthor() string
	GetTitle() string
	GetContent() []string
	GetFileLocation() string
	ParseNotes() ([]string, error)
}

type Note struct {
	Author       string
	Title        string
	Content      []string
	FileLocation string
}

func (note *Note) ParseNotes() ([]string, error) {
	file, err := os.Open(note.FileLocation)
	scanner := bufio.NewScanner(file)
	if err != nil {
		log.Fatal("File not found:", err)
		return nil, err
	}
	// for {
	// buffer := make([]byte, 1024)
	// n, err := file.Read(buffer)
	isNextNote := true
	for scanner.Scan() {
		line := scanner.Text()
		//note.Content = append(note.Content, line) // remember to trim in case \r\n to mac/linux format
		res, isNextNotee, err := checkNotesByTitle(note.Title, line, isNextNote)
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

func checkNotesByTitle(title string, buffLine string, isNextNote bool) (result string, isNextNotee bool, err error) {
	switch {
	// taking in consideration cases where == lines are duplicates so isNextNote is true - not considered
	// and return true already
	case isNextNote:
		titleTrimmed := strings.TrimSpace(title)
		if len(titleTrimmed) == 0 {
			return "", false, errors.New("Title not defined")
		}
		if strings.Contains(buffLine, titleTrimmed) {
			fmt.Printf("Found new note of Title %s\n", titleTrimmed)
			return "", false, nil
		}
		return "", isNextNote, nil
	case buffLine == "==========\r", buffLine == "==========\n", buffLine == "==========", buffLine == "==========\r\n":
		fmt.Printf("delimiter not counted - new note coming")
		return "", true, nil
	case len(strings.TrimSpace(buffLine)) == 0:
		fmt.Println("Empty note line")
		return "", isNextNote, nil
	default:
		log.Fatal("Not handled")
		return "", false, errors.New("Unhandled")
	}
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
