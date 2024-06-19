package pkg

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/acarl005/stripansi"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/gookit/color"
)

func ListDockerContainers() ([]types.Container, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	// Get the list of containers
	return cli.ContainerList(context.Background(), container.ListOptions{})
}

func ContainerStdoutToTmp(containerID string) *os.File {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Printf("Error creating Docker client: %v", err)
		return nil
	}

	// Get container logs
	options := container.LogsOptions{ShowStdout: true, ShowStderr: true}
	out, err := cli.ContainerLogs(context.Background(), containerID, options)
	if err != nil {
		fmt.Printf("Error getting container logs: %v", err)
		return nil
	}
	// defer out.Close() //donot close the stream
	// check if tmpFile already exists in GlobalFilePaths for container ID previously by watcher
	var tmpFile *os.File
	for _, fileInfo := range GlobalFilePaths {
		if fileInfo.Host == containerID[:12] && fileInfo.Type == TypeDocker && strings.HasPrefix(fileInfo.FilePath, TmpContainerPath) {
			tmpFile, err = os.OpenFile(fileInfo.FilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				color.Danger.Println("error opening temp file: ", err)
				return nil
			}
		}
	}
	if tmpFile == nil {
		tmpFile, err = os.Create(GetTmpFileNameForContainer())
		if err != nil {
			color.Danger.Println("error creating temp file: ", err)
			return nil
		}
	}
	scanner := bufio.NewScanner(out)
	lineCount := 0
	for scanner.Scan() {
		line := scanner.Text()
		line = stripansi.Strip(line)
		if lineCount >= 10000 {
			if err := tmpFile.Truncate(0); err != nil {
				color.Danger.Println("error truncating file: ", err)
			}
			if _, err := tmpFile.Seek(0, 0); err != nil {
				color.Danger.Println("error seeking file: ", err)
			}
			lineCount = 0
		}
		if _, err := tmpFile.WriteString(line + "\n"); err != nil {
			color.Danger.Println("error writing to file: ", err)
		}
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading container logs: %v", err)
	}
	return tmpFile
}

