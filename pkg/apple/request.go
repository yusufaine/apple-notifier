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

func RequestParamsFromBody(body io.ReadCloser) *RequestParams {
	b, err := io.ReadAll(body)
	defer body.Close()
	if err != nil {
		panic("unable to read request body")
	}

	// Initialise params needed for Apple query
	var appleReq RequestParams
	if err := json.Unmarshal(b, &appleReq); err != nil {
		panic("unable to unmarshal request body")
	}
	appleReq.MustValidate()

	return &appleReq
}

func (r *RequestParams) MustValidate() {
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

// Returns the raw content of the response to the query
func (r *RequestParams) Do() (*Response, error) {
	ep := GeneratePickupUrl(r.AbbrevCountry, r.Country, r.Models)
	req, err := http.NewRequest(http.MethodGet, ep, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
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
func GeneratePickupUrl(abbrevCountry, country string, models []string) string {
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
