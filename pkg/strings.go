package pkg

import (
	"strings"
	"unicode"
)

func Truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func UniqueStrings(s []string) []string {
	seen := make(map[string]struct{}, len(s))
	j := 0
	for _, v := range s {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		s[j] = v
		j++
	}
	return s[:j]
}

func StringInSlice(s string, ss []string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
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
