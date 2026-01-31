# Default cutoff day 1

Date: 2026-01-31

## Task
- Set default cutoff day to 1 instead of 29.

## Work done
- Updated user schema default and backfill to 1.
- Updated new user creation default cutoff to 1.
- Updated chat cutoff fallback default to 1.
- Updated UI default label to Cutoff 1.

## Files touched
- internal/database/migrations.go
- internal/database/repository.go
- internal/handlers/chat.go
- web/templates/layout.html
