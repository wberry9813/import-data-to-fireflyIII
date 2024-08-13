package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	FireflyToken = "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJhdWQiOiIxIiwianRpIjoiNTQ0ODJiYzhjY2ZiOTFkODhkNmY3ZjRiNTY1MGYwODIzOTM5ODUxMTdmOTI1MmJlZGJlZTRlYzlhODFlOGJiZWZhODY3M2U1ZjkyYmJmODIiLCJpYXQiOjE3MjMwODg3NTcuNzYzMDYxLCJuYmYiOjE3MjMwODg3NTcuNzYzMDY2LCJleHAiOjE3NTQ2MjQ3NTcuMzg2NTE5LCJzdWIiOiIxIiwic2NvcGVzIjpbXX0.IlPup6nzqliHv8_mFydLjM8XTQEv0-bfuUrAMTFQNA57LwmoutgPljtnWZlQumEUrkzVfBAWQHU-hRmuK4DjH2XacVIvZNpmTUIPSKq2ntTZ4c1AEtxzw1RPnoffMNoZ_HNJYQtDBZVEzfSpBWrS3TW7YrviZMjfuqemo75eARCTU-ZAGBXGWFWXHe0IZuL9BL66yKKtxG4veggM8Doi3D6xVcfFiZF71gfOGwPujTfUElZMzhH0P3su7vjHXnBHLORqTCyEk4Y2gJ2VJv4MdRaO91cuZJW0IPIT94xFpnlK5wvewQhyRGtkOoXKwed0jyItMXUThD1XdbduIlx3WeW4-Ivfk9LHUeEAGaXcZYthyMWPy9XCkFJs9pR5MI1OlIQQMZe_Pi7Ka0GxB_f_Jy3302Yi8nXXbcWql3Ha5daK8kJxDfBeUCCjSAzraXQOgwYr7UJVcNDVld6e_6Uf184vLe5KoSvQGi1HUiiDNnEOMtYWlDQDtsLCWAzHu64KT_K5EDBioOz834Zuds8UHxHPyOX5mZ0tJ9b5rcawo2geano1NJvXTBY9CZwQVeiaOnu9CaYDVaNE6CW7pn9LU-xuyGe4M8AHF5ccVpgywetTJXlSft1Dg--iUIHsgrYJgd7UqO8hsd0x4MbDgL4AtXwc2-6S6SiK14-Sv0Aarno"
)

var transactionErrors []TransactionError

func main() {
	fireflyClient := NewFireflyClient("http://192.168.50.32:18888/", FireflyToken)

	wechatFiles := ReadCSVFiles("wechat")
	//alipayFiles := ReadCSVFiles("alipay")

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
	//parseAndImportAlipay(alipayFiles, accountMap, fireflyClient)

	if len(transactionErrors) > 0 {
		log.Println("Failed transactions:")
		for _, te := range transactionErrors {
			log.Printf("CSV Data: %s, Error: %v\n", te.SourceData, te.ErrorInfo)
		}
	} else {
		log.Println("All transactions processed successfully.")
	}
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

			log.Printf("create transaction %v", t)

			sourceId, err := getOrCreateAccount(fireflyClient, accountMap, t.PaymentMethod, "asset")
			if err != nil {
				log.Printf("Failed to get or create asset account: %v", err)

			}
			destinationId, err := getOrCreateAccount(fireflyClient, accountMap, t.Counterparty, "expense")
			if err != nil {
				log.Printf("Failed to get or create expense account: %v", err)
			}

			transaction := FireflyTransaction{}
			if t.InOrOut == "支出" {
				transaction = FireflyTransaction{
					Transactions: []Transaction{
						{
							Type:            "withdrawal",
							Date:            date.Format(time.RFC3339),
							Amount:          strconv.FormatFloat(amount, 'f', 2, 64),
							Description:     t.Goods,
							CurrencyID:      "20", // Adjust as necessary
							CurrencyCode:    "CNY",
							SourceID:        sourceId,
							DestinationID:   destinationId,
							CategoryName:    "Uncategorized",
							SourceName:      t.PaymentMethod,
							DestinationName: t.Counterparty,
						},
					},
				}
			} else {
				transaction = FireflyTransaction{
					Transactions: []Transaction{
						{
							Type:            "deposit",
							Date:            date.Format(time.RFC3339),
							Amount:          strconv.FormatFloat(amount, 'f', 2, 64),
							Description:     t.Goods,
							CurrencyID:      "20", // Adjust as necessary
							CurrencyCode:    "CNY",
							SourceID:        destinationId,
							DestinationID:   sourceId,
							CategoryName:    "Uncategorized",
							SourceName:      t.PaymentMethod,
							DestinationName: t.Counterparty,
						},
					},
				}
			}

			jsonData, err := json.Marshal(transaction)
			if err != nil {
				log.Fatalf("Error converting transaction to JSON: %v", err)
			}

			fmt.Println(string(jsonData))

			if err := fireflyClient.CreateTransaction(transaction); err != nil {
				log.Printf("Failed to create transaction: %v", err)
				recordError(t, err)
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

			log.Printf("create transaction %v", t)

			sourceId, err := getOrCreateAccount(fireflyClient, accountMap, t.Source, "asset")
			if err != nil {
				log.Printf("Failed to get or create asset account: %v", err)

			}
			destinationId, err := getOrCreateAccount(fireflyClient, accountMap, t.Counterparty, "expense")
			if err != nil {
				log.Printf("Failed to get or create expense account: %v", err)
			}

			transaction := FireflyTransaction{
				Transactions: []Transaction{
					{
						Type:            "withdrawal",
						Date:            date.Format(time.RFC3339),
						Amount:          strconv.FormatFloat(amount, 'f', 2, 64),
						Description:     t.Goods,
						CurrencyID:      "20", // Adjust as necessary
						CurrencyCode:    "CNY",
						SourceID:        sourceId,
						DestinationID:   destinationId,
						CategoryName:    "Uncategorized",
						SourceName:      t.Source,
						DestinationName: t.Counterparty,
					},
				},
			}

			if err := fireflyClient.CreateTransaction(transaction); err != nil {
				log.Printf("Failed to create transaction: %v", err)
				recordError(t, err)
			}
		}
	}
}

