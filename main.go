package main

import (
	"fmt"
	"import-data-to-fireflyIII/configs"
	"import-data-to-fireflyIII/internal/api"
	. "import-data-to-fireflyIII/internal/models"
	"import-data-to-fireflyIII/internal/services"
	"import-data-to-fireflyIII/pkg/utils"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	config := configs.LoadConfig()

	fireflyClient := api.NewFireflyClient(config.FireflyBaseURL, config.FireflyToken)
	transactionService := services.NewTransactionService(fireflyClient)

	wechatFiles, err := utils.ReadCSVFiles("billing/wechat")
	if err != nil {
		fmt.Println("Failed to read WeChat CSV files: ", err)
	}
	alipayFiles, err := utils.ReadCSVFiles("billing/alipay")
	if err != nil {
		fmt.Println("Failed to read Alipay CSV files:", err)
	}
	icbcFiles, err := utils.ReadCSVFiles("billing/icbc")
	if err != nil {
		fmt.Println("Failed to read ICBC CSV files: ", err)
	}

	if condition := len(wechatFiles) == 0 && len(alipayFiles) == 0 && len(icbcFiles) == 0; condition {
		fmt.Println("No CSV files found in the billing directories.")
		return
	}

	params := RequestParams{
		Limit: 1000,
		Page:  1,
		Type:  "all",
	}

	response, err := fireflyClient.GetAccounts(params)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Accounts: %+v\n", response)

	//accountMap := make(map[string]string)
	accountMap := make(map[string]map[string]string)
	// Iterate over the accounts data from the response
	for _, accountData := range response.Data {
		accountAttributes := accountData.Attributes
		if accountMap[accountAttributes.Name] == nil {
			accountMap[accountAttributes.Name] = make(map[string]string)
		}
		accountMap[accountAttributes.Name][accountAttributes.Type] = accountData.ID
	}

	transactionService.ParseAndImportWeChat(wechatFiles, accountMap)
	transactionService.ParseAndImportAlipay(alipayFiles, accountMap)
	transactionService.ParseAndImportICBC(icbcFiles, accountMap)

	transactionService.SaveLogAndMoveFiles()

	// if len(transactionErrors) > 0 {
	// 	log.Println("Failed transactions:")
	// 	for _, te := range transactionErrors {
	// 		log.Printf("CSV Data: %s, Error: %v\n", te.SourceData, te.ErrorInfo)
	// 	}
	// } else {
	// 	log.Println("All transactions processed successfully.")
	// }
}

// func recordError(t interface{}, err error) {
// 	var sourceData string
// 	switch v := t.(type) {
// 	case WeChatTransaction:
// 		sourceData = fmt.Sprintf(
// 			"交易时间: %s, 交易类型: %s, 交易对方: %s, 商品: %s, 收/支: %s, 金额: %s, 支付方式: %s, 当前状态: %s, 交易单号: %s, 商户单号: %s, 备注: %s",
// 			v.TransactionTime, v.TransactionType, v.Counterparty, v.Goods, v.InOrOut, v.Amount,
// 			v.PaymentMethod, v.Status, v.TransactionID, v.MerchantOrderID, v.Remark)
// 	case AlipayTransaction:
// 		sourceData = fmt.Sprintf(
// 			"交易时间: %s, 交易分类: %s, 交易对方: %s, 对方账号: %s, 商品说明: %s, 收/支: %s, 金额: %s, 收/付款方式: %s, 交易状态: %s, 交易订单号: %s, 商家订单号: %s, 备注: %s",
// 			v.TransactionTime, v.TransactionCategory, v.Counterparty, v.CounterpartyAccount, v.ProductDescription, v.InOrOut, v.Amount, v.PaymentMethod, v.TransactionStatus, v.TransactionID, v.MerchantOrderID, v.Notes,
// 		)
// 	default:
// 		sourceData = "未知的交易类型"
// 	}

// 	transactionErrors = append(transactionErrors, TransactionError{
// 		SourceData: sourceData,
// 		ErrorInfo:  err.Error(),
// 	})
// }
