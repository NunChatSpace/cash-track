# Regex fallback when Ollama unavailable

Date: 2026-01-31

## Task
- Add a regex-based parser used when Ollama is unavailable or returns invalid JSON.

## Work done
- Added regex fallback parser for text and OCR inputs (amount/date/category/channel/direction/summary intent).
- Updated LLM client to fall back on regex parsing when Ollama errors or JSON parsing fails.
- Marked the backlog task as completed.

## Files touched
- internal/llm/client.go
- internal/llm/regex.go
- .claude/tasks/backlog.md
