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

	file, scanner, err := w.initializeScanner()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	allLines, counts, err := w.collectMatchingLines(scanner)
	if err != nil {
		return nil, err
	}

	lines := w.paginateLines(allLines, page, pageSize, reverse)

	w.appendLogLevel(&lines)

	return &ScanResult{
		FilePath:     w.filePath,
		MatchPattern: w.matchPattern,
		Total:        counts,
		Lines:        lines,
	}, nil
}

func (w *Watcher) initializeScanner() (*os.File, *bufio.Scanner, error) {
	file, err := os.Open(w.filePath)
	if err != nil {
		return nil, nil, tracerr.New(err.Error())
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, nil, tracerr.New(err.Error())
	}
	if fileInfo.Size() == 0 {
		return file, bufio.NewScanner(file), nil
	}

	buffer := make([]byte, 2)
	if _, err := file.Read(buffer); err != nil {
		return nil, nil, tracerr.New(err.Error())
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		return nil, nil, tracerr.New(err.Error())
	}

	if IsGzip(buffer) {
		gzipReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, nil, tracerr.New(err.Error())
		}
		return file, bufio.NewScanner(gzipReader), nil
	}

	return file, bufio.NewScanner(file), nil
}

func (w *Watcher) collectMatchingLines(scanner *bufio.Scanner) ([]LineResult, int, error) {
	re, err := regexp.Compile(w.matchPattern)
	if err != nil {
		return nil, 0, tracerr.New(err.Error())
	}

	var allLines []LineResult
	lineNumber := 0
	counts := 0

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
		return nil, 0, tracerr.New(err.Error())
	}

	return allLines, counts, nil
}

func (w *Watcher) paginateLines(allLines []LineResult, page, pageSize int, reverse bool) []LineResult {
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

	if start < len(allLines) {
		return allLines[start:end]
	}

	return []LineResult{}
}

func (w *Watcher) appendLogLevel(lines *[]LineResult) {
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
