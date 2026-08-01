package main

import (
	"encoding/json"
	"fmt"
	"testing/fstest"

	"github.com/wsnacj/agentx-go/extensions/astock/contracts"
	"github.com/wsnacj/agentx-go/runtime/assetfs"
)

func main() {
	result, err := runConsumer()
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}

func runConsumer() (string, error) {
	provider, err := assetfs.New("astock.contract.fixture", fstest.MapFS{
		"quote.json": &fstest.MapFile{Data: []byte(`{
			"adapter_status":"ok",
			"subject":{"stock_code":"000001","market":"sz","verified":true},
			"readiness":{"answer_ready":true,"requested_fields_ready":true}
		}`)},
	})
	if err != nil {
		return "", err
	}
	content, err := provider.ReadFile("quote.json")
	if err != nil {
		return "", err
	}
	var payload contracts.QuotePayload
	if err := json.Unmarshal(content, &payload); err != nil {
		return "", err
	}
	code, market, ok := contracts.NormalizeAStockCode("sz" + payload.Subject.StockCode)
	if !ok {
		return "", fmt.Errorf("normalize A-stock code %q", payload.Subject.StockCode)
	}
	return fmt.Sprintf(
		"agentx-extension-astock-contract-ok:%s:%s:%t",
		code,
		market,
		payload.Readiness.AnswerReady,
	), nil
}
