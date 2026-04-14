#!/usr/bin/env bash
# join.sh — convert decisions/<id>.md files into a single JSONL stream on stdout.
set -euo pipefail
cd "$(dirname "$0")/.."
python3 scripts/convert.py join
