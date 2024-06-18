package pkg

import (
	"strconv"
	"strings"
	"unicode"
)

func F64NumberToK(num *float64) string {
	if num == nil {
		return "0"
	}

	if *num < 1000 {
		return strconv.FormatFloat(*num, 'f', -1, 64)
	}

	if *num < 1000000 {
		return strconv.FormatFloat(*num/1000, 'f', 1, 64) + "k"
	}

	return strconv.FormatFloat(*num/1000000, 'f', 1, 64) + "m"
}

func StringInSlice(s string, ss []string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}
func FilePathInGlobalFilePaths(filePath string) bool {
	for _, fileInfo := range GlobalFilePaths {
		if fileInfo.FilePath == filePath {
			return true
		}
	}
	return false
}

// CleanString removes non-printable characters from a string
func CleanString(input string) string {
	cleaned := make([]rune, 0, len(input))
	for _, r := range input {
		if unicode.IsPrint(r) {
			cleaned = append(cleaned, r)
		}
	}
	// remove first %
	if len(cleaned) > 0 && cleaned[0] == '%' {
		cleaned = cleaned[1:]
	}
	return string(cleaned)
}

// JudgeLogLevel returns the log level based on the content of the log line if the format is consistent
func JudgeLogLevel(line string, keywordPosition int) string {
	line = strings.ToLower(line) // Convert the line to lowercase for easier comparison

	// Keywords for different log levels
	successKeywords := []string{"success"}
	infoKeywords := []string{"info"}
	errorKeywords := []string{"error"}
	warnKeywords := []string{"warn", "warning"}
	dangerKeywords := []string{"danger", "fatal", "severe", "critical"}
	debugKeywords := []string{"debug"}

	// Helper function to check if a keyword is at a specific position
	isKeywordAtPosition := func(line, keyword string, position int) bool {
		return strings.Index(line, keyword) == position
	}

	// Check for keywords at the specified position
	for _, keyword := range successKeywords {
		if isKeywordAtPosition(line, keyword, keywordPosition) {
			return "success"
		}
	}
	for _, keyword := range infoKeywords {
		if isKeywordAtPosition(line, keyword, keywordPosition) {
			return "info"
		}
	}

	for _, keyword := range errorKeywords {
		if isKeywordAtPosition(line, keyword, keywordPosition) {
			return "error"
		}
	}

	for _, keyword := range warnKeywords {
		if isKeywordAtPosition(line, keyword, keywordPosition) {
			return "warn"
		}
	}

	for _, keyword := range dangerKeywords {
		if isKeywordAtPosition(line, keyword, keywordPosition) {
			return "danger"
		}
	}

	for _, keyword := range debugKeywords {
		if isKeywordAtPosition(line, keyword, keywordPosition) {
			return "debug"
		}
	}

	// Default log level if no keywords match
	return ""
}

// ConsistentFormat checks if all log lines have log levels at the same position
func ConsistentFormat(logLines []string) (bool, int) {
	if len(logLines) == 0 {
		return false, -1
	}

	positions := []int{}

	for _, line := range logLines {
		line = strings.ToLower(line)
		words := strings.FieldsFunc(line, func(c rune) bool {
			return !unicode.IsLetter(c)
		})

		if len(words) == 0 {
			continue
		}

		firstWord := words[0]

		position := strings.Index(line, firstWord)
		positions = append(positions, position)
	}

	consistentPosition := positions[0]
	for _, pos := range positions {
		if pos != consistentPosition {
			return false, -1
		}
	}

	return true, consistentPosition
}
