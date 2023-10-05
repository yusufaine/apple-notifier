package apple

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

/*
Example:

	{
	  "models": ["MU7J3ZP/A", "MU7E3ZP/A", "MU793ZP/A"],
	  "abbrev_country": "sg",
	  "country": "singapore"
	}
*/

// Fields are guaranteed to be not empty
type RequestParams struct {
	AbbrevCountry string   `json:"abbrev_country"`
	Country       string   `json:"country"`
	Models        []string `json:"models"`
}

// Returns the raw content of the response to the query
func (r *RequestParams) QueryApple() (*Response, error) {
	ep := generatePickupUrl(r.AbbrevCountry, r.Country, r.Models)
	req, err := http.NewRequest(http.MethodGet, ep, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp == nil {
		return nil, fmt.Errorf("response body is nil")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ar Response
	if err := json.Unmarshal(data, &ar); err != nil {
		return nil, err
	}

	return &ar, nil
}

// Example:
//
//	abbrevCountry := "sg"
//	country       := "singapore"
func generatePickupUrl(abbrevCountry, country string, models []string) string {
	ep := url.URL{
		Scheme: "https",
		Host:   "www.apple.com",
		Path:   fmt.Sprintf("/%s/shop/retail/pickup-message", abbrevCountry),
	}

	rq := url.Values{
		"pl":       {"true"},
		"location": {country},
	}

	for i, k := range models {
		key := fmt.Sprintf("parts.%d", i)
		rq[key] = []string{k}
	}

	ep.RawQuery = rq.Encode()

	return ep.String()
}
