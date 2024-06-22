package main

import (
	"embed"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/kevincobain2000/gol/pkg"
)

//go:embed all:frontend/dist/*
var publicDir embed.FS

type sliceFlags []string

func (i *sliceFlags) String() string {
	return "my string representation"
}
func (i *sliceFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type Flags struct {
	host        string
	port        int64
	cors        int64
	every       int64
	limit       int
	baseURL     string
	filePaths   sliceFlags
	sshPaths    sliceFlags
	dockerPaths sliceFlags
	access      bool
	open        bool
	version     bool
}

var f Flags

var version = "dev"

func main() {
	pkg.SetupLoggingStdout()
	flags()
	wantsVersion()

	if pkg.IsInputFromPipe() {
		tmpFile, err := os.Create(pkg.GetTmpFileNameForSTDIN())
		if err != nil {
			slog.Error("error creating temp file", err)
			return
		}
		pkg.GlobalPipeTmpFilePath = tmpFile.Name()
		defer tmpFile.Close()
		go func(tmpFile *os.File) {
			err := pkg.PipeLinesToTmp(tmpFile)
			if err != nil {
				slog.Error("error piping lines to temp file", err)
				return
			}
		}(tmpFile)
	}
	defaultFilePaths()

	go watchFilePaths(f.every)
	slog.Info("Flags", "host", f.host, "port", f.port, "baseURL", f.baseURL, "open", f.open, "cors", f.cors, "access", f.access)

	if f.open {
		pkg.OpenBrowser(fmt.Sprintf("http://%s:%d%s", f.host, f.port, f.baseURL))
	}
	defer cleanup()
	pkg.HandleCltrC(cleanup)

	err := pkg.NewEcho(func(o *pkg.EchoOptions) error {
		o.Host = f.host
		o.Port = f.port
		o.Cors = f.cors
		o.Access = f.access
		o.BaseURL = f.baseURL
		o.PublicDir = &publicDir
		return nil
	})
	if err != nil {
		slog.Error("error starting echo", err)
		return
	}
}

func defaultFilePaths() {
	if len(os.Args) > 1 {
		filePaths := sliceFlags{}
		for _, arg := range os.Args[1:] {
			if strings.HasPrefix(arg, "-") {
				filePaths = []string{}
				break
			}
			filePaths = append(filePaths, arg)
		}
		if len(filePaths) > 0 {
			f.filePaths = filePaths
		}
	}
	if pkg.GlobalPipeTmpFilePath != "" {
		f.filePaths = append(f.filePaths, pkg.GlobalPipeTmpFilePath)
	}

	if f.sshPaths != nil {
		for _, sshPath := range f.sshPaths {
			sshFilePathConfig, err := pkg.StringToSSHPathConfig(sshPath)
			if err != nil {
				slog.Error("error parsing SSH path", err)
				break
			}
			if sshFilePathConfig != nil {
				sshConfig := pkg.SSHConfig{
					Host:           sshFilePathConfig.Host,
					Port:           sshFilePathConfig.Port,
					User:           sshFilePathConfig.User,
					Password:       sshFilePathConfig.Password,
					PrivateKeyPath: sshFilePathConfig.PrivateKeyPath,
				}
				fileInfos := pkg.GetFileInfos(sshFilePathConfig.FilePath, f.limit, true, &sshConfig)
				pkg.GlobalFilePaths = append(pkg.GlobalFilePaths, fileInfos...)
			}
		}
	}

	updateGlobalFilePaths()
}

func updateGlobalFilePaths() {
	fileInfos := []pkg.FileInfo{}
	for _, pattern := range f.filePaths {
		fileInfo := pkg.GetFileInfos(pattern, f.limit, false, nil)
		fileInfos = append(fileInfo, fileInfos...)
	}
	for _, pattern := range f.sshPaths {
		sshFilePathConfig, err := pkg.StringToSSHPathConfig(pattern)
		if err != nil {
			slog.Error("error parsing SSH path", err)
			break
		}
		sshConfig := pkg.SSHConfig{
			Host:           sshFilePathConfig.Host,
			Port:           sshFilePathConfig.Port,
			User:           sshFilePathConfig.User,
			Password:       sshFilePathConfig.Password,
			PrivateKeyPath: sshFilePathConfig.PrivateKeyPath,
		}
		pkg.GlobalPathSSHConfig = append(pkg.GlobalPathSSHConfig, *sshFilePathConfig)
		fileInfo := pkg.GetFileInfos(sshFilePathConfig.FilePath, f.limit, true, &sshConfig)
		fileInfos = append(fileInfo, fileInfos...)
	}

	for _, pattern := range f.dockerPaths {
		containers, _ := pkg.ListDockerContainers()
		if pattern == "" || len(strings.Fields(pattern)) == 1 {
			for _, container := range containers {
				if pattern != "" && !strings.Contains(container.Names[0], pattern) {
					continue
				}
				tmpFile := pkg.ContainerStdoutToTmp(container.ID)
				if tmpFile == nil {
					slog.Error("error creating temp file for container logs", "containerID", container.ID)
					continue
				}
				fileInfo := pkg.GetFileInfos(tmpFile.Name(), f.limit, false, nil)
				if len(fileInfo) > 0 {
					fileInfo[0].Host = container.ID[:12]
					fileInfo[0].Type = pkg.TypeDocker
					fileInfo[0].Name = container.Names[0][1:]
					fileInfos = append(fileInfo, fileInfos...)
				}
			}
		}
		if len(strings.Fields(pattern)) == 2 {
			dockerFilePathConfig, err := pkg.StringToDockerPathConfig(pattern)
			if err != nil {
				slog.Error("error parsing Docker path", err)
				break
			}
			fileInfo := pkg.GetContainerFileInfos(dockerFilePathConfig.FilePath, f.limit, dockerFilePathConfig.ContainerID)
			fileInfos = append(fileInfo, fileInfos...)
		}
	}

	pkg.GlobalFilePaths = uniqueFileInfos(fileInfos)
}

func cleanup() {
	slog.Info("cleaning up")
	if pkg.GlobalPipeTmpFilePath != "" {
		err := os.Remove(pkg.GlobalPipeTmpFilePath)
		if err != nil {
			slog.Error("error removing temp file", err)
			return
		}
		slog.Info("temp file removed", "path", pkg.GlobalPipeTmpFilePath)
	}
}

func watchFilePaths(seconds int64) {
	interval := time.Duration(seconds) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info("Checking for filepaths", "interval", interval)

	for range ticker.C {
		updateGlobalFilePaths()
	}
}

func uniqueFileInfos(fileInfos []pkg.FileInfo) []pkg.FileInfo {
	keys := make(map[string]bool)
	list := []pkg.FileInfo{}
	for _, entry := range fileInfos {
		key := entry.FilePath + entry.Type + entry.Host
		if _, value := keys[key]; !value {
			keys[key] = true
			list = append(list, entry)
		}
	}
	return list
}

func flags() {
	flag.Var(&f.filePaths, "f", "full path pattern to the log file")
	flag.Var(&f.sshPaths, "s", "full ssh path pattern to the log file")
	flag.Var(&f.dockerPaths, "d", "docker paths to the log file")
	flag.BoolVar(&f.version, "version", false, "")
	flag.BoolVar(&f.access, "access", false, "print access logs")
	flag.StringVar(&f.host, "host", "localhost", "host to serve")
	flag.Int64Var(&f.port, "port", 3003, "port to serve")
	flag.Int64Var(&f.every, "every", 10, "check for file paths every n seconds")
	flag.IntVar(&f.limit, "limit", 1000, "limit the number of files to read from the file path pattern")
	flag.Int64Var(&f.cors, "cors", 0, "cors port to allow the api (for development)")
	flag.BoolVar(&f.open, "open", true, "open browser on start")
	flag.StringVar(&f.baseURL, "base-url", "/", "base url with slash")

	flag.Parse()
}

func wantsVersion() {
	if f.version {
		fmt.Println(version)
		os.Exit(0)
	}
}