// ContainerLogsFromFile retrieves logs from a file within a container, processes them, and returns a ScanResult
func ContainerLogsFromFile(containerID string, query string, filePath string, page, pageSize int, reverse bool) (*ScanResult, error) {
	lines := []LineResult{}
	// Create a new Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	// Compile the regex pattern
	re, err := regexp.Compile(query)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Execute a command to count the total number of lines in the file
	countCmd := []string{"sh", "-c", fmt.Sprintf("wc -l < %s", filePath)}
	countExecConfig := container.ExecOptions{
		Cmd:          countCmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	// Create exec instance for counting lines
	countExecIDResp, err := cli.ContainerExecCreate(context.Background(), containerID, countExecConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create exec instance for counting lines: %w", err)
	}

	// Attach to the exec instance for counting lines
	countResp, err := cli.ContainerExecAttach(context.Background(), countExecIDResp.ID, container.ExecStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to exec instance for counting lines: %w", err)
	}
	defer countResp.Close()

	// Read the output of the line count command
	countScanner := bufio.NewScanner(countResp.Reader)
	countScanner.Scan()
	totalLines, err := strconv.Atoi(strings.TrimSpace(CleanString(countScanner.Text())))
	if err != nil {
		return nil, fmt.Errorf("failed to parse line count: %w", err)
	}

	// Calculate the lines to fetch
	startLine := (page - 1) * pageSize

	// Build the command to fetch the required lines
	var cmd []string
	if reverse {
		cmd = []string{"sh", "-c", fmt.Sprintf("tac %s | tail -n +%d | head -n %d", filePath, startLine+1, pageSize)}
	} else {
		cmd = []string{"sh", "-c", fmt.Sprintf("tail -n +%d %s | head -n %d", startLine+1, filePath, pageSize)}
	}

	// Execute the command inside the container
	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	// Create exec instance for fetching logs
	execIDResp, err := cli.ContainerExecCreate(context.Background(), containerID, execConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create exec instance: %w", err)
	}

	// Attach to the exec instance for fetching logs
	resp, err := cli.ContainerExecAttach(context.Background(), execIDResp.ID, container.ExecStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to exec instance: %w", err)
	}
	defer resp.Close()

	// Read logs line by line
	scanner := bufio.NewScanner(resp.Reader)
	lineNumber := startLine + 1
	for scanner.Scan() {
		lineContent := stripansi.Strip(scanner.Text())
		lineContent = CleanString(lineContent)
		if re.MatchString(lineContent) {
			// Here, you might want to include logic to determine the 'Level' in the log line
			lineResult := LineResult{
				LineNumber: lineNumber,
				Content:    lineContent,
				Level:      "",
			}
			lines = append(lines, lineResult)
		}
		lineNumber++
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading logs: %w", err)
	}

	// If reverse flag is set, adjust the line numbers accordingly
	if reverse {
		for i := range lines {
			lines[i].LineNumber = totalLines - (startLine + len(lines)) + i + 1
		}
		// line numbers are correct but the order of lines is reversed
		shifted := make([]LineResult, len(lines))
		for i := range lines {
			newIndex := len(lines) - 1 - i
			shifted[newIndex] = LineResult{
				LineNumber: lines[len(lines)-1-i].LineNumber,
				Content:    lines[i].Content,
				Level:      lines[len(lines)-1-i].Level,
			}
		}
		lines = shifted
	}

	AppendLogLevel(&lines)
	scanResult := &ScanResult{
		FilePath:     filePath,
		Host:         containerID, // Assuming containerID represents the host
		MatchPattern: query,
		Total:        totalLines,
		Lines:        lines,
	}

	return scanResult, nil
}

func GetContainerFileInfos(pattern string, limit int, containerID string) []FileInfo {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Printf("Failed to create Docker client: %v\n", err)
		return nil
	}

	execConfig := container.ExecOptions{
		Cmd:          []string{"sh", "-c", fmt.Sprintf("ls -1 %s", pattern)},
		AttachStdout: true,
		AttachStderr: true,
	}

	execIDResp, err := cli.ContainerExecCreate(context.Background(), containerID, execConfig)
	if err != nil {
		fmt.Printf("Failed to create exec instance: %v\n", err)
		return nil
	}

	resp, err := cli.ContainerExecAttach(context.Background(), execIDResp.ID, container.ExecStartOptions{})
	if err != nil {
		fmt.Printf("Failed to attach to exec instance: %v\n", err)
		return nil
	}
	defer resp.Close()

	scanner := bufio.NewScanner(resp.Reader)
	filePaths := []string{}
	for scanner.Scan() {
		filePaths = append(filePaths, CleanString(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading exec output: %v\n", err)
		return nil
	}

	fileInfos := make([]FileInfo, 0)
	if len(filePaths) > limit {
		color.Warn.Printf("limiting to %d files\n", limit)
		filePaths = filePaths[:limit]
	}
	for _, filePath := range filePaths {
		linesCount, fileSize, err := getFileStatsFromContainer(cli, containerID, filePath)
		if err != nil {
			fmt.Printf("Failed to get file stats: %v\n", err)
			continue
		}

		fileInfos = append(fileInfos, FileInfo{
			FilePath:   filePath,
			LinesCount: linesCount,
			FileSize:   fileSize,
			Type:       TypeDocker,
			Host:       containerID[:12],
		})
	}

	return fileInfos
}

func getFileStatsFromContainer(cli *client.Client, containerID string, filePath string) (int, int64, error) {
	execConfig := container.ExecOptions{
		Cmd:          []string{"wc", "-l", filePath},
		AttachStdout: true,
		AttachStderr: true,
	}

	execIDResp, err := cli.ContainerExecCreate(context.Background(), containerID, execConfig)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create exec instance for wc: %w", err)
	}

	resp, err := cli.ContainerExecAttach(context.Background(), execIDResp.ID, container.ExecStartOptions{})
	if err != nil {
		return 0, 0, fmt.Errorf("failed to attach to exec instance for wc: %w", err)
	}
	defer resp.Close()

	scanner := bufio.NewScanner(resp.Reader)
	var linesCount int
	if scanner.Scan() {
		fmt.Sscanf(CleanString(scanner.Text()), "%d", &linesCount) //nolint: errcheck
	}
	if err := scanner.Err(); err != nil {
		return 0, 0, fmt.Errorf("error reading wc output: %w", err)
	}

	execConfig = container.ExecOptions{
		Cmd:          []string{"stat", "-c", "%s", filePath},
		AttachStdout: true,
		AttachStderr: true,
	}

	execIDResp, err = cli.ContainerExecCreate(context.Background(), containerID, execConfig)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create exec instance for stat: %w", err)
	}

	resp, err = cli.ContainerExecAttach(context.Background(), execIDResp.ID, container.ExecStartOptions{})
	if err != nil {
		return 0, 0, fmt.Errorf("failed to attach to exec instance for stat: %w", err)
	}
	defer resp.Close()

	scanner = bufio.NewScanner(resp.Reader)
	var fileSize int64
	if scanner.Scan() {
		fmt.Sscanf(CleanString(scanner.Text()), "%d", &fileSize) //nolint: errcheck
	}
	if err := scanner.Err(); err != nil {
		return 0, 0, fmt.Errorf("error reading stat output: %w", err)
	}

	return linesCount, fileSize, nil
}
