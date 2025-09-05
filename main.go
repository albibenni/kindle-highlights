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
		Author:       "The TCP/IP Guide: A Comprehensive, Illustrated Internet Protocols Reference (Charles M. Kozierok)",
		Content:      []string{},
		FileLocation: currentDir + "/test-file/My Clippings.txt",
	}

	res, err := myNote.ParseNotes()
	fmt.Printf("RESULT: %v", res)
}
