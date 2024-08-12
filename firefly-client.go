package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"io"

	"github.com/go-resty/resty/v2"
)

type FireflyClient struct {
	client *resty.Client
}

func NewFireflyClient(baseURL, token string) *FireflyClient {
	client := resty.New().
		SetBaseURL(baseURL).
		SetHeader("Authorization", "Bearer "+token).
		SetHeader("accept", "application/vnd.api+json").
		SetHeader("Content-Type", "application/json")

	return &FireflyClient{client: client}
}

// getAccounts retrieves accounts from the Firefly API.
func getAccounts(token string, params RequestParams) (FireflyAccountResponse, error) {
	var response FireflyAccountResponse

	client := &http.Client{}
	baseURL := "http://192.168.50.32:18888/api/v1/accounts"

	// Build the request URL with query parameters
	queryParams := url.Values{}
	queryParams.Set("limit", fmt.Sprintf("%d", params.Limit))
	queryParams.Set("page", fmt.Sprintf("%d", params.Page))
	if params.Date != "" {
		queryParams.Set("date", params.Date)
	}
	if params.Type != "" {
		queryParams.Set("type", params.Type)
	}

	req, err := http.NewRequest("GET", baseURL+"?"+queryParams.Encode(), nil)
	if err != nil {
		return response, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	//req.Header.Set("X-Trace-Id", "your-unique-uuid")

	resp, err := client.Do(req)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.Reader(resp.Body))
	if err != nil {
		return response, err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return response, err
	}

	return response, nil
}

func (f *FireflyClient) CreateAccount(account FireflyAccount) (*FireflyAccount, error) {
	var response struct {
		Data FireflyAccount `json:"data"`
	}

	resp, err := f.client.R().SetBody(account).SetResult(&response).Post("/v1/accounts")
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, logError(resp)
	}

	return &response.Data, nil
}

func (f *FireflyClient) CreateTransaction(transaction FireflyTransaction) error {
	resp, err := f.client.R().SetBody(transaction).Post("/v1/transactions")
	if err != nil {
		return err
	}

	if resp.IsError() {
		return logError(resp)
	}

	return nil
}

func logError(resp *resty.Response) error {
	log.Printf("Request failed with status %d: %s", resp.StatusCode(), resp.Body())
	return fmt.Errorf("request failed with status %d", resp.StatusCode())
}
