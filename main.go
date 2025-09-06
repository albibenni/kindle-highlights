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
		Title:        "Linux Basics for Hackers",
		Content:      []string{},
		FileLocation: currentDir + "/test-file/My Clippings.txt",
	}

	_, err = myNote.ParseNotes()
	// for _, w := range res {
	// 	fmt.Printf("RESULT: %s\n", w)
	// }
	if err != nil {
		fmt.Println("Error parsing notes:", err)
		return
	}
	fileDest, err := myNote.WriteFile()
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}
	fmt.Printf("RESULT : %s\n", fileDest)
	fmt.Printf("RESULT len: %d\n", len(fileDest))
}
