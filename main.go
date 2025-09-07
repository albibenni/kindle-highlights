package main

import (
	"fmt"
	"log"
	"os"

	"github.com/albibenni/kindle-highlights/parser"
	"github.com/albibenni/kindle-highlights/types"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Printf("Hello, Note!\n")
	envFile := types.GetEnvFile()
	if envFile == "wrong pc" {
		log.Fatal("You chose wrong!")
	}
	godotenv.Load(types.GetEnvFile())

	args := os.Args
	fmt.Printf("ARGS %v", args)
	// START Logic
	currentDir, err := os.Getwd() //TODO: add from stdin - option 2nd arg = path arg[1] else default to env?
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}
	myNote := parser.Note{
		Author:            "",
		Title:             "Linux Basics for Hackers",
		Content:           []string{},
		FileLocation:      currentDir + "/test-file/My Clippings.txt",
		FileDestination:   "",
		IsLookingForTitle: true,
	}

	_, err = myNote.ParseNotes()
	if err != nil {
		fmt.Println("Error parsing notes:", err)
		return
	}
	_, err = myNote.WriteFile()
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}
	// fmt.Printf("RESULT : %s\n", fileDest)
	// fmt.Printf("RESULT len: %d\n", len(fileDest))
}
