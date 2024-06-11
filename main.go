package main

import (
	"embed"
	"flag"
	"fmt"
	"os"
	"os/signal"
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

	if pkg.IsInputFromPipe() {
		go pkg.ReadLinesFromPipe()
	}
	if f.filePath == "" {
		dir, _ := os.Getwd()
		f.filePath = dir + "/*log"
		color.Info.Println("no file path provided, using ", f.filePath)
	}
	pkg.GlobalFilePaths = getFilePaths()
	go watchFilePaths(f.every)
	pp.Println(f)
	pp.Println("filePaths", pkg.GlobalFilePaths)

	if f.open {
		pkg.OpenBrowser(fmt.Sprintf("http://%s:%d%s", f.host, f.port, f.baseURL))
	}
	defer cleanup()
	handleCltrC()

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

func handleCltrC() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		s := <-c
		color.Warn.Println("got signal:", s)
		cleanup()
		close(c)
		os.Exit(1)
	}()
}

func cleanup() {
	color.Info.Println("cleaning up")
	if pkg.GlobalTempFilePath != "" {
		err := os.Remove(pkg.GlobalTempFilePath)
		if err != nil {
			color.Danger.Println("error removing tmp file:", err)
			return
		}
		color.New(color.FgYellow).Println("tmp file removed:", pkg.GlobalTempFilePath)
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

func getFilePaths() []string {
	filePaths, err := pkg.FilesByPattern(f.filePath)
	if err != nil {
		color.Danger.Println(err)
		return nil
	}
	if len(filePaths) == 0 {
		color.Danger.Println("no files found:", f.filePath)
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
			color.Warn.Println("file is not a text file:", filePath)
			continue
		}
		readableFilePaths = append(readableFilePaths, filePath)
	}
	return readableFilePaths
}

func flags() {

	flag.StringVar(&f.filePath, "f", "", "full path to the log file")
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
