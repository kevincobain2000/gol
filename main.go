package main

import (
	"embed"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

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
	every    int64
	baseURL  string
	filePath string
	access   bool
	open     bool
	version  bool
}

var f Flags

var version = "dev"

func main() {
	flags()
	wantsVersion()
	if f.filePath == "" {
		color.Danger.Println("-f filepath is required")
		os.Exit(1)
	}
	pkg.GlobalFilePaths = getFilePaths()
	go watchFilePaths(f.every)
	pp.Sprintln(f)
	pp.Sprintln("filePaths", pkg.GlobalFilePaths)

	if f.open {
		openBrowser(fmt.Sprintf("http://%s:%d%s", f.host, f.port, f.baseURL))
	}

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
		os.Exit(1)
	}
}

func watchFilePaths(seconds int64) {
	interval := time.Duration(seconds) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	color.Info.Println("Checking for filepaths every", interval)

	for range ticker.C {
		pkg.GlobalFilePaths = getFilePaths()
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

func getFilePaths() []string {
	filePaths, err := pkg.FilesByPattern(f.filePath)
	if err != nil {
		color.Danger.Println(err)
		return nil
	}
	if len(filePaths) == 0 {
		color.Danger.Println("no files found ", f.filePath)
		return nil
	}
	readableFilePaths := make([]string, 0)
	for _, filePath := range filePaths {
		isText, err := pkg.IsReadableFile(filePath)
		if err != nil {
			color.Danger.Println(err)
			return nil
		}
		if !isText {
			color.Warn.Println("file is not a text file ", filePath)
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
	flag.BoolVar(&f.access, "access", false, "print access logs")
	flag.StringVar(&f.host, "host", "localhost", "host to serve")
	flag.Int64Var(&f.port, "port", 3003, "port to serve")
	flag.Int64Var(&f.every, "every", 10, "check for file paths every n seconds")
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
