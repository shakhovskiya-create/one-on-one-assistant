#!/usr/bin/env bash
set -euo pipefail

FROM="${1:-PM}"
TO="${2:-ANALYST}"
SLUG="${3:-task}"

DATE="$(date +%F)"
OUT="ai_org/handoffs/active/${DATE}__${FROM}__${TO}__${SLUG}.md"

if [ ! -f "ai_org/handoffs/templates/HANDOFF_TEMPLATE.md" ]; then
  echo "❌ Missing template: ai_org/handoffs/templates/HANDOFF_TEMPLATE.md"
  exit 1
fi

mkdir -p ai_org/handoffs/active
cp ai_org/handoffs/templates/HANDOFF_TEMPLATE.md "$OUT"

echo "✅ Created: $OUT"
