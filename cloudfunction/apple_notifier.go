package cloudfunction

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/yusufaine/apple-inventory-notifier/internal/app/notifier"
	"github.com/yusufaine/apple-inventory-notifier/pkg/apple"
	_ "github.com/yusufaine/apple-inventory-notifier/pkg/log"
	"github.com/yusufaine/apple-inventory-notifier/pkg/mongodb"
	"github.com/yusufaine/apple-inventory-notifier/pkg/telegram"
)

func init() {
	functions.HTTP("apple_notifier", apple_notifierHandler)
}

func apple_notifierHandler(w http.ResponseWriter, r *http.Request) {
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

	appleReqParams := apple.RequestParamsFromBody(r.Body)
	defer r.Body.Close()

	tgBot := telegram.NewBot(telegram.NewConfig(r.Context()))
	alertCol := mongodb.NewAlertsConnection(mongodb.NewConfig(r.Context()))

	notifier.Start(appleReqParams, tgBot, alertCol)
}
