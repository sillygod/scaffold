package app

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
)

type Price struct {
	Price       string `json:"price"`
	Conf        string `json:"conf"`
	Expo        int    `json:"expo"`
	PublishTime int64  `json:"publish_time"`
}

type EmaPrice struct {
	Price       string `json:"price"`
	Conf        string `json:"conf"`
	Expo        int    `json:"expo"`
	PublishTime int64  `json:"publish_time"`
}

type Metadata struct {
	Slot               int64 `json:"slot"`
	ProofAvailableTime int64 `json:"proof_available_time"`
	PrevPublishTime    int64 `json:"prev_publish_time"`
}

type Parsed struct {
	ID       string   `json:"id"`
	Price    Price    `json:"price"`
	EmaPrice EmaPrice `json:"ema_price"`
	Metadata Metadata `json:"metadata"`
}

type ApiResponse struct {
	Binary struct {
		Encoding string   `json:"encoding"`
		Data     []string `json:"data"`
	} `json:"binary"`
	Parsed []Parsed `json:"parsed"`
}

// Function to build the request URL with dynamic IDs

type PythAPIClient struct {
	baseURL string
}

func NewPythAPIClient(baseURL string) *PythAPIClient {
	return &PythAPIClient{
		baseURL: baseURL,
	}
}

func (p *PythAPIClient) buildGetLatestPricesURL(ids []string) string {
	baseURL, err := url.Parse(p.baseURL)
	if err != nil {
		log.Fatal(err)
	}
	baseURL.Path = "/v2/updates/price/latest"
	params := url.Values{}
	for _, id := range ids {
		params.Add("ids[]", id)
	}
	baseURL.RawQuery = params.Encode()
	return baseURL.String()
}

func (p *PythAPIClient) GetLatestPrices(ids []string) (*ApiResponse, error) {
	url := p.buildGetLatestPricesURL(ids)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	var apiResp ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}
	return &apiResp, nil

}
