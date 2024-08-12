package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	FireflyToken = "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJhdWQiOiIxIiwianRpIjoiNTQ0ODJiYzhjY2ZiOTFkODhkNmY3ZjRiNTY1MGYwODIzOTM5ODUxMTdmOTI1MmJlZGJlZTRlYzlhODFlOGJiZWZhODY3M2U1ZjkyYmJmODIiLCJpYXQiOjE3MjMwODg3NTcuNzYzMDYxLCJuYmYiOjE3MjMwODg3NTcuNzYzMDY2LCJleHAiOjE3NTQ2MjQ3NTcuMzg2NTE5LCJzdWIiOiIxIiwic2NvcGVzIjpbXX0.IlPup6nzqliHv8_mFydLjM8XTQEv0-bfuUrAMTFQNA57LwmoutgPljtnWZlQumEUrkzVfBAWQHU-hRmuK4DjH2XacVIvZNpmTUIPSKq2ntTZ4c1AEtxzw1RPnoffMNoZ_HNJYQtDBZVEzfSpBWrS3TW7YrviZMjfuqemo75eARCTU-ZAGBXGWFWXHe0IZuL9BL66yKKtxG4veggM8Doi3D6xVcfFiZF71gfOGwPujTfUElZMzhH0P3su7vjHXnBHLORqTCyEk4Y2gJ2VJv4MdRaO91cuZJW0IPIT94xFpnlK5wvewQhyRGtkOoXKwed0jyItMXUThD1XdbduIlx3WeW4-Ivfk9LHUeEAGaXcZYthyMWPy9XCkFJs9pR5MI1OlIQQMZe_Pi7Ka0GxB_f_Jy3302Yi8nXXbcWql3Ha5daK8kJxDfBeUCCjSAzraXQOgwYr7UJVcNDVld6e_6Uf184vLe5KoSvQGi1HUiiDNnEOMtYWlDQDtsLCWAzHu64KT_K5EDBioOz834Zuds8UHxHPyOX5mZ0tJ9b5rcawo2geano1NJvXTBY9CZwQVeiaOnu9CaYDVaNE6CW7pn9LU-xuyGe4M8AHF5ccVpgywetTJXlSft1Dg--iUIHsgrYJgd7UqO8hsd0x4MbDgL4AtXwc2-6S6SiK14-Sv0Aarno"
)

func main() {
	fireflyClient := NewFireflyClient("http://192.168.50.32:18888/", FireflyToken)

	wechatFiles := ReadCSVFiles("wechat")
	alipayFiles := ReadCSVFiles("alipay")

	params := RequestParams{
		Limit: 10,
		Page:  1,
		Type:  "all",
	}

	response, err := getAccounts(FireflyToken, params)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Accounts: %+v\n", response)

	accountMap := make(map[string]string)
	// Iterate over the accounts data from the response
	for _, accountData := range response.Data {
		accountAttributes := accountData.Attributes
		accountMap[accountAttributes.Name] = accountData.ID
	}

	parseAndImportWeChat(wechatFiles, accountMap, fireflyClient)
	parseAndImportAlipay(alipayFiles, accountMap, fireflyClient)
}

func parseAndImportWeChat(files []string, accountMap map[string]string, fireflyClient *FireflyClient) {
	for _, file := range files {
		transactions, err := parseWeChatCSV(file)
		if err != nil {
			log.Fatalf("Failed to parse WeChat CSV: %v", err)
		}

		for _, t := range transactions {
			amount, _ := strconv.ParseFloat(strings.TrimPrefix(t.Amount, "¥"), 64)
			date, _ := time.Parse("2006-01-02 15:04:05", t.TransactionTime)

			transaction := FireflyTransaction{
				Transactions: []Transaction{
					{
						Type:            "withdrawal",
						Date:            date.Format(time.RFC3339),
						Amount:          strconv.FormatFloat(amount, 'f', 2, 64),
						Description:     t.Goods,
						CurrencyID:      "20", // Adjust as necessary
						CurrencyCode:    "CNY",
						SourceID:        getOrCreateAccount(fireflyClient, accountMap, t.PaymentMethod),
						DestinationID:   getOrCreateAccount(fireflyClient, accountMap, t.Counterparty),
						CategoryName:    "Uncategorized",
						SourceName:      t.PaymentMethod,
						DestinationName: t.Counterparty,
					},
				},
			}

			if err := fireflyClient.CreateTransaction(transaction); err != nil {
				log.Printf("Failed to create transaction: %v", err)
			}
		}
	}
}

func parseAndImportAlipay(files []string, accountMap map[string]string, fireflyClient *FireflyClient) {
	for _, file := range files {
		transactions, err := parseAlipayCSV(file)
		if err != nil {
			log.Fatalf("Failed to parse Alipay CSV: %v", err)
		}
		for _, t := range transactions {
			amount, _ := strconv.ParseFloat(strings.TrimPrefix(t.Amount, "¥"), 64)
			date, _ := time.Parse("2006-01-02 15:04:05", t.TransactionTime)

			transaction := FireflyTransaction{
				Transactions: []Transaction{
					{
						Type:            "withdrawal",
						Date:            date.Format(time.RFC3339),
						Amount:          strconv.FormatFloat(amount, 'f', 2, 64),
						Description:     t.Goods,
						CurrencyID:      "20", // Adjust as necessary
						CurrencyCode:    "CNY",
						SourceID:        getOrCreateAccount(fireflyClient, accountMap, t.Source),
						DestinationID:   getOrCreateAccount(fireflyClient, accountMap, t.Counterparty),
						CategoryName:    "Uncategorized",
						SourceName:      t.Source,
						DestinationName: t.Counterparty,
					},
				},
			}

			if err := fireflyClient.CreateTransaction(transaction); err != nil {
				log.Printf("Failed to create transaction: %v", err)
			}
		}
	}
}

func getOrCreateAccount(fireflyClient *FireflyClient, accountMap map[string]string, accountName string) string {
	if id, exists := accountMap[accountName]; exists {
		return id
	}
	newAccount := FireflyAccount{Name: accountName, Type: "expense"}
	account, err := fireflyClient.CreateAccount(newAccount)
	if err != nil {
		log.Fatalf("Failed to create account: %v", err)
	}

	accountMap[accountName] = account.ID
	return account.ID
}
