package pkg

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

func SetupLoggingStdout(logLevel slog.Leveler) {
	w := os.Stderr
	handler := tint.NewHandler(w, &tint.Options{
		NoColor:   !isatty.IsTerminal(w.Fd()),
		AddSource: true,
		Level:     logLevel,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				t := a.Value.Time()
				a.Value = slog.StringValue(t.Format(time.DateTime))
			}
			if a.Key == slog.SourceKey {
				source := a.Value.Any().(*slog.Source)
				a.Value = slog.StringValue(filepath.Base(source.File) + ":" + fmt.Sprint(source.Line))
			}
			return a
		},
	})

	slog.SetDefault(slog.New(handler))
}
