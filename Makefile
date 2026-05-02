.PHONY: help export import check clean site site-clean

help:
	@echo "byob-go-cli — a forkable template for Go CLI tools"
	@echo ""
	@echo "Targets:"
	@echo "  make export   Re-sync decisions/ + memories/ from the local beads database"
	@echo "                (also writes .beads/issues.jsonl as a build artifact — not tracked in git)"
	@echo "  make import   Import decisions/ + memories/ into the local beads database"
	@echo "  make check    Verify split/join roundtrip is stable"
	@echo "  make site     Build the static site into _site/"
	@echo "  make clean    Remove decisions/, memories/, and _site/"

export:
	@bd export 2>/dev/null | go run ./cmd/byob split
	@go run ./cmd/byob join > .beads/issues.jsonl
	@echo "Wrote .beads/issues.jsonl ($$(wc -l < .beads/issues.jsonl) records)"

import:
	@tmp=$$(mktemp); \
	go run ./cmd/byob join > "$$tmp"; \
	bd import "$$tmp"; \
	rm -f "$$tmp"

check:
	@tmp=$$(mktemp); \
	go run ./cmd/byob join > "$$tmp"; \
	echo "Joined $$(wc -l < "$$tmp") records from decisions/ + memories/"; \
	rm -f "$$tmp"

site:
	@go run ./cmd/byob site -out _site
	@echo "Wrote _site/ ($$(find _site -type f | wc -l) files)"

site-clean:
	rm -rf _site

clean:
	rm -rf decisions memories _site
