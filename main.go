package main

import (
	"embed"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/gookit/color"
	"github.com/k0kubun/pp"

	"github.com/kevincobain2000/gol/pkg"
)

//go:embed all:frontend/dist/*
var publicDir embed.FS

type Flags struct {
	host     string
	port     int64
	cors     int64
	baseURL  string
	filePath string
	open     bool
	version  bool
}

var f Flags

var version = "dev"

func main() {
	flags()
	wantsVersion()
	filePaths := validateFilePath()
	pp.Sprint(f)
	pp.Sprint("filePaths", filePaths)

	if f.open {
		openBrowser(fmt.Sprintf("http://%s:%d%s", f.host, f.port, f.baseURL))
	}

	err := pkg.NewEcho(func(o *pkg.EchoOptions) error {
		o.Host = f.host
		o.Port = f.port
		o.Cors = f.cors
		o.BaseURL = f.baseURL
		o.PublicDir = &publicDir
		o.FilePaths = filePaths
		return nil
	})
	if err != nil {
		color.Danger.Print(err)
		os.Exit(1)
	}
}

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = exec.Command("xdg-open", url).Start()
	}

	if err != nil {
		color.Warn.Println("Failed to open browser")
	}
}

func validateFilePath() []string {
	if f.filePath == "" {
		color.Danger.Print("file-path is required")
		os.Exit(1)
	}

	filePaths, err := pkg.FilesByPattern(f.filePath)
	if err != nil {
		color.Danger.Print(err)
		os.Exit(1)
	}
	if len(filePaths) == 0 {
		color.Danger.Print("no files found", f.filePath)
		os.Exit(1)
	}
	readableFilePaths := make([]string, 0)
	for _, filePath := range filePaths {
		isText, err := pkg.IsReadableFile(filePath)
		if err != nil {
			color.Danger.Print(err)
			os.Exit(1)
		}
		if !isText {
			color.Warn.Print("file is not a text file", filePath)
			continue
		}
		readableFilePaths = append(readableFilePaths, filePath)
	}
	return readableFilePaths
}

func flags() {
	dir, _ := os.Getwd()
	flag.StringVar(&f.filePath, "f", dir+"/*log", "full path to the log file")
	flag.BoolVar(&f.version, "version", false, "")
	flag.StringVar(&f.host, "host", "localhost", "host to serve")
	flag.Int64Var(&f.port, "port", 3003, "port to serve")
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
