package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/albibenni/kindle-highlights/tui"
	"github.com/albibenni/kindle-highlights/types"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
)

func main() {
	loadEnvironment()

	clippingPath := types.ClippingPath.Value()
	sourceItems := []list.Item{
		tui.Item{TitleStr: "Default Path", DescStr: clippingPath},
		tui.Item{TitleStr: "Custom Path", DescStr: "Manually enter a path to your clippings file"},
	}

	ti := textinput.New()
	ti.Placeholder = "/path/to/My Clippings.txt"
	ti.CharLimit = 255
	ti.Width = 50

	m := tui.Model{
		State:      tui.StateSelectingSource,
		SourceList: list.New(sourceItems, list.NewDefaultDelegate(), 0, 0),
		TextInput:  ti,
	}
	m.SourceList.Title = "Kindle Highlights - Select Source"
	m.SourceList.SetShowStatusBar(false)
	m.SourceList.SetFilteringEnabled(false)

	p := tea.NewProgram(&m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func loadEnvironment() {
	envFile := types.GetEnvFile()
	if envFile == "wrong pc" {
		log.Fatal("Windows is not supported yet.")
	}

	if err := godotenv.Load(envFile); err != nil {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			configPath := filepath.Join(homeDir, ".config", "kindle-highlights", ".env")
			_ = godotenv.Load(configPath)
		}
	}
}
