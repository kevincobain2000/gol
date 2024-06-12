package pkg

import (
	"bufio"
	"os"
	"os/exec"
	"os/signal"
	"runtime"

	"github.com/gookit/color"
	"github.com/k0kubun/pp"
	g "github.com/kevincobain2000/go-human-uuid/lib"
)

func GetHomedir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

// IsInputFromPipe checks if there is input from a pipe
func IsInputFromPipe() bool {
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	// Check if the mode is a character device (i.e., a pipe)
	return (fileInfo.Mode() & os.ModeCharDevice) == 0
}

func PipeLinesToTmp(tmpfile *os.File) error {
	scanner := bufio.NewScanner(os.Stdin)

	color.Info.Println("tmp file created for stdin: ", GlobalTmpFilePath)

	tempFileInfo, err := createTmpFileInfo(GlobalTmpFilePath)
	if err != nil {
		color.Danger.Println("error creating FileInfo for temp file:", err)
	} else {
		GlobalFilePaths = append([]FileInfo{tempFileInfo}, GlobalFilePaths...)
	}
	color.Info.Println("tmp file created added to global filepaths: ", pp.Sprint(GlobalFilePaths))

	lineCount := 0
	for scanner.Scan() {
		line := scanner.Text()
		if lineCount >= 10000 {
			if err := tmpfile.Truncate(0); err != nil {
				color.Danger.Println("error truncating file: ", err)
			}
			if _, err := tmpfile.Seek(0, 0); err != nil {
				color.Danger.Println("error seeking file: ", err)
			}
			lineCount = 0
		}
		if _, err := tmpfile.WriteString(line + "\n"); err != nil {
			color.Danger.Println("error writing to file: ", err)
		}
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		color.New(color.FgRed).Println("error reading from pipe: ", err)
		return err
	}

	return nil
}

func createTmpFileInfo(filePath string) (FileInfo, error) {
	linesCount, fileSize, err := FileStats(filePath)
	if err != nil {
		return FileInfo{}, err
	}
	return FileInfo{FilePath: filePath, LinesCount: linesCount, FileSize: fileSize, Type: "stdin"}, nil
}

func GetTmpFileName() string {
	gen, _ := g.NewGenerator([]g.Option{
		func(opt *g.Options) error {
			opt.Length = 2
			return nil
		},
	}...)
	return "/tmp/GOL-STDIN-" + gen.Generate()
}

func OpenBrowser(url string) {
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

func HandleCltrC(f func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		s := <-c
		color.Warn.Println("got signal:", s)
		f()
		close(c)
		os.Exit(1)
	}()
}
