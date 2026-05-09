package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/albibenni/kindle-highlights/parser"
	"github.com/albibenni/kindle-highlights/types"
	"github.com/joho/godotenv"
)

func main() {
	loadEnvironment()

	testMode := flag.Bool("test", false, "Use the local test clippings file")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Hello, Note!\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <title> [file-location]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nFlags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s \"The Great Gatsby\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -test \"The Great Gatsby\"\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s \"The Great Gatsby\" /path/to/clippings.txt\n", os.Args[0])
	}
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	myNote := parser.Note{
		Title:             args[0],
		FileLocation:      types.ClippingPath.Value(),
		IsLookingForTitle: true,
	}

	// Handle test mode or custom file location
	if *testMode {
		currentDir, err := os.Getwd()
		if err != nil {
			log.Fatalf("Error getting current directory: %v", err)
		}
		myNote.FileLocation = filepath.Join(currentDir, "test-file", "My Clippings.txt")
	} else if len(args) > 1 {
		myNote.FileLocation = args[1]
	}

	if myNote.FileLocation == "" {
		log.Fatal("File location not defined. Please set CLIPPING_PATH in your .env or provide a file path as an argument.")
	}

	fmt.Printf("FileLocation: %s\n", myNote.FileLocation)

	if _, err := myNote.ParseNotes(); err != nil {
		log.Fatalf("Error parsing notes: %v", err)
	}

	filePath, err := myNote.WriteFile()
	if err != nil {
		log.Fatalf("Error writing file: %v", err)
	}

	fmt.Printf("\n✓ Successfully wrote notes to: %s\n", filePath)
}

func loadEnvironment() {
	envFile := types.GetEnvFile()
	if envFile == "wrong pc" {
		log.Fatal("Windows is not supported yet.")
	}

	// Try loading from local file first (e.g. mac.env, linux.env)
	if err := godotenv.Load(envFile); err != nil {
		// Fallback to global config: ~/.config/kindle-highlights/.env
		homeDir, err := os.UserHomeDir()
		if err == nil {
			configPath := filepath.Join(homeDir, ".config", "kindle-highlights", ".env")
			_ = godotenv.Load(configPath)
		}
	}
}
