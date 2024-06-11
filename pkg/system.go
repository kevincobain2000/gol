package pkg

import (
	"bufio"
	"os"
	"os/exec"
	"runtime"

	"github.com/gookit/color"
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

func ReadLinesFromPipe() error {
	scanner := bufio.NewScanner(os.Stdin)
	tmpfile, err := os.Create(getTmpFileName())
	if err != nil {
		color.New(color.FgRed).Println("error creating temp file: ", err)
		return err
	}
	GlobalTempFilePath = tmpfile.Name()
	color.Info.Println("tmp file created for stdin: ", GlobalTempFilePath)
	defer tmpfile.Close()

	lineCount := 0
	for scanner.Scan() {
		line := scanner.Text()
		if lineCount >= 10000 {
			tmpfile.Truncate(0)
			tmpfile.Seek(0, 0)
			lineCount = 0
		}
		tmpfile.WriteString(line + "\n")
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		color.New(color.FgRed).Println("error reading from pipe: ", err)
		return err
	}

	return nil
}

func getTmpFileName() string {
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
