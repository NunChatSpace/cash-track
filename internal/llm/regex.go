package llm

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	amountWithUnitRegex = regexp.MustCompile(`(?i)(\d+(?:[.,]\d+)?)\s*(บาท|฿|thb)`)
	amountRegex         = regexp.MustCompile(`\d+(?:[.,]\d+)?`)
	dateISORegex        = regexp.MustCompile(`\b(\d{4})-(\d{2})-(\d{2})\b`)
	dateSlashRegex      = regexp.MustCompile(`\b(\d{1,2})[/-](\d{1,2})[/-](\d{4})\b`)
)

func parseWithRegex(message string, ocrText *string) *ChatResponse {
	if ocrText != nil && *ocrText != "" {
		return parseSlipRegex(*ocrText)
	}
	return parseTextRegex(message)
}

func parseTextRegex(message string) *ChatResponse {
	lower := strings.ToLower(message)

	if isSummaryQuery(lower) {
		filters := parseSummaryFilters(lower)
		return &ChatResponse{
			Intent:  "query_summary",
			Filters: &filters,
		}
	}

	tx := ParsedTransaction{
		Currency: "THB",
	}

	tx.Amount = parseAmount(lower)
	tx.TxnDate = parseDate(lower)
	tx.Direction = parseDirection(lower, "expense")
	tx.Channel = parseChannel(lower)
	tx.Category = parseCategory(lower)
	tx.Description = strings.TrimSpace(message)

	if tx.Amount == 0 {
		return &ChatResponse{Intent: "unknown"}
	}

	return &ChatResponse{
		Intent: "add_transaction",
		Transaction: &tx,
		Confidence:  0.2,
	}
}

func parseSlipRegex(ocrText string) *ChatResponse {
	lower := strings.ToLower(ocrText)

	tx := ParsedTransaction{
		Currency:  "THB",
		Direction: "expense",
	}

	tx.Amount = parseAmount(lower)
	tx.TxnDate = parseDate(lower)
	tx.Channel = parseChannel(lower)
	tx.Category = parseCategory(lower)
	tx.Description = "slip payment"

	if tx.Amount == 0 {
		return &ChatResponse{Intent: "unknown"}
	}

	return &ChatResponse{
		Intent: "bill_payment",
		Transaction: &tx,
		Confidence:  0.2,
	}
}

func parseAmount(text string) float64 {
	if matches := amountWithUnitRegex.FindStringSubmatch(text); len(matches) > 1 {
		return parseNumber(matches[1])
	}
	if matches := amountRegex.FindAllString(text, -1); len(matches) > 0 {
		return parseNumber(matches[len(matches)-1])
	}
	return 0
}

func parseNumber(value string) float64 {
	value = strings.ReplaceAll(value, ",", "")
	num, _ := strconv.ParseFloat(value, 64)
	return num
}

func parseDate(text string) string {
	if matches := dateISORegex.FindStringSubmatch(text); len(matches) == 4 {
		return matches[1] + "-" + matches[2] + "-" + matches[3]
	}

	if matches := dateSlashRegex.FindStringSubmatch(text); len(matches) == 4 {
		day, _ := strconv.Atoi(matches[1])
		month, _ := strconv.Atoi(matches[2])
		year, _ := strconv.Atoi(matches[3])
		if year >= 2500 {
			year -= 543
		}
		return formatDate(year, month, day)
	}

	now := time.Now()
	if strings.Contains(text, "เมื่อวาน") {
		return now.AddDate(0, 0, -1).Format("2006-01-02")
	}
	if strings.Contains(text, "วันนี้") || strings.Contains(text, "เมื่อกี้") {
		return now.Format("2006-01-02")
	}

	return ""
}

func formatDate(year, month, day int) string {
	if year == 0 || month == 0 || day == 0 {
		return ""
	}
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local).Format("2006-01-02")
}

func parseDirection(text string, fallback string) string {
	if strings.Contains(text, "รายรับ") || strings.Contains(text, "ได้เงิน") || strings.Contains(text, "รับเงิน") || strings.Contains(text, "เงินเข้า") {
		return "income"
	}
	if strings.Contains(text, "โอน") && strings.Contains(text, "ไป") {
		return "transfer"
	}
	if strings.Contains(text, "รายจ่าย") || strings.Contains(text, "ใช้ไป") || strings.Contains(text, "จ่าย") {
		return "expense"
	}
	return fallback
}

