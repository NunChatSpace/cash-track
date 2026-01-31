# Query summary via chat + regex tests

Date: 2026-01-31

## Task
- Implement chat query summary handling and add tests for regex fallback.

## Work done
- Fixed query summary aggregation to return correct expense/income totals.
- Added period text to chat summary replies and supported "all" range.
- Enhanced regex summary parsing to detect "ทั้งหมด".
- Added unit tests for regex parsing (amount/date/summary/slip).
- Marked the backlog task as completed.

## Tests
- go test ./... (failed: go binary not available in environment)

## Files touched
- internal/database/repository.go
- internal/handlers/chat.go
- internal/llm/regex.go
- internal/llm/regex_test.go
- .claude/tasks/backlog.md
