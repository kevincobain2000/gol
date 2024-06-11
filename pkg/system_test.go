package pkg

import (
	"os"
	"testing"
)

// TestGetHomedir tests the GetHomedir function
func TestGetHomedir(t *testing.T) {
	home := GetHomedir()
	if home == "" {
		t.Error("Expected a non-empty home directory")
	}
}

// TestIsInputFromPipe tests the IsInputFromPipe function
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

	go func() {
		writer.WriteString("test input\n")
		writer.Close()
	}()

	if !IsInputFromPipe() {
		t.Error("Expected IsInputFromPipe to return true")
	}
}
