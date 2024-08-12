package main

import (
	"os"

	"github.com/gocarina/gocsv"
)

func parseWeChatCSV(file string) ([]WeChatTransaction, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var transactions []WeChatTransaction
	if err := gocsv.UnmarshalFile(f, &transactions); err != nil {
		return nil, err
	}

	return transactions, nil
}

func parseAlipayCSV(file string) ([]AlipayTransaction, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var transactions []AlipayTransaction
	if err := gocsv.UnmarshalFile(f, &transactions); err != nil {
		return nil, err
	}

	return transactions, nil
}
