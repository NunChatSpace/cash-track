package ocr

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ParsedSlip struct {
	Amount          float64
	TransactionDate time.Time
	FromAccount     string
	ToAccount       string
	Channel         string
}

func ParseSlipText(text string) ParsedSlip {
	result := ParsedSlip{}

	result.Amount = parseAmount(text)
	result.TransactionDate = parseDate(text)
	result.FromAccount = parseFromAccount(text)
	result.ToAccount = parseToAccount(text)
	result.Channel = parseChannel(text)

	return result
}

func parseAmount(text string) float64 {
	patterns := []string{
		`(?i)(?:จำนวนเงิน|amount|THB|฿)\s*[:\s]*([0-9,]+\.?\d*)`,
		`([0-9,]+\.[0-9]{2})\s*(?:บาท|THB|฿)`,
		`(?i)(?:ยอดเงิน|total)\s*[:\s]*([0-9,]+\.?\d*)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			amountStr := strings.ReplaceAll(matches[1], ",", "")
			if amount, err := strconv.ParseFloat(amountStr, 64); err == nil && amount > 0 {
				return amount
			}
		}
	}

	// Fallback: find any decimal number that looks like money
	re := regexp.MustCompile(`([0-9,]+\.[0-9]{2})`)
	matches := re.FindAllStringSubmatch(text, -1)
	for _, match := range matches {
		amountStr := strings.ReplaceAll(match[1], ",", "")
		if amount, err := strconv.ParseFloat(amountStr, 64); err == nil && amount > 0 {
			return amount
		}
	}

	return 0
}

func parseDate(text string) time.Time {
	patterns := []struct {
		regex  string
		layout string
	}{
		{`(\d{2}/\d{2}/\d{4}\s+\d{2}:\d{2})`, "02/01/2006 15:04"},
		{`(\d{2}-\d{2}-\d{4}\s+\d{2}:\d{2})`, "02-01-2006 15:04"},
		{`(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2})`, "2006-01-02 15:04"},
		{`(\d{2}/\d{2}/\d{2}\s+\d{2}:\d{2})`, "02/01/06 15:04"},
		{`(\d{1,2}\s+(?:ม\.ค\.|ก\.พ\.|มี\.ค\.|เม\.ย\.|พ\.ค\.|มิ\.ย\.|ก\.ค\.|ส\.ค\.|ก\.ย\.|ต\.ค\.|พ\.ย\.|ธ\.ค\.)\s+\d{2,4})`, ""},
	}

	for _, p := range patterns {
		re := regexp.MustCompile(p.regex)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 && p.layout != "" {
			if t, err := time.Parse(p.layout, matches[1]); err == nil {
				// Handle 2-digit year
				if t.Year() < 100 {
					t = t.AddDate(2000, 0, 0)
				}
				return t
			}
		}
	}

	return time.Now()
}

func parseFromAccount(text string) string {
	patterns := []string{
		`(?i)(?:จาก|from)\s*[:\s]*([^\n]+)`,
		`(?i)(?:ผู้โอน|sender)\s*[:\s]*([^\n]+)`,
		`(?i)(?:บัญชี|account)\s*[:\s]*([^\n]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	return ""
}

func parseToAccount(text string) string {
	patterns := []string{
		`(?i)(?:ไปยัง|ถึง|to)\s*[:\s]*([^\n]+)`,
		`(?i)(?:ผู้รับ|receiver|recipient)\s*[:\s]*([^\n]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	return ""
}

func parseChannel(text string) string {
	channels := map[string][]string{
		"PromptPay": {"promptpay", "พร้อมเพย์", "prompt pay"},
		"K PLUS":    {"k plus", "kplus", "k-plus", "กสิกร"},
		"SCB Easy":  {"scb easy", "scbeasy", "ไทยพาณิชย์"},
		"KMA":       {"kma", "กรุงศรี"},
		"Bualuang":  {"bualuang", "กรุงเทพ", "bangkok bank"},
		"ttb touch": {"ttb", "ทหารไทยธนชาต"},
		"MAKE":      {"make by kbank"},
	}

	lowerText := strings.ToLower(text)

	for channel, keywords := range channels {
		for _, keyword := range keywords {
			if strings.Contains(lowerText, keyword) {
				return channel
			}
		}
	}

	return ""
}
