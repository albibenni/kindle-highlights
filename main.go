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
		Author:       "Linux Basics for Hackers",
		Title:       "Linux Basics for Hackers",
		Content:      []string{},
		FileLocation: currentDir + "/test-file/My Clippings.txt",
	}

	res, err := myNote.ParseNotes()
	for _, w := range res {
		fmt.Printf("RESULT: %s\n", w)
	}
	fmt.Printf("RESULT len: %d\n", len(res))
}
