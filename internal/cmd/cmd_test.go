package cmd

import "testing"

func TestStockTradeBody(t *testing.T) {
	body, err := stockTradeBody("AAPL", "1", "")
	if err != nil || body["symbol"] != "AAPL" || body["shares"] != "1" {
		t.Fatalf("%v %v", body, err)
	}
	if _, err := stockTradeBody("AAPL", "1", "10"); err == nil {
		t.Fatal("expected error for both shares and amount")
	}
	if _, err := stockTradeBody("AAPL", "", ""); err == nil {
		t.Fatal("expected error for missing quantity")
	}
}
