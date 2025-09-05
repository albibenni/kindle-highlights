package main

import (
	"fmt"
	"github.com/albibenni/kindle-highlights/parser"
	"os"
)

func main() {
	// This is a placeholder for the main function.
	fmt.Printf("Hello, Note!\n")
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}
	myNote := parser.Note{
		Author:       "John Doe",
		Content:      []string{},
		FileLocation: currentDir + "/My Clipping.txt",
	}

	myNote.ParseNotes()
}
