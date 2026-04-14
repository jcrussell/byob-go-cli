#!/usr/bin/env bash
# split.sh — export beads (issues + memories) to decisions/*.md + memories/*.md.
set -euo pipefail
cd "$(dirname "$0")/.."
bd export 2>/dev/null | python3 scripts/convert.py split
