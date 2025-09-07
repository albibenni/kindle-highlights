package parser

import (
	"errors"
	"strings"
)

func checkNotesByTitle(title string, buffLine string, isNextNote bool) (result string, isNextNotee bool, err error) {
	switch {
	// taking in consideration cases where == lines are duplicates so isNextNote is true - not considered
	// and return true already
	case isNextNote:
		titleTrimmed := strings.TrimSpace(title)
		if len(titleTrimmed) == 0 {
			return "", false, errors.New("Title not defined")
		}
		if strings.Contains(buffLine, titleTrimmed) {
			//fmt.Printf("Found new note of Title %s\n", titleTrimmed)
			return "", false, nil
		}
		return "", isNextNote, nil
	case buffLine == "==========\r", buffLine == "==========\n", buffLine == "==========", buffLine == "==========\r\n":
		//fmt.Printf("delimiter not counted - new note coming")
		return "", true, nil
	case len(strings.TrimSpace(buffLine)) == 0:
		//fmt.Println("Empty note line")
		return "", isNextNote, nil

	case strings.Contains(strings.ToLower(buffLine), "your highlight "):
		//fmt.Println("Skipping boilerplate lane - your hightlight...")
		return "", isNextNote, nil
	default:
		return buffLine, isNextNote, nil
	}
}
