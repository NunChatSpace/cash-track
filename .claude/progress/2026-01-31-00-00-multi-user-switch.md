# Multi-user support (name-based switching)

Date: 2026-01-31

## Task
- Add multi-user support without auth, allowing switching by user name.

## Work done
- Added users table and user_id to transactions with default user backfill.
- Added user CRUD/select API and cookie-based current user selection.
- Scoped all transaction and dashboard queries to current user.
- Added navbar user switch UI with create-and-select flow.
- Marked backlog item as completed.

## Files touched
- internal/database/migrations.go
- internal/database/repository.go
- internal/models/transaction.go
- internal/models/user.go
- internal/handlers/handlers.go
- internal/handlers/users.go
- internal/handlers/transactions.go
- internal/handlers/chat.go
- internal/handlers/dashboard.go
- cmd/server/main.go
- web/templates/layout.html
- web/static/style.css
- .claude/tasks/backlog.md
