package main

import (
	"embed"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gookit/color"
	"github.com/k0kubun/pp"

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
	host      string
	port      int64
	cors      int64
	every     int64
	limit     int
	baseURL   string
	filePaths sliceFlags
	sshPaths  sliceFlags
	access    bool
	open      bool
	version   bool
}

var f Flags

var version = "dev"

func main() {
	flags()
	wantsVersion()

	if pkg.IsInputFromPipe() {
		tmpFile, err := os.Create(pkg.GetTmpFileName())
		if err != nil {
			color.New(color.FgRed).Println("error creating temp file: ", err)
			return
		}
		pkg.GlobalPipeTmpFilePath = tmpFile.Name()
		defer tmpFile.Close()
		go func(tmpFile *os.File) {
			err := pkg.PipeLinesToTmp(tmpFile)
			if err != nil {
				color.Danger.Println(err)
				return
			}
		}(tmpFile)
	}
	defaultFilePaths()

	go watchFilePaths(f.every)
	pp.Println(f)
	pp.Println("global filepaths", pkg.GlobalFilePaths)

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
		color.Danger.Println(err)
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

	if f.filePaths == nil && !pkg.IsInputFromPipe() {
		dir, _ := os.Getwd()
		f.filePaths = []string{
			dir + "/*/*log",
			dir + "/*log",
		}
		color.Info.Println("no file path provided, using ", f.filePaths)
	}

	if f.sshPaths != nil {
		for _, sshPath := range f.sshPaths {
			sshFilePathConfig, err := pkg.StringToSSHPathConfig(sshPath)
			if err != nil {
				color.Danger.Println(err)
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
			color.Danger.Println(err)
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
			pkg.GlobalPathSSHConfig = append(pkg.GlobalPathSSHConfig, *sshFilePathConfig)
			fileInfo := pkg.GetFileInfos(sshFilePathConfig.FilePath, f.limit, true, &sshConfig)
			fileInfos = append(fileInfo, fileInfos...)
		}
	}

	pkg.GlobalFilePaths = uniqueFileInfos(fileInfos)
}

func cleanup() {
	color.Info.Println("cleaning up")
	if pkg.GlobalPipeTmpFilePath != "" {
		err := os.Remove(pkg.GlobalPipeTmpFilePath)
		if err != nil {
			color.Danger.Println("error removing tmp file:", err)
			return
		}
		color.New(color.FgYellow).Println("tmp file removed:", pkg.GlobalPipeTmpFilePath)
	}
}

func watchFilePaths(seconds int64) {
	interval := time.Duration(seconds) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	color.Info.Println("Checking for filepaths every", interval)

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
		color.Secondary.Println(version)
		os.Exit(0)
	}
}
