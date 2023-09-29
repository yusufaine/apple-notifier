// Initialise slog for cloud function, should only be imported at entrypoint
package log

import (
	"log/slog"
	"os"
	"path/filepath"
)

const CloudFuncKey = "FUNCTION_SIGNATURE_TYPE"

func init() {
	// if local environment, don't change slog
	if _, found := os.LookupEnv(CloudFuncKey); !found {
		return
	}

	replaceFunc := func(groups []string, a slog.Attr) slog.Attr {
		if (a.Key == slog.TimeKey && len(groups) == 0) ||
			a.Key == slog.SourceKey {
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
