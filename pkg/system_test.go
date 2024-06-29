package pkg

import (
	"os"
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

// func TestUniqueFileInfos(t *testing.T)
// {
// 	GlobalFilePaths = []FileInfo{
// 		{FilePath: "/tmp/file1"},
// 		{FilePath: "/tmp/file2"},
// 		{FilePath: "/tmp/file1"},
// 	}

// 	uniqueFileInfos := UniqueFileInfos(GlobalFilePaths)
// 	if len(uniqueFileInfos) != 2 {
// 		t.Errorf("Expected 2 unique file infos, got %d", len(uniqueFileInfos))
// 	}
// }
