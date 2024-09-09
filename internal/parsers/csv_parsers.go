package parsers

import (
	"bytes"
	"os"

	"io"

	"import-data-to-fireflyIII/internal/models"

	"github.com/gocarina/gocsv"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func ParseWeChatCSV(file string) ([]models.WeChatTransaction, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var transactions []models.WeChatTransaction
	if err := gocsv.UnmarshalFile(f, &transactions); err != nil {
		return nil, err
	}

	return transactions, nil
}

func ParseAlipayCSV(file string) ([]models.AlipayTransaction, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// 将 GB2312 编码转换为 UTF-8
	reader := transform.NewReader(f, simplifiedchinese.GB18030.NewDecoder())
	utf8Data, err := io.ReadAll(io.Reader(reader))
	if err != nil {
		return nil, err
	}

	var transactions []models.AlipayTransaction
	if err := gocsv.Unmarshal(bytes.NewReader(utf8Data), &transactions); err != nil {
		return nil, err
	}

	return transactions, nil
}

func ParseICBCCSV(file string) ([]models.ICBCTransaction, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var transactions []models.ICBCTransaction
	if err := gocsv.UnmarshalFile(f, &transactions); err != nil {
		return nil, err
	}

	return transactions, nil
}
