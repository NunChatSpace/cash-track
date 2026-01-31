# Delete transaction endpoint and UI

Date: 2026-01-31

## Task
- Add delete transaction API and wire it into the UI.

## Work done
- Added repository delete method with not-found detection.
- Added DELETE /api/transactions/{id} handler to remove DB row and slip file.
- Wired new delete route in the server router.
- Added delete buttons in history and confirm pages with client-side calls.
- Added danger button styling and action layout tweaks.
- Marked the backlog task as completed.

## Files touched
- internal/database/repository.go
- internal/handlers/transactions.go
- cmd/server/main.go
- web/templates/history.html
- web/templates/confirm.html
- web/static/style.css
- .claude/tasks/backlog.md
