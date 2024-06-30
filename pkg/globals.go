package pkg

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

var GlobalFilePaths []FileInfo
var GlobalPipeTmpFilePath string
var GlobalPathSSHConfig []SSHPathConfig
var GlobalSSHClients = make(map[string]*ssh.Client)

func WatchFilePaths(seconds int64, filePaths SliceFlags, sshPaths SliceFlags, dockerPaths SliceFlags, limit int) {
	interval := time.Duration(seconds) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		slog.Info("Checking for filepaths", "interval", interval)
		UpdateGlobalFilePaths(filePaths, sshPaths, dockerPaths, limit)
	}
}

func HandleStdinPipe() {
	tmpFile, err := os.Create(GetTmpFileNameForSTDIN())
	if err != nil {
		slog.Error("creating temp file", tmpFile.Name(), err)
		return
	}
	GlobalPipeTmpFilePath = tmpFile.Name()
	defer tmpFile.Close()
	go func(tmpFile *os.File) {
		err := PipeLinesToTmp(tmpFile)
		if err != nil {
			slog.Error("piping lines to temp file", tmpFile.Name(), err)
			return
		}
	}(tmpFile)
}

func UpdateGlobalFilePaths(filePaths SliceFlags, sshPaths SliceFlags, dockerPaths SliceFlags, limit int) {
	fileInfos := []FileInfo{}
	for _, pattern := range filePaths {
		fileInfo := GetFileInfos(pattern, limit, false, nil)
		fileInfos = append(fileInfo, fileInfos...)
	}
	for _, pattern := range sshPaths {
		sshFilePathConfig, err := StringToSSHPathConfig(pattern)
		if err != nil {
			slog.Error("parsing SSH path", pattern, err)
			break
		}
		sshConfig := SSHConfig{
			Host:           sshFilePathConfig.Host,
			Port:           sshFilePathConfig.Port,
			User:           sshFilePathConfig.User,
			Password:       sshFilePathConfig.Password,
			PrivateKeyPath: sshFilePathConfig.PrivateKeyPath,
		}
		GlobalPathSSHConfig = append(GlobalPathSSHConfig, *sshFilePathConfig)
		fileInfo := GetFileInfos(sshFilePathConfig.FilePath, limit, true, &sshConfig)
		fileInfos = append(fileInfo, fileInfos...)
	}

	for _, pattern := range dockerPaths {
		containers, err := ListDockerContainers()
		if err != nil {
			slog.Error("listing Docker containers", pattern, err)
			break
		}
		if pattern == "" || len(strings.Fields(pattern)) == 1 {
			for _, container := range containers {
				if pattern != "" && !strings.Contains(container.Names[0], pattern) {
					continue
				}
				tmpFile := ContainerStdoutToTmp(container.ID)
				if tmpFile == nil {
					slog.Error("creating temp file for container logs", "containerID", container.ID)
					continue
				}
				fileInfo := GetFileInfos(tmpFile.Name(), limit, false, nil)
				if len(fileInfo) > 0 {
					fileInfo[0].Host = container.ID[:12]
					fileInfo[0].Type = TypeDocker
					fileInfo[0].Name = container.Names[0][1:]
					fileInfos = append(fileInfo, fileInfos...)
				}
			}
		}
		if len(strings.Fields(pattern)) == 2 {
			dockerFilePathConfig, err := StringToDockerPathConfig(pattern)
			if err != nil {
				slog.Error("parsing Docker path", pattern, err)
				break
			}
			fileInfo := GetContainerFileInfos(dockerFilePathConfig.FilePath, limit, dockerFilePathConfig.ContainerID)
			fileInfos = append(fileInfo, fileInfos...)
		}
	}

	GlobalFilePaths = UniqueFileInfos(fileInfos)
}
