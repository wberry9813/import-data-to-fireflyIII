package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	FireflyToken = "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJhdWQiOiIxIiwianRpIjoiNTQ0ODJiYzhjY2ZiOTFkODhkNmY3ZjRiNTY1MGYwODIzOTM5ODUxMTdmOTI1MmJlZGJlZTRlYzlhODFlOGJiZWZhODY3M2U1ZjkyYmJmODIiLCJpYXQiOjE3MjMwODg3NTcuNzYzMDYxLCJuYmYiOjE3MjMwODg3NTcuNzYzMDY2LCJleHAiOjE3NTQ2MjQ3NTcuMzg2NTE5LCJzdWIiOiIxIiwic2NvcGVzIjpbXX0.IlPup6nzqliHv8_mFydLjM8XTQEv0-bfuUrAMTFQNA57LwmoutgPljtnWZlQumEUrkzVfBAWQHU-hRmuK4DjH2XacVIvZNpmTUIPSKq2ntTZ4c1AEtxzw1RPnoffMNoZ_HNJYQtDBZVEzfSpBWrS3TW7YrviZMjfuqemo75eARCTU-ZAGBXGWFWXHe0IZuL9BL66yKKtxG4veggM8Doi3D6xVcfFiZF71gfOGwPujTfUElZMzhH0P3su7vjHXnBHLORqTCyEk4Y2gJ2VJv4MdRaO91cuZJW0IPIT94xFpnlK5wvewQhyRGtkOoXKwed0jyItMXUThD1XdbduIlx3WeW4-Ivfk9LHUeEAGaXcZYthyMWPy9XCkFJs9pR5MI1OlIQQMZe_Pi7Ka0GxB_f_Jy3302Yi8nXXbcWql3Ha5daK8kJxDfBeUCCjSAzraXQOgwYr7UJVcNDVld6e_6Uf184vLe5KoSvQGi1HUiiDNnEOMtYWlDQDtsLCWAzHu64KT_K5EDBioOz834Zuds8UHxHPyOX5mZ0tJ9b5rcawo2geano1NJvXTBY9CZwQVeiaOnu9CaYDVaNE6CW7pn9LU-xuyGe4M8AHF5ccVpgywetTJXlSft1Dg--iUIHsgrYJgd7UqO8hsd0x4MbDgL4AtXwc2-6S6SiK14-Sv0Aarno"
)

func main() {
	fireflyClient := NewFireflyClient("http://192.168.50.32:18888/", FireflyToken)

	// wechatFiles := ReadCSVFiles("wechat")
	// alipayFiles := ReadCSVFiles("alipay")
	icbcFiles := ReadCSVFiles("icbc")

	params := RequestParams{
		Limit: 1000,
		Page:  1,
		Type:  "all",
	}

	response, err := getAccounts(FireflyToken, params)
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

	//parseAndImportWeChat(wechatFiles, accountMap, fireflyClient)
	//parseAndImportAlipay(alipayFiles, accountMap, fireflyClient)
	parseAndImportICBC(icbcFiles, accountMap, fireflyClient)

	if len(transactionErrors) > 0 {
		err := saveErrorsToCSVAndLog(transactionErrors)
		if err != nil {
			log.Fatalf("Failed to save errors: %v", err)
		}
	} else {
		log.Println("All transactions processed successfully.")
	}

	// 执行完所有业务逻辑后，移动文件
	err = moveFilesToBackup("wechat", "backup")
	if err != nil {
		fmt.Println("移动微信文件失败:", err)
	}

	err = moveFilesToBackup("alipay", "backup")
	if err != nil {
		fmt.Println("移动支付宝文件失败:", err)
	}

	fmt.Println("文件已成功移动到备份文件夹。")

	// if len(transactionErrors) > 0 {
	// 	log.Println("Failed transactions:")
	// 	for _, te := range transactionErrors {
	// 		log.Printf("CSV Data: %s, Error: %v\n", te.SourceData, te.ErrorInfo)
	// 	}
	// } else {
	// 	log.Println("All transactions processed successfully.")
	// }
}

