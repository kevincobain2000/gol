package pkg

import (
	"bufio"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/gookit/color"
	"github.com/ztrue/tracerr"
)

func IsReadableFile(filename string) (bool, error) {
	file, err := os.Open(filename)
	if err != nil {
		return false, tracerr.New(err.Error())
	}
	defer file.Close()

	// Check if the file is empty
	fileInfo, err := file.Stat()
	if err != nil {
		return false, tracerr.New(err.Error())
	}
	if fileInfo.Size() == 0 {
		return true, nil
	}
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil {
		return false, tracerr.New(err.Error())
	}
	// Check if the file is gzip compressed
	if IsGzip(buffer[:n]) {
		_, err := file.Seek(0, io.SeekStart) // Reset file pointer
		if err != nil {
			return false, tracerr.New(err.Error())
		}

		gzipReader, err := gzip.NewReader(file)
		if err != nil {
			return false, tracerr.New(err.Error())
		}
		defer gzipReader.Close()

		n, err = gzipReader.Read(buffer)
		if err != nil && !errors.Is(err, io.EOF) {
			return false, tracerr.New(err.Error())
		}

		return utf8.Valid(buffer[:n]), nil
	}

	return utf8.Valid(buffer[:n]), nil
}

// IsGzip checks if the given buffer starts with the gzip magic number
func IsGzip(buffer []byte) bool {
	return len(buffer) >= 2 && buffer[0] == 0x1f && buffer[1] == 0x8b
}

func FilesByPattern(pattern string) ([]string, error) {
	// Check if the pattern is a directory
	info, err := os.Stat(pattern)
	if err == nil && info.IsDir() {
		// List all files in the directory
		var files []string
		err := filepath.Walk(pattern, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return tracerr.New(err.Error())
			}
			if !info.IsDir() {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		return files, nil
	}

	// If pattern is not a directory, use Glob to match the pattern
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	return files, nil
}

// FileStats returns the number of lines and size of the file at the given path.
func FileStats(filePath string) (int, int64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	var linesCount int
	var fileSize int64

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		linesCount++
	}

	if err := scanner.Err(); err != nil {
		return 0, 0, err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return 0, 0, err
	}
	fileSize = fileInfo.Size()

	return linesCount, fileSize, nil
}

func GetFileInfos(pattern string, limit int) []FileInfo {
	filePaths, err := FilesByPattern(pattern)
	if err != nil {
		color.Danger.Println(err)
		return nil
	}
	if len(filePaths) == 0 {
		color.Danger.Println("no files found:", pattern)
		return nil
	}
	fileInfos := make([]FileInfo, 0)
	if len(filePaths) > limit {
		color.Warn.Printf("limiting to %d files\n", limit)
		filePaths = filePaths[:limit]
	}
	for _, filePath := range filePaths {
		isText, err := IsReadableFile(filePath)
		if err != nil {
			color.Danger.Println(err)
			return nil
		}
		if !isText {
			color.Warn.Println("file is not a text file:", filePath)
			continue
		}
		linesCount, fileSize, err := FileStats(filePath)
		if err != nil {
			color.Danger.Println(err)
			return nil
		}
		fileInfos = append(fileInfos, FileInfo{FilePath: filePath, LinesCount: linesCount, FileSize: fileSize, Type: "file"})
	}
	return fileInfos
}
