package notifier

import (
	"github.com/yusufaine/apple-inventory-notifier/pkg/alert"
	"github.com/yusufaine/apple-inventory-notifier/pkg/apple"
	"github.com/yusufaine/apple-inventory-notifier/pkg/tg"
)

func Start(ap *apple.RequestParams, tgBot *tg.Bot) error {
	parsedResponse, err := ap.Do()
	if err != nil {
		return err
	}

	alerts := alert.GenerateFromResponse(parsedResponse)
	for _, alert := range alerts {
		tgBot.Write(alert.ToTelegramHTMLString(), tg.ParseHTML)
	}

	return nil
}
