package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type Config struct {
	Key string `json:"Key"`
}

type Stocks struct {
	Symbol        string      `json:"symbol"`
	Name          string      `json:"name"`
	Price         interface{} `json:"price"`
	PercentChange interface{} `json:"percent_change"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

func loadApiKey(configPath string) string {
	data, err := os.ReadFile(configPath) 
	if err != nil {
		log.Fatalf("failed to read config file:%v", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("Failed to parse config file:%v", err)
	}
	return cfg.Key
}

func main() {
	apiKey := loadApiKey(".apiConfig")
	tickets := []string{
		"IBM",
		"WMT",
		"MMM",
		"INTC",
		"AXP",
	}

	symbols := strings.Join(tickets, ",")
	url := fmt.Sprintf("https://api.twelvedata.com/quote?symbol=%s&apikey=%s", symbols, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("HTTP req failed:", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Failed to read response body:", err)
	}

	fmt.Println("API Response:", string(body))

	var errorResp ErrorResponse
	if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Code != 0 {
		log.Fatalf("API Error %d: %s", errorResp.Code, errorResp.Message)
	}

	var data map[string]Stocks
	if err := json.Unmarshal(body, &data); err != nil {
		log.Fatal("JSON decode err:", err)
	}

	if len(data) == 0 {
		log.Fatal("No stock data received")
	}

	file, err := os.Create("stocklist.csv")
	if err != nil {
		log.Fatalln("Error creating the file", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{
		"Symbol",
		"Company",
		"Price",
		"Change %",
	}
	writer.Write(headers)

	for symbol, stock := range data {
		priceStr := fmt.Sprintf("%v", stock.Price)
		changeStr := fmt.Sprintf("%v", stock.PercentChange)

		fmt.Printf("%s (%s): $%s (%s%%)\n", stock.Name, symbol, priceStr, changeStr)
		writer.Write([]string{
			stock.Symbol,
			stock.Name,
			priceStr,
			changeStr,
		})
	}

	fmt.Printf("Successfully processed %d stocks\n", len(data))
}
