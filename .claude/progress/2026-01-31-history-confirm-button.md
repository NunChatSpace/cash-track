# History confirm button + status colors

Date: 2026-01-31

## Task
- Make pending transactions easier to see and confirm without opening edit.

## Work done
- Added pastel background colors for pending (yellow) and confirmed (green) cards.
- Added inline Confirm button for pending transactions in history.
- Confirm button calls PATCH confirm with existing card data and updates UI.

## Files touched
- web/templates/history.html
- web/static/style.css
