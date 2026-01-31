# Per-user cutoff day + improved user switch UI

Date: 2026-01-31

## Task
- Make cutoff day configurable per user and improve the user switch UI.

## Work done
- Added cutoff_day to users with default 29.
- Added API to update a user’s cutoff day.
- Applied cutoff day to chat summary monthly ranges and dashboard range picker.
- Added clamped date handling for months shorter than cutoff day.
- Redesigned user switch UI with badge, labels, and cutoff button.

## Files touched
- internal/database/migrations.go
- internal/models/user.go
- internal/database/repository.go
- internal/handlers/users.go
- internal/handlers/chat.go
- cmd/server/main.go
- web/templates/dashboard.html
- web/templates/layout.html
- web/static/style.css