func getOrCreateAccount(fireflyClient *FireflyClient, accountMap map[string]string, accountName string, accountType string) (string, error) {
	if id, exists := accountMap[accountName]; exists {
		return id, nil
	}
	accountRole := "defaultAsset"
	if accountType != "asset" {
		accountRole = ""
	}
	newAccount := FireflyAccountRequest{
		Name:            accountName,
		Type:            accountType,
		CurrencyID:      "20", // Adjust as necessary
		CurrencyCode:    "CNY",
		Active:          true,
		IncludeNetWorth: true,
		AccountRole:     accountRole,
	}

	account, err := fireflyClient.CreateAccount(newAccount)
	if err != nil {
		log.Printf("Failed to create account: %v", err)
		return "", err
	}

	accountMap[accountName] = account.Data.ID
	return account.Data.ID, nil
}

func recordError(t interface{}, err error) {
	var sourceData string
	switch v := t.(type) {
	case WeChatTransaction:
		sourceData = fmt.Sprintf(
			"交易时间: %s, 交易类型: %s, 交易对方: %s, 商品: %s, 收/支: %s, 金额: %s, 支付方式: %s, 当前状态: %s, 交易单号: %s, 商户单号: %s, 备注: %s",
			v.TransactionTime, v.TransactionType, v.Counterparty, v.Goods, v.InOrOut, v.Amount,
			v.PaymentMethod, v.Status, v.TransactionID, v.MerchantOrderID, v.Remark)
	case AlipayTransaction:
		sourceData = fmt.Sprintf(
			"交易号: %s, 商家订单号: %s, 交易创建时间: %s, 付款时间: %s, 最近修改时间: %s, 交易来源地: %s, 类型: %s, 交易对方: %s, 商品名称: %s, 金额: %s, 收/支: %s, 交易状态: %s, 服务费: %s, 成功退款: %s, 备注: %s, 资金状态: %s",
			v.TransactionID, v.MerchantOrderID, v.TransactionTime, v.PaymentTime, v.LastModifiedTime,
			v.Source, v.Type, v.Counterparty, v.Goods, v.Amount, v.InOrOut, v.Status, v.ServiceFee,
			v.SuccessRefund, v.Remark, v.FundStatus)
	default:
		sourceData = "未知的交易类型"
	}

	transactionErrors = append(transactionErrors, TransactionError{
		SourceData: sourceData,
		ErrorInfo:  err.Error(),
	})
}
