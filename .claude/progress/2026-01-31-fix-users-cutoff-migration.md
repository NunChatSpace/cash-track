# Fix users cutoff migration

Date: 2026-01-31

## Task
- Ensure existing users table includes cutoff_day so user list loads.

## Work done
- Added ALTER TABLE migration for users.cutoff_day.

## Files touched
- internal/database/migrations.go
