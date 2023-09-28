package notifier

import (
	"log/slog"

	"github.com/yusufaine/apple-inventory-notifier/pkg/alert"
	"github.com/yusufaine/apple-inventory-notifier/pkg/apple"
	"github.com/yusufaine/apple-inventory-notifier/pkg/mongodb"
	"github.com/yusufaine/apple-inventory-notifier/pkg/set"
	"github.com/yusufaine/apple-inventory-notifier/pkg/telegram"
	"go.mongodb.org/mongo-driver/bson"
)

func Start(ap *apple.RequestParams, tgBot *telegram.Bot, alertCol *mongodb.Collection) {
	// Prune alerts
	msgIdsToDelete := pruneStaleMongoAlertsByModel(alertCol, ap.Models)
	if len(msgIdsToDelete) != 0 {
		for _, msgId := range msgIdsToDelete {
			if err := tgBot.Delete(msgId); err != nil {
				slog.Error("unable to prune message",
					slog.Int("msg_id", msgId),
					slog.String("error", err.Error()))
			}
		}
	} else {
		slog.Info("no alerts to prune")
	}

	// Get fresh alerts
	parsedResponse, err := ap.Do()
	if err != nil {
		panic(err)
	}
	generatedAlerts := alert.GenerateFromResponse(parsedResponse)
	slog.Info("generated new alerts from apple stores", slog.Int("alert_count", len(*generatedAlerts)))

	reqModels := set.FromStrings(ap.Models...)
	for _, alert := range *generatedAlerts {
		if reqModels.Contains(alert.Model) {
			continue
		}
		slog.Warn("unable to generate alert for " + alert.Model)
	}

	// Get difference between existing and fresh alerts
	existingAlerts := alertCol.GetAlerts()
	toUpdate := *existingAlerts.GetDifferenceWithOldIDs(generatedAlerts)
	if len(toUpdate) == 0 {
		slog.Info("messages are all up-to-date")
		return
	}

	// Collect message IDs to delete
	msgIdsToDelete = make([]int, 0)
	for i, alert := range toUpdate {
		// Delete old message
		if alert.MsgId != 0 {
			if err := tgBot.Delete(alert.MsgId); err != nil {
				slog.Error("unable to delete message",
					slog.Int("msg_id", alert.MsgId),
					slog.String("error", err.Error()))
			}
		}

		// Send new message and update msgId
		newId, err := tgBot.Send(alert.ToTelegramHTMLString(), telegram.ParseHTML)
		if err != nil {
			slog.Error("unable to send message", slog.String("error", err.Error()))
			continue
		}

		// Collect msgIds for deletion
		if alert.MsgId != 0 {
			msgIdsToDelete = append(msgIdsToDelete, alert.MsgId)
		}

		// Update msgIds for reinsertion
		alert.MsgId = newId
		toUpdate[i] = alert
	}

	// Update mongo, delete dangling alerts, reinsert updated alerts
	alertCol.DeleteAlertsByFilter(bson.M{"msg_id": bson.M{"$in": msgIdsToDelete}})
	alertCol.InsertAlerts(&toUpdate)

	return
}

// Deletes stale mongo alerts and returns their corresponding message IDs
func pruneStaleMongoAlertsByModel(alertCol *mongodb.Collection, newModels []string) []int {
	var msgIds []int
	filter := bson.M{"_id": bson.M{"$nin": newModels}}
	toPrune := alertCol.GetAlertsByFilter(filter)
	if len(*toPrune) == 0 {
		return msgIds
	}

	// Collect message IDs
	for _, alert := range *toPrune {
		msgIds = append(msgIds, alert.MsgId)
	}

	count := alertCol.DeleteAlertsByFilter(filter)
	if count != 0 {
		slog.Info("pruned alerts", slog.Int64("count", count))
	}

	return msgIds
}
