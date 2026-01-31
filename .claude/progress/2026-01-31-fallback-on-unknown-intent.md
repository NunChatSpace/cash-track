# Fallback on unknown intent

Date: 2026-01-31

## Task
- Ensure chat summary works when LLM responds with unknown intent.

## Work done
- Added regex fallback when LLM returns intent "unknown" or empty.

## Files touched
- internal/llm/client.go
