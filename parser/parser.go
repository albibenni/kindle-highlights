package parser

import (
	"log"
	"os"
)

type NoteInterface interface {
	GetAuthor() string
	GetContent() []string
	GetFileLocation() string
	ParseNotes() ([]string, error)
}

type Note struct {
	Author       string
	Content      []string
	FileLocation string
}

func (note *Note) ParseNotes() ([]string, error) {
	file, err := os.Open(note.FileLocation)
	if err != nil {
		log.Fatal("File not found:", err)
		return nil, err
	}
	for {
		buffer := make([]byte, 1024)
		n, err := file.Read(buffer)
		if err != nil {
			break
		}
		note.Content = append(note.Content, string(buffer[:n]))
	}
	return note.Content, nil
}