func parseChannel(text string) string {
	switch {
	case strings.Contains(text, "cash") || strings.Contains(text, "เงินสด"):
		return "cash"
	case strings.Contains(text, "scb") || strings.Contains(text, "ไทยพาณิชย์"):
		return "scb"
	case strings.Contains(text, "kbank") || strings.Contains(text, "กสิกร"):
		return "kbank"
	case strings.Contains(text, "tmw") || strings.Contains(text, "truemoney") || strings.Contains(text, "ทรูมันนี่"):
		return "tmw"
	case strings.Contains(text, "bbl") || strings.Contains(text, "กรุงเทพ"):
		return "bbl"
	case strings.Contains(text, "ktb") || strings.Contains(text, "กรุงไทย"):
		return "ktb"
	default:
		return ""
	}
}

func parseCategory(text string) string {
	switch {
	case strings.Contains(text, "อาหาร") || strings.Contains(text, "กิน") || strings.Contains(text, "ข้าว") || strings.Contains(text, "ร้าน"):
		return "food"
	case strings.Contains(text, "เดินทาง") || strings.Contains(text, "รถ") || strings.Contains(text, "แท็กซี่") || strings.Contains(text, "bts") || strings.Contains(text, "mrt") || strings.Contains(text, "grab"):
		return "transport"
	case strings.Contains(text, "shopee") || strings.Contains(text, "lazada") || strings.Contains(text, "ซื้อ") || strings.Contains(text, "ช้อปปิ้ง"):
		return "shopping"
	case strings.Contains(text, "บิล") || strings.Contains(text, "ค่าไฟ") || strings.Contains(text, "ค่าน้ำ") || strings.Contains(text, "โทรศัพท์") || strings.Contains(text, "internet"):
		return "bill"
	case strings.Contains(text, "เช่า"):
		return "rent"
	case strings.Contains(text, "หนี้"):
		return "debt"
	default:
		return ""
	}
}

func isSummaryQuery(text string) bool {
	return strings.Contains(text, "เท่าไหร่") ||
		strings.Contains(text, "สรุป") ||
		strings.Contains(text, "เดือนนี้") ||
		strings.Contains(text, "เดือนที่แล้ว") ||
		strings.Contains(text, "ปีนี้") ||
		strings.Contains(text, "ปีที่แล้ว")
}

func parseSummaryFilters(text string) QueryFilters {
	now := time.Now()
	filters := QueryFilters{
		Direction: "expense",
		Period: PeriodFilter{
			Type: "month",
		},
	}

	if strings.Contains(text, "รายรับ") || strings.Contains(text, "ได้เงิน") || strings.Contains(text, "เงินเข้า") {
		filters.Direction = "income"
	} else if strings.Contains(text, "ทั้งสอง") || strings.Contains(text, "รวม") {
		filters.Direction = "both"
	}

	switch {
	case strings.Contains(text, "เดือนที่แล้ว"):
		previous := now.AddDate(0, -1, 0)
		first := time.Date(previous.Year(), previous.Month(), 1, 0, 0, 0, 0, previous.Location())
		last := first.AddDate(0, 1, -1)
		filters.Period = PeriodFilter{
			Type: "range",
			From: first.Format("2006-01-02"),
			To:   last.Format("2006-01-02"),
		}
	case strings.Contains(text, "ปีนี้"):
		first := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		last := time.Date(now.Year(), 12, 31, 0, 0, 0, 0, now.Location())
		filters.Period = PeriodFilter{
			Type: "range",
			From: first.Format("2006-01-02"),
			To:   last.Format("2006-01-02"),
		}
	case strings.Contains(text, "ปีที่แล้ว"):
		prevYear := now.Year() - 1
		first := time.Date(prevYear, 1, 1, 0, 0, 0, 0, now.Location())
		last := time.Date(prevYear, 12, 31, 0, 0, 0, 0, now.Location())
		filters.Period = PeriodFilter{
			Type: "range",
			From: first.Format("2006-01-02"),
			To:   last.Format("2006-01-02"),
		}
	case strings.Contains(text, "วันนี้"):
		today := now.Format("2006-01-02")
		filters.Period = PeriodFilter{
			Type: "day",
			From: today,
			To:   today,
		}
	case strings.Contains(text, "เมื่อวาน"):
		yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
		filters.Period = PeriodFilter{
			Type: "day",
			From: yesterday,
			To:   yesterday,
		}
	default:
		first := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		last := first.AddDate(0, 1, -1)
		filters.Period = PeriodFilter{
			Type: "range",
			From: first.Format("2006-01-02"),
			To:   last.Format("2006-01-02"),
		}
	}

	filters.Category = parseCategory(text)
	filters.Channel = parseChannel(text)

	return filters
}
