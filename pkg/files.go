package pkg

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/gookit/color"
	"github.com/ztrue/tracerr"
	"golang.org/x/crypto/ssh"
)

// IsReadableFile checks if the file is readable and optionally checks for valid UTF-8 encoded content
func IsReadableFile(filename string, isRemote bool, sshConfig *SSHConfig, checkUTF8 bool) (bool, error) {
	var file *os.File
	var err error

	if isRemote {
		file, err = sshOpenFile(filename, sshConfig)
	} else {
		file, err = os.Open(filename)
	}
	if err != nil {
		return false, tracerr.Wrap(err)
	}
	defer file.Close()

	// Check if the file is empty
	fileInfo, err := file.Stat()
	if err != nil {
		return false, tracerr.Wrap(err)
	}
	if fileInfo.Size() == 0 {
		return true, nil
	}

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil {
		return false, tracerr.Wrap(err)
	}

	// Check if the file is gzip compressed
	if IsGzip(buffer[:n]) {
		_, err = file.Seek(0, io.SeekStart) // Reset file pointer
		if err != nil {
			return false, tracerr.Wrap(err)
		}

		gzipReader, err := gzip.NewReader(file)
		if err != nil {
			return false, tracerr.Wrap(err)
		}
		defer gzipReader.Close()

		n, err = gzipReader.Read(buffer)
		if err != nil && !errors.Is(err, io.EOF) {
			return false, tracerr.Wrap(err)
		}

		if checkUTF8 {
			return utf8.Valid(buffer[:n]), nil
		}
		return true, nil
	}

	if checkUTF8 {
		return utf8.Valid(buffer[:n]), nil
	}
	return true, nil
}

// IsGzip checks if the given buffer starts with the gzip magic number
func IsGzip(buffer []byte) bool {
	return len(buffer) >= 2 && buffer[0] == 0x1f && buffer[1] == 0x8b
}

func FilesByPattern(pattern string, isRemote bool, sshConfig *SSHConfig) ([]string, error) {
	if isRemote {
		return sshFilesByPattern(pattern, sshConfig)
	}

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
func FileStats(filePath string, isRemote bool, sshConfig *SSHConfig) (int, int64, error) {
	var file *os.File
	var err error

	if isRemote {
		file, err = sshOpenFile(filePath, sshConfig)
	} else {
		file, err = os.Open(filePath)
	}
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

func GetFileInfos(pattern string, limit int, isRemote bool, sshConfig *SSHConfig) []FileInfo {
	filePaths, err := FilesByPattern(pattern, isRemote, sshConfig)
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
		isText, err := IsReadableFile(filePath, isRemote, sshConfig, false)
		if err != nil {
			color.Danger.Println(err)
			return nil
		}
		if !isText {
			color.Warn.Println("file is not a text file:", filePath)
			continue
		}
		linesCount, fileSize, err := FileStats(filePath, isRemote, sshConfig)
		if err != nil {
			color.Danger.Println(err)
			return nil
		}
		t := TypeFile
		h := ""
		if isRemote {
			t = TypeSSH
			h = sshConfig.Host
		}
		if filePath == GlobalPipeTmpFilePath {
			t = TypeStdin
		}
		fileInfos = append(fileInfos, FileInfo{FilePath: filePath, LinesCount: linesCount, FileSize: fileSize, Type: t, Host: h})
	}
	return fileInfos
}

// SSHConfig holds the SSH connection parameters
type SSHConfig struct {
	Host           string
	Port           string
	User           string
	Password       string
	PrivateKeyPath string
}

type SSHPathConfig struct {
	Host           string
	Port           string
	User           string
	Password       string
	PrivateKeyPath string
	FilePath       string
}

type DockerPathConfig struct {
	ContainerID string
	FilePath    string
}

// s is an input of the form "container_id /path/to/file"
func StringToDockerPathConfig(s string) (*DockerPathConfig, error) {
	// Split the input string into parts
	parts := strings.Fields(s)

	// There should be 2 parts: "container_id" and "/path/to/file"
	if len(parts) < 2 {
		return nil, fmt.Errorf("input string does not have the correct format")
	}
	return &DockerPathConfig{
		ContainerID: parts[0],
		FilePath:    parts[1],
	}, nil
}

// s is an input of the form "user@host[:port] [password=/path/to/password] [private_key=/path/to/key] /path/to/file"
func StringToSSHPathConfig(s string) (*SSHPathConfig, error) {
	config := &SSHPathConfig{}

	// Split the input string into parts
	parts := strings.Fields(s)

	// There should be at least 2 parts: "user@host[:port]" and "/path/to/file"
	if len(parts) < 2 {
		return nil, fmt.Errorf("input string does not have the correct format")
	}

	// Extract user@host[:port]
	userHostPort := strings.Split(parts[0], "@")
	if len(userHostPort) != 2 {
		return nil, fmt.Errorf("user@host[:port] part does not have the correct format")
	}

	userHost := strings.Split(userHostPort[1], ":")
	config.User = userHostPort[0]
	config.Host = userHost[0]

	// Set the default port if not specified
	if len(userHost) == 2 {
		config.Port = userHost[1]
	} else {
		config.Port = "22" // Default SSH port
	}

	// Default private key path
	config.PrivateKeyPath = fmt.Sprintf("%s/.ssh/id_rsa", os.Getenv("HOME"))

	// Extract optional parts and file path
	for _, part := range parts[1:] {
		// nolint: gocritic
		if strings.HasPrefix(part, "password=") {
			config.Password = strings.TrimPrefix(part, "password=")
		} else if strings.HasPrefix(part, "private_key=") {
			config.PrivateKeyPath = strings.TrimPrefix(part, "private_key=")
		} else {
			config.FilePath = part
		}
	}

	if config.FilePath == "" {
		return nil, fmt.Errorf("file path is missing")
	}

	return config, nil
}

func sshConnect(config *SSHConfig) (*ssh.Client, error) {
	var auth []ssh.AuthMethod

	if config.Password != "" {
		auth = append(auth, ssh.Password(config.Password))
	}
	if config.PrivateKeyPath != "" {
		key, err := os.ReadFile(config.PrivateKeyPath)
		if err != nil {
			return nil, err
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, err
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}

	clientConfig := &ssh.ClientConfig{
		User:            config.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // nolint:gosec
	}

	client, err := ssh.Dial("tcp", config.Host+":"+config.Port, clientConfig)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func sshOpenFile(filename string, config *SSHConfig) (*os.File, error) {
	client, err := sshConnect(config)
	if err != nil {
		return nil, err
	}
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	tmpFile, err := os.Create(GetTmpFileNameForSTDIN())
	if err != nil {
		return nil, err
	}

	// Execute the cat command to read the file
	var stdout bytes.Buffer
	session.Stdout = &stdout
	if err := session.Run("cat " + filename); err != nil {
		return nil, err
	}

	// Write the remote file content to the temporary file
	if _, err := tmpFile.Write(stdout.Bytes()); err != nil {
		return nil, err
	}

	// Seek to the beginning of the temporary file
	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	return tmpFile, nil
}

func sshFilesByPattern(pattern string, config *SSHConfig) ([]string, error) {
	client, err := sshConnect(config)
	if err != nil {
		return nil, err
	}
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	var buf bytes.Buffer
	session.Stdout = &buf

	// Execute the ls command to list files matching the pattern
	if err := session.Run("ls " + pattern); err != nil {
		return nil, err
	}

	filePaths := buf.String()
	return strings.Split(strings.TrimSpace(filePaths), "\n"), nil
}
