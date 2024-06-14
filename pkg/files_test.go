package pkg

import (
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestIsReadableFile(t *testing.T) {
	// Create a temporary directory for the test files
	dir := t.TempDir()

	// Create a temporary UTF-8 encoded file
	utf8File := filepath.Join(dir, "utf8.txt")
	if err := os.WriteFile(utf8File, []byte("hello, world!"), 0600); err != nil {
		t.Fatalf("failed to create UTF-8 file: %v", err)
	}

	// Create a temporary gzip-compressed UTF-8 file
	gzipFile := filepath.Join(dir, "utf8.txt.gz")
	gzipContent := []byte("hello, gzipped world!")
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	if _, err := gzipWriter.Write(gzipContent); err != nil {
		t.Fatalf("failed to write gzip content: %v", err)
	}
	gzipWriter.Close()
	if err := os.WriteFile(gzipFile, buf.Bytes(), 0600); err != nil {
		t.Fatalf("failed to create gzip file: %v", err)
	}

	// Test cases
	tests := []struct {
		filename   string
		expectErr  bool
		expectBool bool
	}{
		{utf8File, false, true},
		{gzipFile, false, true},
		{"nonexistent.txt", true, false},
	}

	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			result, err := IsReadableFile(test.filename, false, nil)
			if (err != nil) != test.expectErr {
				t.Errorf("IsReadableFile(%q) error = %v, wantErr %v", test.filename, err, test.expectErr)
				return
			}
			if result != test.expectBool {
				t.Errorf("IsReadableFile(%q) = %v, want %v", test.filename, result, test.expectBool)
			}
		})
	}
}

func TestIsGzip(t *testing.T) {
	tests := []struct {
		name   string
		buffer []byte
		want   bool
	}{
		{"gzip header", []byte{0x1f, 0x8b}, true},
		{"not gzip header", []byte{0x00, 0x00}, false},
		{"empty buffer", []byte{}, false},
		{"partial gzip header", []byte{0x1f}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsGzip(test.buffer); got != test.want {
				t.Errorf("IsGzip(%v) = %v, want %v", test.buffer, got, test.want)
			}
		})
	}
}

func TestFilesByPattern(t *testing.T) {
	// Create a temporary directory for the test files
	dir := t.TempDir()

	// Create some test files
	files := []string{
		filepath.Join(dir, "file1.txt"),
		filepath.Join(dir, "file2.txt"),
		filepath.Join(dir, "file3.log"),
	}
	for _, file := range files {
		if err := os.WriteFile(file, []byte("test"), 0600); err != nil {
			t.Fatalf("failed to create file %q: %v", file, err)
		}
	}

	tests := []struct {
		pattern     string
		expectErr   bool
		expectFiles []string
	}{
		{dir, false, files},
		{filepath.Join(dir, "*.txt"), false, files[:2]},
		{filepath.Join(dir, "*.log"), false, files[2:3]},
		{filepath.Join(dir, "*.none"), false, []string{}},
		{"nonexistent", false, nil},
	}

	for _, test := range tests {
		t.Run(test.pattern, func(t *testing.T) {
			result, err := FilesByPattern(test.pattern, false, nil)
			if (err != nil) != test.expectErr {
				t.Errorf("FilesByPattern(%q) error = %v, wantErr %v", test.pattern, err, test.expectErr)
				return
			}
			if len(result) != len(test.expectFiles) {
				t.Errorf("FilesByPattern(%q) = %v, want %v", test.pattern, result, test.expectFiles)
			}
		})
	}
}
