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
		tmpfile, err := os.Create(pkg.GetTmpFileName())
		if err != nil {
			color.New(color.FgRed).Println("error creating temp file: ", err)
			return
		}
		pkg.GlobalTmpFilePath = tmpfile.Name()
		defer tmpfile.Close()
		go func(tmpfile *os.File) {
			err := pkg.PipeLinesToTmp(tmpfile)
			if err != nil {
				color.Danger.Println(err)
				return
			}
		}(tmpfile)
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
	if pkg.GlobalTmpFilePath != "" {
		f.filePaths = append(f.filePaths, pkg.GlobalTmpFilePath)
	}

	if f.filePaths == nil && !pkg.IsInputFromPipe() {
		dir, _ := os.Getwd()
		f.filePaths = []string{
			dir + "/*/*log",
			dir + "/*log",
		}
		color.Info.Println("no file path provided, using ", f.filePaths)
	}

	updateGlobalFilePaths()
}

func updateGlobalFilePaths() {
	fileInfos := []pkg.FileInfo{}
	for _, pattern := range f.filePaths {
		fileInfo := pkg.GetFileInfos(pattern, f.limit)
		fileInfos = append(fileInfo, fileInfos...)
	}
	// update type to stdin in GlobalFilePaths if it has a file with name tmpfile.Name()
	for i, fileInfo := range fileInfos {
		if fileInfo.FilePath == pkg.GlobalTmpFilePath {
			fileInfos[i].Type = "stdin"
		}
	}
	pkg.GlobalFilePaths = uniqueFileInfos(fileInfos)
}

func cleanup() {
	color.Info.Println("cleaning up")
	if pkg.GlobalTmpFilePath != "" {
		err := os.Remove(pkg.GlobalTmpFilePath)
		if err != nil {
			color.Danger.Println("error removing tmp file:", err)
			return
		}
		color.New(color.FgYellow).Println("tmp file removed:", pkg.GlobalTmpFilePath)
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
		if _, value := keys[entry.FilePath]; !value {
			keys[entry.FilePath] = true
			list = append(list, entry)
		}
	}
	return list
}

func flags() {
	flag.Var(&f.filePaths, "f", "full path pattern to the log file")
	flag.BoolVar(&f.version, "version", false, "")
	flag.BoolVar(&f.access, "access", false, "print access logs")
	flag.StringVar(&f.host, "host", "localhost", "host to serve")
	flag.Int64Var(&f.port, "port", 3003, "port to serve")
	flag.Int64Var(&f.every, "every", 10, "check for file paths every n seconds")
	flag.IntVar(&f.limit, "limit", 1000, "limit the number of files to read from the file path pattern")
	flag.Int64Var(&f.cors, "cors", 0, "cors port to allow")
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
