package alert

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/yusufaine/apple-inventory-notifier/pkg/apple"
	"github.com/yusufaine/apple-inventory-notifier/pkg/set"
	"go.mongodb.org/mongo-driver/bson"
)

type Alerts []Alert

// Returns a slice of alerts with the newAlerts' message ID being the old alerts.
// If alert is new (not replacing old), new alert's ID is left as default.
// Note that the calling struct must be the old one.
//
//	oldAlerts.GetDifferenceWithOldIDs(newAlerts)
func (a *Alerts) GetDifferenceWithOldIDs(newAlerts *Alerts) *Alerts {
	oldMap := a.asModelAlertMap()
	newMap := newAlerts.asModelAlertMap()

	var alerts Alerts
	for model, newAlert := range newMap {
		oldAlert, found := oldMap[model]
		// new alert is indeed new
		if !found {
			alerts = append(alerts, newAlert)
		} else if hasShopChanges(oldAlert, newAlert) {
			// update new alerts' msg id to be the old alerts'
			// used for deletion later
			newAlert.MsgId = oldAlert.MsgId
			alerts = append(alerts, newAlert)
		}
	}

	return &alerts
}

func (a *Alerts) asModelAlertMap() map[string]Alert {
	m := make(map[string]Alert)
	for _, alert := range *a {
		m[alert.Model] = alert
	}
	return m
}

func hasShopChanges(a1, a2 Alert) bool {
	if len(a1.Shops) != len(a2.Shops) {
		return true
	}

	for shop, avail := range a1.Shops {
		if avail != a2.Shops[shop] {
			return true
		}
	}

	return false
}

type Alert struct {
	Model string          `bson:"_id"` // index, unique ID
	Name  string          `bson:"name"`
	Shops map[string]bool `bson:"shops"`
	MsgId int             `bson:"msg_id"`
}

func (a *Alert) ToBSON() bson.M {
	return bson.M{
		"_id":    a.Model,
		"name":   a.Name,
		"shops":  a.Shops,
		"msg_id": a.MsgId,
	}
}

func (a *Alert) ToTelegramHTMLString() string {
	var sb strings.Builder
	sb.WriteString("<b>" + a.Name + "</b>\n")
	sb.WriteString(a.Model + "\n")
	sb.WriteString("\n<b>Availability</b>\n")

	sortedShops := make([]string, 0, len(a.Shops))
	for k := range a.Shops {
		sortedShops = append(sortedShops, k)
	}

	slices.Sort(sortedShops)

	for _, shop := range sortedShops {
		avail := a.Shops[shop]
		sb.WriteString(fmt.Sprintf("%s -- %s\n", shop, availEmoji(avail)))
	}

	return sb.String()
}

func availEmoji(b bool) string {
	if b {
		return "✅"
	}
	return "❌"
}

func GenerateFromResponse(ar *apple.Response) *Alerts {
	// map shops to their inventory
	shopModelsLookup := make(map[string][]string)
	for _, store := range ar.Body.Stores {
		models, found := shopModelsLookup[store.Name]

		// Populate modelSet
		for model, part := range store.PartInfo {
			if isAvailable(part.StoreIsEligible, part.PartIsAvailable, model) {
				models = append(models, model)
			}
		}
		if !found {
			shopModelsLookup[store.Name] = models
		}
	}

	// map model to name
	modelNameLookup := make(map[string]string)
	for _, store := range ar.Body.Stores {
		for model, part := range store.PartInfo {
			if _, found := modelNameLookup[model]; found {
				continue
			}

			mt, msgFound := part.MessageTypes[apple.MessageTypesRegularKey]
			if !msgFound {
				panic("'regular' not found in response's message types")
			}
			modelNameLookup[model] = mt.ProductName
		}
	}

	var alerts Alerts
	for model, name := range modelNameLookup {
		alert := Alert{Name: name, Model: model}
		shops := make(map[string]bool)
		for shop, models := range shopModelsLookup {
			modelSet := set.FromStrings(models...)
			shops[shop] = modelSet.Contains(model)
		}
		alert.Shops = shops
		alerts = append(alerts, alert)
	}

	slog.Info("generated alerts",
		slog.Int("alert_count", len(alerts)),
		slog.Int("model_count", len(modelNameLookup)))

	return &alerts
}

func isAvailable(storeAvail bool, partAvail string, model string) bool {
	part := false
	switch strings.ToLower(partAvail) {
	case "available":
		part = true
	case "unavailable": // do nothing
	default:
		panic(fmt.Sprintf("unable to parse availability: %s", partAvail))
	}

	return part && storeAvail
}
