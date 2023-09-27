package cloudfunction

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/yusufaine/apple-inventory-notifier/internal/app/notifier"
	"github.com/yusufaine/apple-inventory-notifier/pkg/apple"
	"github.com/yusufaine/apple-inventory-notifier/pkg/log"
	"github.com/yusufaine/apple-inventory-notifier/pkg/tg"
)

func init() {
	functions.HTTP("apple_notifier", apple_notifierHandler)
}

func apple_notifierHandler(w http.ResponseWriter, r *http.Request) {
	log.SetupSlog()
	defer func() {
		if r := recover(); r != nil {
			logMsg := "apple notifier error"
			switch err := r.(type) {
			case error:
				slog.Error(logMsg, slog.String("error", err.Error()))
			case string:
				slog.Error(logMsg, slog.String("error", err))
			case fmt.Stringer:
				slog.Error(logMsg, slog.String("error", err.String()))
			default:
				slog.Error(logMsg, slog.String("error", fmt.Sprintf("%#v", err)))
			}
			fmt.Fprintln(w, r)
			return
		}

		fmt.Fprintln(w, "notifier completed")
	}()

	appleReqParams := apple.RequestParamsFromBody(r.Body)
	defer r.Body.Close()

	tgBot := tg.NewBot(tg.NewConfig(r.Context()))

	if err := notifier.Start(appleReqParams, tgBot); err != nil {
		slog.Error("notifier failed", slog.String("error", err.Error()))
		fmt.Fprintln(w, "notifier failed")
	}
	slog.Info("notifier completed")
}
