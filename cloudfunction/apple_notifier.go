package cloudfunction

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/yusufaine/apple-inventory-notifier/internal/app/notifier"
	"github.com/yusufaine/apple-inventory-notifier/pkg/apple"
	"github.com/yusufaine/apple-inventory-notifier/pkg/mongodb"
	"github.com/yusufaine/apple-inventory-notifier/pkg/telegram"
)

func init() {
	functions.HTTP("apple_notifier", apple_notifierHandler)
}

func apple_notifierHandler(w http.ResponseWriter, r *http.Request) {
	setupSlog()
	defer func() {
		if r := recover(); r != nil {
			switch err := r.(type) {
			case error:
				slog.Error(err.Error())
			case string:
				slog.Error(err)
			case fmt.Stringer:
				slog.Error(err.String())
			default:
				slog.Error(fmt.Sprintf("%#v", err))
			}
			fmt.Fprintln(w, r)
			return
		}

		fmt.Fprintln(w, "notifier completed")
		slog.Info("notifier completed")
	}()

	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		slog.Error("failed to read request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var appleReqParams apple.RequestParams
	if err := json.Unmarshal(b, &appleReqParams); err != nil {
		slog.Error("failed to unmarshal request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	validateRequest(appleReqParams)

	tgBot := telegram.NewBot(telegram.NewConfig(r.Context()))
	alertCol := mongodb.NewAlertsConnection(mongodb.NewConfig(r.Context()))
	defer alertCol.Close()

	notifier.Start(&appleReqParams, tgBot, alertCol)
}

func setupSlog() {
	const cloudFuncKey = "FUNCTION_SIGNATURE_TYPE"
	// if local environment, don't change slog
	if _, found := os.LookupEnv(cloudFuncKey); !found {
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

func validateRequest(r apple.RequestParams) {
	if len(r.AbbrevCountry) == 0 {
		panic("'abbrev_country' must be a non-empty string")
	}
	if len(r.Country) == 0 {
		panic("'country' must be a non-empty string")
	}
	if len(r.Models) == 0 {
		panic("'models' must be a non-empty string array")
	}
}
