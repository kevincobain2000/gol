package pkg

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/acarl005/stripansi"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
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
		slog.Error("creating Docker client", "docker", err)
		return nil
	}

	// Get container logs
	options := container.LogsOptions{ShowStdout: true, ShowStderr: true}
	out, err := cli.ContainerLogs(context.Background(), containerID, options)
	if err != nil {
		slog.Error("getting container logs", containerID, err)
		return nil
	}

	// Check if tmpFile already exists in GlobalFilePaths for container ID previously by watcher
	var tmpFile *os.File
	for _, fileInfo := range GlobalFilePaths {
		if fileInfo.Host == containerID[:12] && fileInfo.Type == TypeDocker && strings.HasPrefix(fileInfo.FilePath, TmpContainerPath) {
			tmpFile, err = os.OpenFile(fileInfo.FilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				slog.Error("opening temp file", fileInfo.FilePath, err)
				return nil
			}
		}
	}
	if tmpFile == nil {
		tmpFile, err = os.Create(GetTmpFileNameForContainer())
		if err != nil {
			slog.Error("creating temp file", "tmp", err)
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
				slog.Error("truncating file", "scan", err)
			}
			if _, err := tmpFile.Seek(0, 0); err != nil {
				slog.Error("seeking file", "scan", err)
			}
			lineCount = 0
		}
		if _, err := tmpFile.WriteString(line + "\n"); err != nil {
			slog.Error("writing to file", "scan", err)
		}
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		slog.Error("reading container logs", containerID, err)
	}
	return tmpFile
}

func ContainerLogsFromFile(containerID string, query string, ignorePattern string, filePath string, page, pageSize int, reverse bool) (*ScanResult, error) {
	lines := []LineResult{}
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	re, err := regexp.Compile(query)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	var reIgnore *regexp.Regexp
	if ignorePattern != "" {
		reIgnore, err = regexp.Compile(ignorePattern)
		if err != nil {
			return nil, fmt.Errorf("invalid ignore regex pattern: %w", err)
		}
	}

	countCmd := []string{"sh", "-c", fmt.Sprintf("wc -l < %s", filePath)}
	countExecConfig := container.ExecOptions{
		Cmd:          countCmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	countExecIDResp, err := cli.ContainerExecCreate(context.Background(), containerID, countExecConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create exec instance for counting lines: %w", err)
	}

	countResp, err := cli.ContainerExecAttach(context.Background(), countExecIDResp.ID, container.ExecStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to exec instance for counting lines: %w", err)
	}
	defer countResp.Close()

	countScanner := bufio.NewScanner(countResp.Reader)
	countScanner.Scan()
	totalLines, err := strconv.Atoi(strings.TrimSpace(CleanString(countScanner.Text())))
	if err != nil {
		return nil, fmt.Errorf("failed to parse line count: %w", err)
	}

	startLine := (page - 1) * pageSize

	var cmd []string
	if reverse {
		cmd = []string{"sh", "-c", fmt.Sprintf("tac %s | tail -n +%d | head -n %d", filePath, startLine+1, pageSize)}
	} else {
		cmd = []string{"sh", "-c", fmt.Sprintf("tail -n +%d %s | head -n %d", startLine+1, filePath, pageSize)}
	}

	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	execIDResp, err := cli.ContainerExecCreate(context.Background(), containerID, execConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create exec instance: %w", err)
	}

	resp, err := cli.ContainerExecAttach(context.Background(), execIDResp.ID, container.ExecStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to exec instance: %w", err)
	}
	defer resp.Close()

	scanner := bufio.NewScanner(resp.Reader)
	lineNumber := startLine + 1
	for scanner.Scan() {
		lineContent := stripansi.Strip(scanner.Text())
		lineContent = CleanString(lineContent)
		if reIgnore != nil && reIgnore.MatchString(lineContent) {
			continue
		}
		if re.MatchString(lineContent) {
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
		return nil, fmt.Errorf("reading logs: %w", err)
	}

	if reverse {
		for i := range lines {
			lines[i].LineNumber = totalLines - (startLine + len(lines)) + i + 1
		}
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
	AppendDates(&lines)
	scanResult := &ScanResult{
		FilePath:     filePath,
		Host:         containerID,
		MatchPattern: query,
		Total:        totalLines,
		Lines:        lines,
	}

	return scanResult, nil
}

func GetContainerFileInfos(pattern string, limit int, containerID string) []FileInfo {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		slog.Error("Failed to create Docker client", "docker", err)
		return nil
	}

	execConfig := container.ExecOptions{
		Cmd:          []string{"sh", "-c", fmt.Sprintf("ls -1 %s", pattern)},
		AttachStdout: true,
		AttachStderr: true,
	}

	execIDResp, err := cli.ContainerExecCreate(context.Background(), containerID, execConfig)
	if err != nil {
		slog.Error("Failed to create exec instance", "container", err)
		return nil
	}

	resp, err := cli.ContainerExecAttach(context.Background(), execIDResp.ID, container.ExecStartOptions{})
	if err != nil {
		slog.Error("Failed to attach to exec instance", "container", err)
		return nil
	}
	defer resp.Close()

	scanner := bufio.NewScanner(resp.Reader)
	filePaths := []string{}
	for scanner.Scan() {
		filePaths = append(filePaths, CleanString(scanner.Text()))
	}
	if err := scanner.Err(); err != nil {
		slog.Error("reading exec output", "scanner", err)
		return nil
	}

	fileInfos := make([]FileInfo, 0)
	if len(filePaths) > limit {
		slog.Warn("Limiting to files", "docker", limit)
		filePaths = filePaths[:limit]
	}
	for _, filePath := range filePaths {
		linesCount, fileSize, err := getFileStatsFromContainer(cli, containerID, filePath)
		if err != nil {
			slog.Error("Failed to get file stats", filePath, err)
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
		return 0, 0, fmt.Errorf("reading wc output: %w", err)
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
		return 0, 0, fmt.Errorf("reading stat output: %w", err)
	}

	return linesCount, fileSize, nil
}
