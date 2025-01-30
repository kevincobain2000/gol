package pkg

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/acarl005/stripansi"
	"golang.org/x/crypto/ssh"
)

type Watcher struct {
	filePath      string
	matchPattern  string
	ignorePattern string
	mutex         sync.Mutex
	sshConfig     *ssh.ClientConfig
	sshHost       string
	sshPort       string
	isRemote      bool
}

func NewWatcher(
	filePath string,
	matchPattern string,
	ignorePattern string,
	isRemote bool,
	sshHost string,
	sshPort string,
	sshUser string,
	sshPassword string,
	sshPrivateKeyPath string,
) (*Watcher, error) {
	var authMethod ssh.AuthMethod
	if sshPrivateKeyPath != "" {
		key, err := os.ReadFile(sshPrivateKeyPath)
		if err != nil {
			return nil, err
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, err
		}
		authMethod = ssh.PublicKeys(signer)
	} else {
		authMethod = ssh.Password(sshPassword)
	}

	watcher := &Watcher{
		filePath:      filePath,
		matchPattern:  matchPattern,
		ignorePattern: ignorePattern,
		isRemote:      isRemote,
		sshHost:       sshHost,
		sshPort:       sshPort,
		sshConfig: &ssh.ClientConfig{
			User: sshUser,
			Auth: []ssh.AuthMethod{
				authMethod,
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(), // nolint:gosec
		},
	}

	return watcher, nil
}

type LineResult struct {
	LineNumber int    `json:"line_number"`
	Content    string `json:"content"`
	Level      string `json:"level"`
	Date       string `json:"date"`
	Agent      struct {
		Device string `json:"device"`
	} `json:"agent"`
}

type ScanResult struct {
	FilePath     string       `json:"file_path"`
	Host         string       `json:"host"`
	Type         string       `json:"type"`
	MatchPattern string       `json:"match_pattern"`
	Total        int          `json:"total"`
	Lines        []LineResult `json:"lines"`
}

func (w *Watcher) Scan(page, pageSize int, reverse bool) (*ScanResult, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	file, scanner, err := w.initializeScanner()
	scanner.Buffer(make([]byte, 0, 64*1024), 5*1024*1024) // 5MB buffer
	if err != nil {
		return nil, err
	}
	if file != nil {
		defer file.Close()
	}

	allLines, counts, err := w.collectMatchingLines(scanner)
	if err != nil {
		return nil, err
	}

	lines := w.paginateLines(allLines, page, pageSize, reverse)

	AppendGeneralInfo(&lines)
	return &ScanResult{
		FilePath:     w.filePath,
		Host:         w.sshHost,
		MatchPattern: w.matchPattern,
		Total:        counts,
		Lines:        lines,
	}, nil
}

func (w *Watcher) initializeScanner() (*os.File, *bufio.Scanner, error) {
	if w.isRemote {
		return w.initializeRemoteScanner()
	}

	file, err := os.Open(w.filePath)
	if err != nil {
		return nil, nil, err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, nil, err
	}
	if fileInfo.Size() == 0 {
		return file, bufio.NewScanner(file), nil
	}

	buffer := make([]byte, 2)
	if _, err := file.Read(buffer); err != nil {
		return nil, nil, err
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		return nil, nil, err
	}

	if IsGzip(buffer) {
		gzipReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, nil, err
		}
		return file, bufio.NewScanner(gzipReader), nil
	}

	return file, bufio.NewScanner(file), nil
}

func (w *Watcher) initializeRemoteScanner() (*os.File, *bufio.Scanner, error) {
	sshConfig := SSHConfig{
		Host: w.sshHost,
		Port: w.sshPort,
	}
	session, err := NewSession(&sshConfig)
	if err != nil {
		return nil, nil, err
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run(fmt.Sprintf("cat %s", w.filePath)); err != nil {
		if err.Error() != ErrorMsgSessionAlreadyStarted {
			return nil, nil, err
		}
	}

	scanner := bufio.NewScanner(strings.NewReader(b.String()))

	return nil, scanner, nil
}

func (w *Watcher) collectMatchingLines(scanner *bufio.Scanner) ([]LineResult, int, error) {
	re, err := regexp.Compile(w.matchPattern)
	if err != nil {
		return nil, 0, err
	}

	var reIgnore *regexp.Regexp
	if w.ignorePattern != "" {
		reIgnore, err = regexp.Compile(w.ignorePattern)
		if err != nil {
			return nil, 0, err
		}
	}

	var allLines []LineResult
	lineNumber := 0
	counts := 0

	for scanner.Scan() {
		line := scanner.Text()
		line = stripansi.Strip(line)
		lineNumber++
		if reIgnore != nil && reIgnore.MatchString(line) {
			continue
		}
		if re.MatchString(line) {
			allLines = append(allLines, LineResult{
				LineNumber: lineNumber,
				Content:    line,
			})
			counts++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, 0, err
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