func parseAndImportWeChat(files []string, accountMap map[string]map[string]string, fireflyClient *FireflyClient) {
	for _, file := range files {
		transactions, err := parseWeChatCSV(file)
		if err != nil {
			log.Fatalf("Failed to parse WeChat CSV: %v", err)
		}
		if len(transactions) == 0 {
			continue
		}

		for _, t := range transactions {
			amount, _ := strconv.ParseFloat(strings.TrimPrefix(t.Amount, "¥"), 64)

			location, _ := time.LoadLocation("Asia/Shanghai")
			date, _ := time.ParseInLocation("2006-01-02 15:04:05", t.TransactionTime, location)

			log.Printf("create transaction %v", t)

			sourceId, err := getOrCreateAccount(fireflyClient, accountMap, t.PaymentMethod, "asset")
			if err != nil {
				log.Printf("Failed to get or create asset account: %v", err)

			}

			transaction := FireflyTransaction{}
			if t.InOrOut == "支出" {
				destinationId, err := getOrCreateAccount(fireflyClient, accountMap, t.Counterparty, "expense")
				if err != nil {
					log.Printf("Failed to get or create expense account: %v", err)
				}

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
							CategoryName:    t.TransactionType,
							SourceName:      t.PaymentMethod,
							DestinationName: t.Counterparty,
							Tags:            []string{"WeChat"},
						},
					},
				}
			} else {
				destinationId, err := getOrCreateAccount(fireflyClient, accountMap, t.Counterparty, "revenue")
				if err != nil {
					log.Printf("Failed to get or create revenue account: %v", err)
				}

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
							Tags:            []string{"WeChat"},
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

func parseAndImportAlipay(files []string, accountMap map[string]map[string]string, fireflyClient *FireflyClient) {
	for _, file := range files {
		transactions, err := parseAlipayCSV(file)
		if err != nil {
			log.Fatalf("Failed to parse Alipay CSV: %v", err)
		}
		if len(transactions) == 0 {
			continue
		}

		var filteredTransactions []AlipayTransaction

		for _, t := range transactions {
			// 检查 InOrOut 字段是否为 "不计收支"
			if t.InOrOut != "不计收支" {
				// 如果不是 "不计收支"，则将交易添加到 filteredTransactions 中
				filteredTransactions = append(filteredTransactions, t)
			}
		}

		for _, t := range filteredTransactions {
			amount, _ := strconv.ParseFloat(strings.TrimPrefix(t.Amount, "¥"), 64)

			location, _ := time.LoadLocation("Asia/Shanghai")
			date, _ := time.ParseInLocation("2006-01-02 15:04:05", t.TransactionTime, location)

			log.Printf("create transaction %v", t)

			transaction := FireflyTransaction{}
			if t.InOrOut == "支出" {
				if strings.Contains(t.PaymentMethod, "花呗") {
					t.PaymentMethod = "花呗"
				}
				sourceId, err := getOrCreateAccount(fireflyClient, accountMap, t.PaymentMethod, "asset")
				if err != nil {
					log.Printf("Failed to get or create asset account: %v", err)
				}

				destinationId, err := getOrCreateAccount(fireflyClient, accountMap, t.Counterparty, "expense")
				if err != nil {
					log.Printf("Failed to get or create expense account: %v", err)
				}
				transaction = FireflyTransaction{
					Transactions: []Transaction{
						{
							Type:            "withdrawal",
							Date:            date.UTC().Format(time.RFC3339),
							Amount:          strconv.FormatFloat(amount, 'f', 2, 64),
							Description:     t.ProductDescription,
							CurrencyID:      "20", // Adjust as necessary
							CurrencyCode:    "CNY",
							SourceID:        sourceId,
							DestinationID:   destinationId,
							CategoryName:    t.TransactionCategory,
							SourceName:      t.PaymentMethod,
							DestinationName: t.Counterparty,
							Tags:            []string{"Alipay"},
						},
					},
				}
			} else if t.InOrOut == "收入" {
				if t.PaymentMethod == "" {
					t.PaymentMethod = "支付宝余额"
				}
				sourceId, err := getOrCreateAccount(fireflyClient, accountMap, t.PaymentMethod, "asset")
				if err != nil {
					log.Printf("Failed to get or create asset account: %v", err)
				}

				destinationId, err := getOrCreateAccount(fireflyClient, accountMap, t.Counterparty, "revenue")
				if err != nil {
					log.Printf("Failed to get or create revenue account: %v", err)
				}
				transaction = FireflyTransaction{
					Transactions: []Transaction{
						{
							Type:            "deposit",
							Date:            date.Format(time.RFC3339),
							Amount:          strconv.FormatFloat(amount, 'f', 2, 64),
							Description:     t.ProductDescription,
							CurrencyID:      "20", // Adjust as necessary
							CurrencyCode:    "CNY",
							SourceID:        destinationId,
							DestinationID:   sourceId,
							CategoryName:    t.TransactionCategory,
							SourceName:      t.Counterparty,
							DestinationName: t.PaymentMethod,
							Tags:            []string{"Alipay"},
						},
					},
				}
			} else {
				recordError(t, nil)
				continue
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

func parseAndImportICBC(files []string, accountMap map[string]map[string]string, fireflyClient *FireflyClient) {
	for _, file := range files {
		transactions, err := parseICBCCSV(file)
		if err != nil {
			log.Fatalf("Failed to parse WeChat CSV: %v", err)
		}
		if len(transactions) == 0 {
			continue
		}

		for _, t := range transactions {
			amount, _ := strconv.ParseFloat(strings.TrimPrefix(t.IncomeAmount, "+"), 64)

			location, _ := time.LoadLocation("Asia/Shanghai")
			date, _ := time.ParseInLocation("2006-01-02 15:04:05", t.TransactionTime, location)

			log.Printf("create transaction %v", t)

			PaymentMethod := "工商银行储蓄卡(3258)"
			sourceId, err := getOrCreateAccount(fireflyClient, accountMap, PaymentMethod, "asset")
			if err != nil {
				log.Printf("Failed to get or create asset account: %v", err)

			}

			transaction := FireflyTransaction{}

			destinationId, err := getOrCreateAccount(fireflyClient, accountMap, t.Counterparty, "revenue")
			if err != nil {
				log.Printf("Failed to get or create revenue account: %v", err)
			}

			transaction = FireflyTransaction{
				Transactions: []Transaction{
					{
						Type:            "deposit",
						Date:            date.Format(time.RFC3339),
						Amount:          strconv.FormatFloat(amount, 'f', 2, 64),
						Description:     t.Summary,
						CurrencyID:      "20", // Adjust as necessary
						CurrencyCode:    "CNY",
						SourceID:        destinationId,
						DestinationID:   sourceId,
						CategoryName:    t.Summary,
						SourceName:      t.Counterparty,
						DestinationName: PaymentMethod,
						Tags:            []string{t.Summary},
					},
				},
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

func getOrCreateAccount(fireflyClient *FireflyClient, accountMap map[string]map[string]string, accountName string, accountType string) (string, error) {
	accountName = strings.TrimSpace(accountName)
	if id, exists := accountMap[accountName][accountType]; exists {
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

	if accountMap[accountName] == nil {
		accountMap[accountName] = make(map[string]string)
	}
	accountMap[accountName][accountType] = account.Data.ID
	return account.Data.ID, nil
}

var weChatTransactionErrors []WeChatTransaction
var alipayTransactionErrors []AlipayTransaction

func recordError(t interface{}, err error) {
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	switch v := t.(type) {
	case WeChatTransaction:
		// 记录 WeChatTransaction 错误
		transactionErrors = append(transactionErrors, TransactionError{
			WeChatTransactionError: v,
			ErrorInfo:              errStr,
		})
	case AlipayTransaction:
		// 记录 AlipayTransaction 错误
		transactionErrors = append(transactionErrors, TransactionError{
			AlipayTransactionError: v,
			ErrorInfo:              errStr,
		})
	}
}

var transactionErrors []TransactionError

func saveErrorsToCSVAndLog(transactionErrors []TransactionError) error {
	timestamp := time.Now().Format("20060102_150405")

	logFilePath := fmt.Sprintf("error/transaction_errors_%s.log", timestamp)
	// 获取目录路径
	logDir := filepath.Dir(logFilePath)

	// 如果目录不存在，创建目录
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		log.Fatalf("无法创建日志目录: %v", err)
	}
	// 创建日志文件
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer logFile.Close()

	// 设置日志输出到文件
	logger := log.New(logFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	// 创建带有时间戳的文件名
	wechatFileName := fmt.Sprintf("error/wechat_errors_%s.csv", timestamp)
	alipayFileName := fmt.Sprintf("error/alipay_errors_%s.csv", timestamp)

	// 保存WeChatTransactionErrors到CSV文件
	wechatFile, err := os.Create(wechatFileName)
	if err != nil {
		return err
	}
	defer wechatFile.Close()

	wechatWriter := csv.NewWriter(wechatFile)
	defer wechatWriter.Flush()

	// 保存AlipayTransactionErrors到CSV文件
	alipayFile, err := os.Create(alipayFileName)
	if err != nil {
		return err
	}
	defer alipayFile.Close()

	alipayWriter := csv.NewWriter(alipayFile)
	defer alipayWriter.Flush()

	// 遍历 transactionErrors 列表，分别保存到CSV文件，并输出到日志
	for _, tErr := range transactionErrors {
		if tErr.WeChatTransactionError != (WeChatTransaction{}) {
			// 将 WeChatTransactionError 保存到 CSV
			record := []string{
				tErr.WeChatTransactionError.TransactionTime,
				tErr.WeChatTransactionError.TransactionType,
				tErr.WeChatTransactionError.Counterparty,
				tErr.WeChatTransactionError.Goods,
				tErr.WeChatTransactionError.InOrOut,
				tErr.WeChatTransactionError.Amount,
				tErr.WeChatTransactionError.PaymentMethod,
				tErr.WeChatTransactionError.Status,
				tErr.WeChatTransactionError.TransactionID,
				tErr.WeChatTransactionError.MerchantOrderID,
				tErr.WeChatTransactionError.Remark,
				tErr.ErrorInfo, // 错误信息
			}
			if err := wechatWriter.Write(record); err != nil {
				return err
			}

			// 记录日志
			logger.Printf("WeChat Transaction Error: %+v, Error: %s\n", tErr.WeChatTransactionError, tErr.ErrorInfo)
		}

		if tErr.AlipayTransactionError != (AlipayTransaction{}) {
			// 将 AlipayTransactionError 保存到 CSV
			record := []string{
				tErr.AlipayTransactionError.TransactionTime,
				tErr.AlipayTransactionError.TransactionCategory,
				tErr.AlipayTransactionError.Counterparty,
				tErr.AlipayTransactionError.CounterpartyAccount,
				tErr.AlipayTransactionError.ProductDescription,
				tErr.AlipayTransactionError.InOrOut,
				tErr.AlipayTransactionError.Amount,
				tErr.AlipayTransactionError.PaymentMethod,
				tErr.AlipayTransactionError.TransactionStatus,
				tErr.AlipayTransactionError.TransactionID,
				tErr.AlipayTransactionError.MerchantOrderID,
				tErr.AlipayTransactionError.Notes,
				tErr.ErrorInfo,
			}
			if err := alipayWriter.Write(record); err != nil {
				return err
			}

			// 记录日志
			logger.Printf("Alipay Transaction Error: %+v, Error: %s\n", tErr.AlipayTransactionError, tErr.ErrorInfo)
		}
	}

	return nil
}

func moveFilesToBackup(srcFolder, backupFolder string) error {
	// 确保备份文件夹存在
	if _, err := os.Stat(backupFolder); os.IsNotExist(err) {
		err := os.Mkdir(backupFolder, os.ModePerm)
		if err != nil {
			return fmt.Errorf("创建备份文件夹失败: %v", err)
		}
	}

	// 读取源文件夹中的所有文件
	files, err := filepath.Glob(filepath.Join(srcFolder, "*"))
	if err != nil {
		return fmt.Errorf("读取源文件夹失败: %v", err)
	}

	for _, file := range files {
		// 获取文件的基础名称
		fileName := filepath.Base(file)

		// 构造目标路径
		destPath := filepath.Join(backupFolder, fileName)

		// 移动文件
		err := moveFile(file, destPath)
		if err != nil {
			return fmt.Errorf("移动文件失败: %v", err)
		}
	}

	return nil
}

// 移动文件的辅助函数
func moveFile(srcPath, destPath string) error {
	// 打开源文件
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %v", err)
	}
	defer srcFile.Close()

	// 创建目标文件
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %v", err)
	}
	defer destFile.Close()

	// 复制文件内容
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("复制文件内容失败: %v", err)
	}

	// 删除源文件
	err = os.Remove(srcPath)
	if err != nil {
		return fmt.Errorf("删除源文件失败: %v", err)
	}

	return nil
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
