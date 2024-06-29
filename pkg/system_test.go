package pkg

import (
	"os"
	"reflect"
	"sync"
	"testing"
)

// TestGetHomedir tests the GetHomedir function
func TestGetHomedir(t *testing.T) {
	home := GetHomedir()
	if home == "" {
		t.Error("Expected a non-empty home directory")
	}
}
func TestIsInputFromPipe(t *testing.T) {
	// Create a pipe to simulate input from stdin
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer reader.Close()
	defer writer.Close()

	// Redirect stdin to the reader end of the pipe
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	os.Stdin = reader

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		_, err := writer.WriteString("test input\n")
		if err != nil {
			t.Errorf("Failed to write to pipe: %v", err)
		}
		writer.Close()
	}()

	// Wait for the goroutine to finish writing to the pipe
	wg.Wait()

	if !IsInputFromPipe() {
		t.Error("Expected IsInputFromPipe to return true")
	}
}

func TestUniqueFileInfos(t *testing.T) {
	tests := []struct {
		name     string
		input    []FileInfo
		expected []FileInfo
	}{
		{
			name: "no duplicates",
			input: []FileInfo{
				{FilePath: "path1", Type: "type1", Host: "host1"},
				{FilePath: "path2", Type: "type2", Host: "host2"},
			},
			expected: []FileInfo{
				{FilePath: "path1", Type: "type1", Host: "host1"},
				{FilePath: "path2", Type: "type2", Host: "host2"},
			},
		},
		{
			name: "with duplicates",
			input: []FileInfo{
				{FilePath: "path1", Type: "type1", Host: "host1"},
				{FilePath: "path1", Type: "type1", Host: "host1"},
				{FilePath: "path2", Type: "type2", Host: "host2"},
				{FilePath: "path2", Type: "type2", Host: "host2"},
			},
			expected: []FileInfo{
				{FilePath: "path1", Type: "type1", Host: "host1"},
				{FilePath: "path2", Type: "type2", Host: "host2"},
			},
		},
		{
			name: "all duplicates",
			input: []FileInfo{
				{FilePath: "path1", Type: "type1", Host: "host1"},
				{FilePath: "path1", Type: "type1", Host: "host1"},
				{FilePath: "path1", Type: "type1", Host: "host1"},
			},
			expected: []FileInfo{
				{FilePath: "path1", Type: "type1", Host: "host1"},
			},
		},
		{
			name:     "empty input",
			input:    []FileInfo{},
			expected: []FileInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := UniqueFileInfos(tt.input)
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("UniqueFileInfos(%v) = %v; want %v", tt.input, actual, tt.expected)
			}
		})
	}
}
