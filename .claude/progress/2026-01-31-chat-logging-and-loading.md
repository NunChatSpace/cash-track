# Chat logging and loading indicator

Date: 2026-01-31

## Task
- Add backend logging for chat requests and improve chat UX feedback.

## Work done
- Logged chat requests, parsed intent, and created transaction details.
- Added fallback amount extraction from message/OCR when LLM returns zero.
- Added loading message and disabled send button while awaiting response.

## Files touched
- internal/llm/regex.go
- internal/handlers/chat.go
- web/templates/chat.html
- web/static/style.css
