.PHONY: help export import check clean

help:
	@echo "byob-go-cli — a forkable template for Go CLI tools"
	@echo ""
	@echo "Targets:"
	@echo "  make export   Re-sync decisions/ + memories/ from the local beads database"
	@echo "                (also writes .beads/issues.jsonl as a build artifact — not tracked in git)"
	@echo "  make import   Import decisions/ + memories/ into the local beads database"
	@echo "  make check    Verify split/join roundtrip is stable"
	@echo "  make clean    Remove the decisions/ and memories/ directories"

export:
	@scripts/split.sh
	@scripts/join.sh > .beads/issues.jsonl
	@echo "Wrote .beads/issues.jsonl ($$(wc -l < .beads/issues.jsonl) records)"

import:
	@tmp=$$(mktemp); \
	scripts/join.sh > "$$tmp"; \
	bd import "$$tmp"; \
	rm -f "$$tmp"

check:
	@tmp=$$(mktemp); \
	scripts/join.sh > "$$tmp"; \
	echo "Joined $$(wc -l < "$$tmp") records from decisions/ + memories/"; \
	rm -f "$$tmp"

clean:
	rm -rf decisions memories
