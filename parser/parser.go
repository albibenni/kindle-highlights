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
	var res []string
	if err != nil {
		log.Fatal("File not found:",err)
		return nil, err
	}
	buffer := make([]byte, 1024)
	for {
		line, err := file.Read(buffer)
		if err != nil {
			break
		}
		res = append(note.Content, string(buffer[:line]))
	}
	return res, nil
}
