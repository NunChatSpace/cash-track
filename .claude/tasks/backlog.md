# Cash Track - Backlog

## Current Status (2026-01-31)

### Spec v3 Implementation - COMPLETED

All 6 core tasks from the spec have been implemented:

| Task | Description | Status |
|------|-------------|--------|
| #1 | Database schema update (direction, category, channel, etc.) | Done |
| #2 | POST /api/chat endpoint | Done |
| #3 | LLM prompts for all intents | Done |
| #4 | Dashboard API endpoints | Done |
| #5 | Chat UI page | Done |
| #6 | Dashboard UI with charts | Done |

### Additional Improvements Made

- [x] OCR client Content-Type header fix
- [x] Go template rendering fix (page isolation)
- [x] SQLite datetime handling fix
- [x] Dashboard date fallback to created_at when txn_date empty
- [x] Chat loads recent transactions on page load
- [x] Dashboard date inputs show current period by default
- [x] Timezone fix for date formatting
- [x] History shows "unknown"/"uncategorized" badges when empty
- [x] Pending status when category/channel missing (user can edit)
- [x] Edit button for all transactions (not just pending)
- [x] Custom spending period cutoff (29th-28th pay cycle)

---

## What's Working Now

### Chat (`/chat`)
- Text input for expenses: "กินข้าว 50 บาท"
- Image upload for slips (OCR + LLM parsing)
- Shows recent transaction history on load
- Links to edit pending transactions
- Thai language responses

### Dashboard (`/dashboard`)
- Expense/Income totals
- Category breakdown (pie chart)
- Channel breakdown (bar chart)
- Date range: เดือนนี้ (29th-28th), เดือนที่แล้ว, ปีนี้, custom
- Percentage calculations

### History (`/history`)
- All transactions list
- Edit button for all (pending + confirmed)
- Category/Channel badges (shows unknown if empty)
- Direction-based coloring (red expense, green income)

### Upload (`/`)
- Drag & drop slip upload
- OCR text extraction
- LLM parsing to structured data

---

## Remaining Features

### High Priority
- [ ] Delete transaction endpoint and UI
- [ ] Fallback to regex parser when Ollama unavailable
- [ ] Query summary via chat ("เดือนนี้ใช้ไปเท่าไหร่")

### Medium Priority
- [ ] Export transactions to CSV
- [ ] Export transactions to Excel
- [ ] Search/filter transactions
- [ ] Income tracking (currently expense-focused)

### Low Priority
- [ ] User authentication
- [ ] Multi-user support
- [ ] Recurring transaction detection
- [ ] Duplicate slip detection
- [ ] Backup/restore database
- [ ] Configurable cutoff day (currently hardcoded to 29)

### Technical Debt
- [ ] Unit tests
- [ ] Integration tests
- [ ] API documentation (OpenAPI/Swagger)
- [ ] Graceful shutdown handling
- [ ] Health check endpoint for Go server
- [ ] Logging to file
- [ ] Environment config (.env support)

---

## Architecture

```
cash-track/
├── cmd/server/main.go          # Entry point, routes
├── internal/
│   ├── config/config.go        # Configuration
│   ├── database/
│   │   ├── db.go               # SQLite connection
│   │   ├── migrations.go       # Schema
│   │   └── repository.go       # Data access
│   ├── handlers/
│   │   ├── handlers.go         # Page handlers
│   │   ├── chat.go             # Chat API
│   │   ├── dashboard.go        # Dashboard API
│   │   └── transactions.go     # Transaction API
│   ├── llm/
│   │   ├── client.go           # Ollama client
│   │   └── prompts.go          # LLM prompt templates
│   ├── models/
│   │   ├── transaction.go      # Transaction model
│   │   └── dashboard.go        # Dashboard models
│   ├── ocr/client.go           # EasyOCR client
│   └── storage/storage.go      # File storage
├── web/
│   ├── templates/              # HTML templates
│   │   ├── layout.html
│   │   ├── index.html
│   │   ├── chat.html
│   │   ├── dashboard.html
│   │   ├── history.html
│   │   └── confirm.html
│   └── static/style.css        # CSS styles
├── ocr-service/                # Python EasyOCR service
├── docker-compose.yml          # OCR + Ollama services
└── Makefile                    # Build commands
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /api/chat | Chat message (text + optional image) |
| POST | /api/transactions/slip | Upload slip image |
| GET | /api/transactions/recent | Recent transactions for chat |
| GET | /api/transactions/{id} | Get single transaction |
| PATCH | /api/transactions/{id}/confirm | Confirm/update transaction |
| GET | /api/dashboard/summary | Dashboard summary with charts |
| GET | /api/dashboard/by-category | Category breakdown |
| GET | /api/dashboard/by-channel | Channel breakdown |
