package parser

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func uniteNotes(lines []string, title string) (string, error) {
	if len(lines) == 0 {
		return "", errors.New("No notes provided - no lines obtained")
	}
	res := "# " + title + "\n\n" + strings.Join(lines, "\n\n---\n\n")
	return res, nil //TODO: check line breaker
}

func writeContentToFile(filePath string, content string) error {
	// Ensure directory exists
	// Check permissions
	// Write the real file using a buffer
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	if err := checkWritePermission(dir); err != nil {
		return fmt.Errorf("no write permission for directory %s: %w", dir, err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer func() {
		if flushErr := writer.Flush(); flushErr != nil {
			err = fmt.Errorf("failed to flush buffer: %w", flushErr)
		}
	}()

	_, err = writer.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write content to %s: %w", filePath, err)
	}

	return nil
}

func checkWritePermission(dir string) error {
	testFile := filepath.Join(dir, ".write_test_tmp")
	file, err := os.Create(testFile)
	if err != nil {
		return err
	}
	file.Close()
	os.Remove(testFile)
	return nil
}
