package apple

const MessageTypesRegularKey = "regular" // if not present, response shape changed

// Kinda disgusting but splitting it up is worst, at least this can be collapsed :cope:
type Response struct {
	Body struct {
		Stores []struct {
			Name     string `json:"storeName"`
			PartInfo map[string]struct {
				StoreIsEligible bool   `json:"storePickEligible"`
				PartIsAvailable string `json:"pickupDisplay"`
				MessageTypes    map[string]struct {
					ProductName string `json:"storePickupProductTitle"`
				} `json:"messageTypes"`
			} `json:"partsAvailability"` // key is model
		} `json:"stores"`
	} `json:"body"`
}
