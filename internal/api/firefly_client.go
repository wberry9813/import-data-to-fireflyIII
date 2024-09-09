package api

import (
	"fmt"
	"log"
	"net/url"

	. "import-data-to-fireflyIII/internal/models"

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

// GetAccounts retrieves accounts from the Firefly API.
func (f *FireflyClient) GetAccounts(params RequestParams) (FireflyAccountResponse, error) {
	var response FireflyAccountResponse

	// 构建查询参数
	queryParams := url.Values{}
	queryParams.Set("limit", fmt.Sprintf("%d", params.Limit))
	queryParams.Set("page", fmt.Sprintf("%d", params.Page))
	if params.Date != "" {
		queryParams.Set("date", params.Date)
	}
	if params.Type != "" {
		queryParams.Set("type", params.Type)
	}

	// 发送请求
	resp, err := f.client.R().
		SetQueryParamsFromValues(queryParams).
		SetResult(&response).
		Get("/api/v1/accounts")

	if err != nil {
		return response, err
	}

	if resp.IsError() {
		return response, fmt.Errorf("request failed with status %d", resp.StatusCode())
	}

	return response, nil
}

func (f *FireflyClient) CreateAccount(account FireflyAccountRequest) (*CreateFireflyAccountResponse, error) {
	var response CreateFireflyAccountResponse

	resp, err := f.client.R().SetBody(account).SetResult(&response).Post("/api/v1/accounts")
	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, logError(resp)
	}

	return &response, nil
}

func (f *FireflyClient) CreateTransaction(transaction FireflyTransaction) error {
	resp, err := f.client.R().SetBody(transaction).Post("/api/v1/transactions")
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
