package pkg

import (
	"bufio"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"runtime"

	"github.com/acarl005/stripansi"
	"github.com/kevincobain2000/go-human-uuid/lib"
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

func PipeLinesToTmp(tmpFile *os.File) error {
	scanner := bufio.NewScanner(os.Stdin)

	slog.Info("Temporary file created for stdin", "path", GlobalPipeTmpFilePath)

	linesCount, fileSize, err := FileStats(GlobalPipeTmpFilePath, false, nil)
	if err != nil {
		slog.Error("Error creating FileInfo for temp file", err)
		return err
	}
	tempFileInfo := FileInfo{FilePath: GlobalPipeTmpFilePath, LinesCount: linesCount, FileSize: fileSize, Type: TypeStdin}

	GlobalFilePaths = append([]FileInfo{tempFileInfo}, GlobalFilePaths...)
	slog.Info("Temporary file added to global file paths", "filePaths", GlobalFilePaths)

	lineCount := 0
	for scanner.Scan() {
		line := scanner.Text()
		line = stripansi.Strip(line)
		if lineCount >= 10000 {
			if err := tmpFile.Truncate(0); err != nil {
				slog.Error("Error truncating file", err)
			}
			if _, err := tmpFile.Seek(0, 0); err != nil {
				slog.Error("Error seeking file", err)
			}
			lineCount = 0
		}
		if _, err := tmpFile.WriteString(line + "\n"); err != nil {
			slog.Error("Error writing to file", err)
		}
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		slog.Error("Error reading from pipe", err)
		return err
	}

	return nil
}

func GetTmpFileNameForSTDIN() string {
	gen, _ := lib.NewGenerator([]lib.Option{
		func(opt *lib.Options) error {
			opt.Length = 2
			return nil
		},
	}...)
	return TmpStdinPath + gen.Generate()
}

func GetTmpFileNameForContainer() string {
	gen, _ := lib.NewGenerator([]lib.Option{
		func(opt *lib.Options) error {
			opt.Length = 6
			return nil
		},
	}...)
	return TmpContainerPath + gen.Generate()
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
		slog.Warn("Failed to open browser", "url", url)
	}
}

func HandleCltrC(f func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		s := <-c
		slog.Warn("Got signal", "signal", s)
		f()
		close(c)
		os.Exit(1)
	}()
}
