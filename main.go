package main

import (
	"fmt"
	"log"
	"os"
	"strings"

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

	// START Logic
	myNote := parser.Note{
		Author:            "",
		Title:             "",
		Content:           []string{},
		FileLocation:      types.ClippingPath.Value(),
		FileDestination:   "",
		IsLookingForTitle: true,
	}

	handleArgs(&myNote, args)
	fmt.Printf("FileLocation: %s", myNote.FileLocation)
	_, err := myNote.ParseNotes()
	if err != nil {
		fmt.Println("Error parsing notes:", err)
		return
	}
	_, err = myNote.WriteFile()
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}
}

func handleArgs(note *parser.Note, args []string) {
	switch len(args) {
	case 1:
		log.Fatal("Should add title name")
		break
	case 2:
		// helper
		if strings.ToLower(args[1]) == "help" {
			fmt.Printf("1. %sOne Arg%s: \n - %shelp%s: helper\n", types.Bold, types.Reset, types.Italic, types.Reset)
			fmt.Printf("2. %sOne Arg%s: \n - %s<title-name>%s: title to search FileLocation path default\n", types.Bold, types.Reset, types.Italic, types.Reset)
			fmt.Printf("3. %sTwo Args%s: \n - %stest%s: to setup test FileLocation path\n - %s<title-name>%s: title to search\n", types.Bold, types.Reset, types.Italic, types.Reset, types.Italic, types.Reset)
			fmt.Printf("4. %sTwo Args%s: \n - %s<title-name>%s: title to search\n - %s<file-location>%s: absolute path to the file\n", types.Bold, types.Reset, types.Italic, types.Reset, types.Italic, types.Reset)
			os.Exit(0)
		}
	case 3:

		if strings.ToLower(args[1]) == "test" {
			currentDir, err := os.Getwd()
			if err != nil {
				fmt.Println("Error getting current directory:", err)
				return
			}
			note.FileLocation = currentDir + "/test-file/My Clippings.txt"
			note.Title = args[2]
		} else {
			note.Title = args[1]
			note.FileLocation = args[2]
		}
	}
}
