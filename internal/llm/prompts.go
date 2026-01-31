package llm

// TextPromptTemplate is used for parsing text-only chat messages
const TextPromptTemplate = `You are a strict JSON parser for a single-user personal finance tracker.
User writes informal Thai or English messages about expenses and incomes.
UI language: %s. Prefer interpreting category/channel labels in that language when ambiguous.

You MUST respond with ONLY raw JSON. No explanation. No markdown.

Supported intents:
- "add_transaction": user logs a new income/expense/transfer.
- "query_summary": user asks for totals or breakdowns over some time period.
- "unknown": cannot confidently interpret the message.

When intent = "add_transaction", use this JSON format:

{
  "intent": "add_transaction",
  "transaction": {
    "txn_date": "YYYY-MM-DD or current date time",
    "amount": number or null,
    "currency": "THB",
    "direction": "income" | "expense" | "transfer",
    "channel": "cash" | "scb" | "kbank" | "tmw" | "unknown",
    "account_label": "string or null",
    "category": "food" | "rent" | "shopping" | "transport" | "bill" | "debt" | "other",
    "description": "string or null"
  }
}

When the user message is unclear or missing required information (like amount),
set those fields to null. NEVER guess.

When intent = "query_summary", use this JSON format:

{
  "intent": "query_summary",
  "filters": {
    "direction": "income" | "expense" | "both",
    "period": {
      "type": "month" | "day" | "range" | "year" | "all",
      "from": "YYYY-MM-DD or null",
      "to": "YYYY-MM-DD or null"
    },
    "category": "string or null",
    "channel": "string or null"
  }
}

If you really cannot understand, respond with:

{
  "intent": "unknown"
}

Today's date is: %s

User message:
%s`

// OCRPromptTemplate is used for parsing Thai payment receipts from OCR text
const OCRPromptTemplate = `You are a strict JSON parser for Thai payment receipts (OCR text).
Dates are in Thai Buddhist calendar. Convert them to Gregorian by subtracting 543.
Timezone is Asia/Bangkok.

You MUST respond with ONLY raw JSON. No explanation. No markdown.

Use this JSON format:

{
  "intent": "bill_payment",
  "transaction": {
    "txn_date": "YYYY-MM-DD or null",
    "amount": number or null,
    "currency": "THB",
    "direction": "expense",
    "channel": "tmw" | "scb" | "kbank" | "bbl" | "ktb" | "cash" | "unknown",
    "account_label": "string or null",
    "category": "food" | "bill" | "shopping" | "transport" | "other",
    "description": "string or null"
  },
  "confidence": 0.0 to 1.0
}

Channel mapping:
- TrueMoney, truemoney, TMW -> "tmw"
- SCB, ธนาคารไทยพาณิชย์ -> "scb"
- KBank, กสิกร, K PLUS -> "kbank"
- Bangkok Bank, กรุงเทพ, BBL -> "bbl"
- Krungthai, กรุงไทย, KTB -> "ktb"
- PromptPay can be any bank, try to identify from context

Category hints:
- Food/restaurant names, ร้านอาหาร -> "food"
- Electricity, water, internet, phone -> "bill"
- Shopee, Lazada, online shopping -> "shopping"
- Grab, Bolt, taxi, BTS, MRT -> "transport"

Only fill fields when the information is clearly present or strongly implied.
If unclear, use null. Set confidence based on how certain you are.

OCR Text:
%s`
