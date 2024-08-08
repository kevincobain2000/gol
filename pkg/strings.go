package pkg

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/acarl005/stripansi"
	"github.com/gravwell/gravwell/v3/timegrinder"
	"github.com/mileusna/useragent"
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
	successKeywords := []string{"success", "SUCCESS", "succ", "SUCC", "Success"}
	infoKeywords := []string{"info", "inf", "INFO", "INF", "Info", "Inf"}
	errorKeywords := []string{"error", "err", "fail", "ERROR", "ERR", "FAIL", "Error", "Err", "Fail"}
	warnKeywords := []string{"warn", "warning", "alert", "wrn", "WARN", "WARNING", "ALERT", "Wrn", "Wrning", "Alert"}
	dangerKeywords := []string{"danger", "fatal", "severe", "critical", "DANGER", "FATAL", "SEVERE", "CRITICAL", "Danger", "Fatal", "Severe", "Critical"}
	debugKeywords := []string{"debug", "dbg", "DEBUG", "DBG", "Debug"}

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

func AppendGeneralInfo(lines *[]LineResult) {
	appendAgent(lines)
	appendDates(lines)
	appendLogLevel(lines)
}

func appendLogLevel(lines *[]LineResult) {
	logLines := []string{}
	for _, line := range *lines {
		line.Content = stripansi.Strip(line.Content)
		logLines = append(logLines, line.Content)
	}

	isConsistent, keywordPosition := ConsistentFormat(logLines)
	if isConsistent {
		for i, line := range *lines {
			(*lines)[i].Level = JudgeLogLevel(line.Content, keywordPosition)
		}
	}
}

func appendAgent(lines *[]LineResult) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()
	for i, line := range *lines {
		ua := useragent.Parse(line.Content)
		device := "server"

		switch {
		case ua.Desktop:
			device = "desktop"
		case ua.Mobile:
			device = "mobile"
		case ua.Tablet:
			device = "tablet"
		}

		(*lines)[i].Agent.Device = device
	}
}

func appendDates(lines *[]LineResult) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()
	for i, line := range *lines {
		date := searchDate(line.Content)
		(*lines)[i].Date = date
	}
}

var (
	tg   *timegrinder.TimeGrinder
	once sync.Once
)

func initTimeGrinder() error {
	cfg := timegrinder.Config{}
	var err error
	tg, err = timegrinder.NewTimeGrinder(cfg)
	if err != nil {
		return err
	}
	return nil
}

func searchDate(input string) string {
	var initErr error
	once.Do(func() {
		initErr = initTimeGrinder()
	})
	if initErr != nil {
		slog.Error("Error initializing", "timegrinder", initErr)
		return ""
	}
	ts, ok, err := tg.Extract([]byte(input))
	if err != nil {
		return ""
	}
	if !ok {
		return ""
	}
	return ts.String()
}
