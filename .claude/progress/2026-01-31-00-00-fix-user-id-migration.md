# Fix user_id migration/index

Date: 2026-01-31

## Task
- Fix startup failure when existing DB lacks user_id column.

## Work done
- Moved user_id index creation to after ALTER TABLE migrations to avoid missing-column errors.

## Files touched
- internal/database/migrations.go
