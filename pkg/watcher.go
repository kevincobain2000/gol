package pkg

import (
	"bufio"
	"compress/gzip"
	"os"
	"regexp"
	"sync"

	"github.com/acarl005/stripansi"
	"github.com/ztrue/tracerr"
)

type Watcher struct {
	filePath     string
	matchPattern string
	mutex        sync.Mutex
}

func NewWatcher(
	filePath string,
	matchPattern string,
) (*Watcher, error) {
	watcher := &Watcher{
		filePath:     filePath,
		matchPattern: matchPattern,
	}

	return watcher, nil
}

type LineResult struct {
	LineNumber int    `json:"line_number"`
	Content    string `json:"content"`
	Level      string `json:"level"`
}

type ScanResult struct {
	FilePath     string       `json:"file_path"`
	MatchPattern string       `json:"match_pattern"`
	Total        int          `json:"total"`
	Lines        []LineResult `json:"lines"`
}

func (w *Watcher) Scan(page, pageSize int, reverse bool) (*ScanResult, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	counts := 0
	var lines []LineResult

	file, err := os.Open(w.filePath)
	if err != nil {
		return nil, tracerr.New(err.Error())
	}
	defer file.Close()

	// Check if the file is empty
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, tracerr.New(err.Error())
	}
	if fileInfo.Size() == 0 {
		return &ScanResult{
			Total:        counts,
			Lines:        lines,
			FilePath:     w.filePath,
			MatchPattern: w.matchPattern,
		}, nil
	}

	var scanner *bufio.Scanner

	// Check if the file is gzip compressed
	buffer := make([]byte, 2)
	if _, err := file.Read(buffer); err != nil {
		return nil, tracerr.New(err.Error())
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		return nil, tracerr.New(err.Error())
	}

	if IsGzip(buffer) {
		gzipReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, tracerr.New(err.Error())
		}
		defer gzipReader.Close()
		scanner = bufio.NewScanner(gzipReader)
	} else {
		scanner = bufio.NewScanner(file)
	}

	re, err := regexp.Compile(w.matchPattern)
	if err != nil {
		return nil, tracerr.New(err.Error())
	}

	// Collect all matching lines
	var allLines []LineResult
	lineNumber := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineNumber++
		if re.MatchString(line) {
			allLines = append(allLines, LineResult{
				LineNumber: lineNumber,
				Content:    line,
			})
			counts++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, tracerr.New(err.Error())
	}

	// Determine the start and end indices based on the page and pageSize
	var start, end int
	if reverse {
		start = len(allLines) - (page * pageSize)
		if start < 0 {
			start = 0
		}
		end = start + pageSize
		if end > len(allLines) {
			end = len(allLines)
		}
	} else {
		start = (page - 1) * pageSize
		end = start + pageSize
		if end > len(allLines) {
			end = len(allLines)
		}
	}

	// Slice the lines to get the required page
	if start < len(allLines) {
		lines = allLines[start:end]
	}

	// append log level
	logLines := []string{}
	for index := range lines {
		lines[index].Content = stripansi.Strip(lines[index].Content)
		logLines = append(logLines, lines[index].Content)
	}

	isConsistent, keywordPosition := ConsistentFormat(logLines)
	if isConsistent {
		for index := range lines {
			lines[index].Level = JudgeLogLevel(lines[index].Content, keywordPosition)
		}
	}

	return &ScanResult{
		FilePath:     w.filePath,
		MatchPattern: w.matchPattern,
		Total:        counts,
		Lines:        lines,
	}, nil
}
