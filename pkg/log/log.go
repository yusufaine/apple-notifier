// Initialise slog for cloud function, should only be imported at entrypoint
package log

import (
	"log/slog"
	"os"
	"path/filepath"
)

var ProjectID = os.Getenv("GOOGLE_CLOUD_PROJECT")

func SetupSlog() {
	// if local environment, don't change slog
	if ProjectID == "" {
		return
	}

	replaceFunc := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey && len(groups) == 0 {
			return slog.Attr{}
		}
		if a.Key == slog.LevelKey {
			a.Key = "severity"
		}
		if a.Key == slog.MessageKey {
			a.Key = "message"
		}
		if a.Key == slog.SourceKey {
			source := a.Value.Any().(*slog.Source)
			source.File = filepath.Base(source.File)
		}
		return a
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		ReplaceAttr: replaceFunc,
	})))
}
