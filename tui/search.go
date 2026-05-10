package tui

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) updatePathSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		var path string
		if len(m.SearchResults) > 0 && m.SearchIndex < len(m.SearchResults) {
			path = m.SearchResults[m.SearchIndex]
		} else {
			path = m.TextInput.Value()
		}

		if path != "" {
			if m.State == StateCustomPathInput {
				m.Path = path
				return m.LoadBooks()
			} else {
				m.BasePath = path
				return m.prepareExport()
			}
		}
		return m, nil

	case "esc":
		if m.State == StateCustomPathInput {
			m.State = StateSelectingSource
		} else {
			m.State = StateSelectingDest
		}
		m.Searching = false
		m.SearchResults = nil
		m.TextInput.Blur()
		m.TextInput.Reset()
		return m, nil

	case "up":
		if m.SearchIndex > 0 {
			m.SearchIndex--
		}
		return m, nil
	case "down":
		if m.SearchIndex < len(m.SearchResults)-1 {
			m.SearchIndex++
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.TextInput, cmd = m.TextInput.Update(msg)

	query := m.TextInput.Value()

	// Allow searching with 1 char if it's a path starter, otherwise min 3 chars
	shouldSearch := len(query) >= 3 || (len(query) >= 1 && (strings.HasPrefix(query, "/") || strings.HasPrefix(query, "~")))

	if shouldSearch {
		m.Searching = true
		m.SearchID++
		isDir := m.State == StateCustomDestInput
		return m, tea.Batch(cmd, m.searchSystem(query, m.SearchID, isDir))
	}

	m.Searching = false
	m.SearchResults = nil
	return m, cmd
}

func (m *Model) searchSystem(query string, id int, isDir bool) tea.Cmd {
	return func() tea.Msg {
		// 1. Setup a context with a strict 2-second timeout
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		searchRoot, _ := os.UserHomeDir()
		displayQuery := query

		// If it looks like an absolute path, try to use it as root
		if strings.HasPrefix(query, "/") {
			if info, err := os.Stat(query); err == nil && info.IsDir() {
				searchRoot = query
				displayQuery = ""
			} else {
				parent := filepath.Dir(query)
				if info, err := os.Stat(parent); err == nil && info.IsDir() {
					searchRoot = parent
					displayQuery = filepath.Base(query)
				}
			}
		}

		// Use rg --files to get a list of all files, then we'll filter them in Go.
		cmd := exec.CommandContext(ctx, "rg",
			"--files",
			"--hidden",
			"--max-depth", "6",
			"-j1",
			"-g", "!Library",
			"-g", "!.Trash",
			"-g", "!node_modules",
			"-g", "!.git",
			"-g", "!.vim",
			"-g", "!.cache",
			searchRoot)

		output, err := cmd.Output()
		if err != nil {
			return SearchResultMsg{ID: id, Results: []string{}}
		}

		lines := strings.Split(string(output), "\n")
		results := []string{}
		dirMap := make(map[string]bool)
		isQueryLower := displayQuery == strings.ToLower(displayQuery)

		for _, line := range lines {
			if line == "" {
				continue
			}

			fullPath := line
			if !filepath.IsAbs(line) {
				fullPath = filepath.Join(searchRoot, line)
			}

			// For efficiency, check if the full path contains the query at all first
			matchFound := false
			if displayQuery == "" {
				matchFound = true
			} else if isQueryLower {
				matchFound = strings.Contains(strings.ToLower(fullPath), strings.ToLower(displayQuery))
			} else {
				matchFound = strings.Contains(fullPath, displayQuery)
			}

			if !matchFound {
				continue
			}

			if isDir {
				parts := strings.Split(fullPath, string(os.PathSeparator))
				currentPath := ""
				if strings.HasPrefix(fullPath, string(os.PathSeparator)) {
					currentPath = string(os.PathSeparator)
				}

				for _, part := range parts {
					if part == "" {
						continue
					}
					currentPath = filepath.Join(currentPath, part)

					segmentMatch := false
					if displayQuery == "" {
						segmentMatch = true
					} else {
						// If the query contains a slash, we match against the cumulative path.
						targetToMatch := part
						if strings.Contains(displayQuery, string(os.PathSeparator)) {
							targetToMatch = currentPath
						}

						if isQueryLower {
							segmentMatch = strings.Contains(strings.ToLower(targetToMatch), strings.ToLower(displayQuery))
						} else {
							segmentMatch = strings.Contains(targetToMatch, displayQuery)
						}
					}

					if segmentMatch {
						if !dirMap[currentPath] {
							dirMap[currentPath] = true
							results = append(results, currentPath)
						}
					}
				}
			} else {
				if !dirMap[fullPath] {
					dirMap[fullPath] = true
					results = append(results, fullPath)
				}
			}

			if len(results) >= 10 {
				break
			}
		}

		return SearchResultMsg{ID: id, Results: results}
	}
}
