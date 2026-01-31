# Chat OCR path fix

Date: 2026-01-31

## Task
- Fix OCR failing in chat when image_path is a filename.

## Work done
- Resolved image_path to storage path when not absolute before OCR.

## Files touched
- internal/handlers/chat.go
