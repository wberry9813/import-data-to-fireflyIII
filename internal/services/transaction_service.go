package services

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	. "import-data-to-fireflyIII/internal/api"
	. "import-data-to-fireflyIII/internal/models"
	"import-data-to-fireflyIII/internal/parsers"
	"import-data-to-fireflyIII/pkg/utils"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type TransactionService struct {
	fireflyClient *FireflyClient
}

func NewTransactionService(fireflyClient *FireflyClient) *TransactionService {
	return &TransactionService{fireflyClient: fireflyClient}
}

func (s *TransactionService) ParseAndImportWeChat(files []string, accountMap map[string]map[string]string) {
	for _, file := range files {
		transactions, err := parsers.ParseWeChatCSV(file)
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

			sourceId, err := s.getOrCreateAccount(accountMap, t.PaymentMethod, "asset")
			if err != nil {
				log.Printf("Failed to get or create asset account: %v", err)
			}

			transaction := FireflyTransaction{}
			if t.InOrOut == "支出" {
				destinationId, err := s.getOrCreateAccount(accountMap, t.Counterparty, "expense")
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
				destinationId, err := s.getOrCreateAccount(accountMap, t.Counterparty, "revenue")
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

			if err := s.fireflyClient.CreateTransaction(transaction); err != nil {
				log.Printf("Failed to create transaction: %v", err)
				recordError(t, err)
			}
		}
	}
}

func (s *TransactionService) ParseAndImportAlipay(files []string, accountMap map[string]map[string]string) {
	for _, file := range files {
		transactions, err := parsers.ParseAlipayCSV(file)
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
				sourceId, err := s.getOrCreateAccount(accountMap, t.PaymentMethod, "asset")
				if err != nil {
					log.Printf("Failed to get or create asset account: %v", err)
				}

				destinationId, err := s.getOrCreateAccount(accountMap, t.Counterparty, "expense")
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
				sourceId, err := s.getOrCreateAccount(accountMap, t.PaymentMethod, "asset")
				if err != nil {
					log.Printf("Failed to get or create asset account: %v", err)
				}

				destinationId, err := s.getOrCreateAccount(accountMap, t.Counterparty, "revenue")
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

			if err := s.fireflyClient.CreateTransaction(transaction); err != nil {
				log.Printf("Failed to create transaction: %v", err)
				recordError(t, err)
			}
		}
	}
}

func (s *TransactionService) ParseAndImportICBC(files []string, accountMap map[string]map[string]string) {
	for _, file := range files {
		transactions, err := parsers.ParseICBCCSV(file)
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
			sourceId, err := s.getOrCreateAccount(accountMap, PaymentMethod, "asset")
			if err != nil {
				log.Printf("Failed to get or create asset account: %v", err)

			}

			transaction := FireflyTransaction{}

			destinationId, err := s.getOrCreateAccount(accountMap, t.Counterparty, "revenue")
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

			if err := s.fireflyClient.CreateTransaction(transaction); err != nil {
				log.Printf("Failed to create transaction: %v", err)
				recordError(t, err)
			}
		}
	}
}

func (s *TransactionService) getOrCreateAccount(accountMap map[string]map[string]string, accountName string, accountType string) (string, error) {
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

	account, err := s.fireflyClient.CreateAccount(newAccount)
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

func (s *TransactionService) SaveLogAndMoveFiles() {
	if len(transactionErrors) > 0 {
		err := saveErrorsToCSVAndLog(transactionErrors)
		if err != nil {
			log.Fatalf("Failed to save errors: %v", err)
		}
	} else {
		log.Println("All transactions processed successfully.")
	}

	// 执行完所有业务逻辑后，移动文件
	err := utils.MoveFilesToBackup("billing/wechat", "backup")
	if err != nil {
		fmt.Println("移动微信文件失败:", err)
	}

	err = utils.MoveFilesToBackup("billing/alipay", "backup")
	if err != nil {
		fmt.Println("移动支付宝文件失败:", err)
	}

	fmt.Println("文件已成功移动到备份文件夹。")
}
