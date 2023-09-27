package alert

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/yusufaine/apple-inventory-notifier/pkg/apple"
	"github.com/yusufaine/apple-inventory-notifier/pkg/set"
)

type Alert struct {
	Name  string
	Model string
	Shops map[string]bool
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

// type Alerts []Alert

func GenerateFromResponse(ar *apple.Response) []Alert {
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

	var alerts []Alert
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
	return alerts
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
