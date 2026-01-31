package llm

import (
	"testing"
	"time"
)

func TestParseAmount(t *testing.T) {
	cases := []struct {
		input    string
		expected float64
	}{
		{"กินข้าว 50 บาท", 50},
		{"ค่าไฟ 1,234.50 บาท", 1234.50},
		{"ยอด 96.00", 96.00},
	}

	for _, tc := range cases {
		if got := parseAmount(tc.input); got != tc.expected {
			t.Fatalf("parseAmount(%q) = %v, want %v", tc.input, got, tc.expected)
		}
	}
}

func TestParseDate(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"วันที่ 2026-01-30", "2026-01-30"},
		{"30/01/2026", "2026-01-30"},
		{"30/01/2569", "2026-01-30"},
	}

	for _, tc := range cases {
		if got := parseDate(tc.input); got != tc.expected {
			t.Fatalf("parseDate(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestParseTextRegexAddTransaction(t *testing.T) {
	msg := "วันนี้กินข้าว 50 บาท เงินสด"
	resp := parseTextRegex(msg)

	if resp.Intent != "add_transaction" {
		t.Fatalf("intent = %q, want add_transaction", resp.Intent)
	}
	if resp.Transaction == nil {
		t.Fatal("transaction is nil")
	}
	if resp.Transaction.Amount != 50 {
		t.Fatalf("amount = %v, want 50", resp.Transaction.Amount)
	}
	if resp.Transaction.Channel != "cash" {
		t.Fatalf("channel = %q, want cash", resp.Transaction.Channel)
	}
	if resp.Transaction.Category != "food" {
		t.Fatalf("category = %q, want food", resp.Transaction.Category)
	}

	today := time.Now().Format("2006-01-02")
	if resp.Transaction.TxnDate != today {
		t.Fatalf("txn_date = %q, want %q", resp.Transaction.TxnDate, today)
	}
}

func TestParseTextRegexSummary(t *testing.T) {
	msg := "เดือนนี้ใช้ไปเท่าไหร่"
	resp := parseTextRegex(msg)

	if resp.Intent != "query_summary" {
		t.Fatalf("intent = %q, want query_summary", resp.Intent)
	}
	if resp.Filters == nil {
		t.Fatal("filters is nil")
	}
	if resp.Filters.Period.Type == "" {
		t.Fatal("period type is empty")
	}
	if resp.Filters.Period.From == "" || resp.Filters.Period.To == "" {
		t.Fatalf("period range is empty: %+v", resp.Filters.Period)
	}
}

func TestParseSummaryFiltersLastMonth(t *testing.T) {
	now := time.Now()
	prev := now.AddDate(0, -1, 0)
	first := time.Date(prev.Year(), prev.Month(), 1, 0, 0, 0, 0, prev.Location())
	last := first.AddDate(0, 1, -1)

	filters := parseSummaryFilters("เดือนที่แล้วค่าอาหารเท่าไหร่")

	if filters.Period.Type != "range" {
		t.Fatalf("period type = %q, want range", filters.Period.Type)
	}
	if filters.Period.From != first.Format("2006-01-02") {
		t.Fatalf("from = %q, want %q", filters.Period.From, first.Format("2006-01-02"))
	}
	if filters.Period.To != last.Format("2006-01-02") {
		t.Fatalf("to = %q, want %q", filters.Period.To, last.Format("2006-01-02"))
	}
	if filters.Category != "food" {
		t.Fatalf("category = %q, want food", filters.Category)
	}
}

func TestParseSlipRegex(t *testing.T) {
	ocr := "จ่ายบิลสำเร็จ 30/01/2569 จำนวนเงิน 96.00 TrueMoney ร้านอาหาร"
	resp := parseSlipRegex(ocr)

	if resp.Intent != "bill_payment" {
		t.Fatalf("intent = %q, want bill_payment", resp.Intent)
	}
	if resp.Transaction == nil {
		t.Fatal("transaction is nil")
	}
	if resp.Transaction.Amount != 96.00 {
		t.Fatalf("amount = %v, want 96.00", resp.Transaction.Amount)
	}
	if resp.Transaction.TxnDate != "2026-01-30" {
		t.Fatalf("txn_date = %q, want 2026-01-30", resp.Transaction.TxnDate)
	}
	if resp.Transaction.Channel != "tmw" {
		t.Fatalf("channel = %q, want tmw", resp.Transaction.Channel)
	}
	if resp.Transaction.Category != "food" {
		t.Fatalf("category = %q, want food", resp.Transaction.Category)
	}
}
