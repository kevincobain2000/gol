package pkg

import (
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewWatcher tests the NewWatcher function
func TestNewWatcher(t *testing.T) {
	watcher, err := NewWatcher("testfile.log", "ERROR")
	assert.NoError(t, err)
	assert.NotNil(t, watcher)
	assert.Equal(t, "testfile.log", watcher.filePath)
	assert.Equal(t, "ERROR", watcher.matchPattern)
}

// TestWatcher_Scan tests the Scan method of the Watcher struct
func TestWatcher_Scan(t *testing.T) {
	dir := t.TempDir()

	// Create a temporary log file
	logFile := filepath.Join(dir, "test.log")
	content := `INFO Starting service
ERROR An error occurred
INFO Service running
ERROR Another error occurred`
	err := os.WriteFile(logFile, []byte(content), 0644)
	assert.NoError(t, err)

	// Create the Watcher
	watcher, err := NewWatcher(logFile, "ERROR")
	assert.NoError(t, err)

	// Run the Scan method
	result, err := watcher.Scan(1, 10, false)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, result.Total)
	assert.Len(t, result.Lines, 2)
	assert.Equal(t, 2, result.Lines[0].LineNumber)
	assert.Equal(t, "ERROR An error occurred", result.Lines[0].Content)
	assert.Equal(t, 4, result.Lines[1].LineNumber)
	assert.Equal(t, "ERROR Another error occurred", result.Lines[1].Content)
}

// TestWatcher_InitializeScanner tests the initializeScanner method of the Watcher struct
func TestWatcher_InitializeScanner(t *testing.T) {
	dir := t.TempDir()

	// Create a temporary gzip log file
	logFile := filepath.Join(dir, "test.log.gz")
	var buf strings.Builder
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write([]byte("INFO Starting service\nERROR An error occurred\n"))
	assert.NoError(t, err)
	assert.NoError(t, gz.Close())
	err = os.WriteFile(logFile, []byte(buf.String()), 0644)
	assert.NoError(t, err)

	// Create the Watcher
	watcher, err := NewWatcher(logFile, "ERROR")
	assert.NoError(t, err)

	// Initialize the scanner
	file, scanner, err := watcher.initializeScanner()
	assert.NoError(t, err)
	assert.NotNil(t, file)
	assert.NotNil(t, scanner)

	// Read the lines
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	assert.NoError(t, scanner.Err())
	assert.Equal(t, 2, len(lines))
	assert.Equal(t, "INFO Starting service", lines[0])
	assert.Equal(t, "ERROR An error occurred", lines[1])
}

// TestWatcher_CollectMatchingLines tests the collectMatchingLines method of the Watcher struct
func TestWatcher_CollectMatchingLines(t *testing.T) {
	// Create a temporary log file
	dir := t.TempDir()
	logFile := filepath.Join(dir, "test.log")
	content := `INFO Starting service
ERROR An error occurred
INFO Service running
ERROR Another error occurred`
	err := os.WriteFile(logFile, []byte(content), 0644)
	assert.NoError(t, err)

	// Create the Watcher
	watcher, err := NewWatcher(logFile, "ERROR")
	assert.NoError(t, err)

	// Initialize the scanner
	file, scanner, err := watcher.initializeScanner()
	assert.NoError(t, err)
	defer file.Close()

	// Collect matching lines
	lines, counts, err := watcher.collectMatchingLines(scanner)
	assert.NoError(t, err)
	assert.Equal(t, 2, counts)
	assert.Len(t, lines, 2)
	assert.Equal(t, 2, lines[0].LineNumber)
	assert.Equal(t, "ERROR An error occurred", lines[0].Content)
	assert.Equal(t, 4, lines[1].LineNumber)
	assert.Equal(t, "ERROR Another error occurred", lines[1].Content)
}
