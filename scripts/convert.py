#!/usr/bin/env python3
"""Convert between bd-export JSONL and per-bead / per-memory Markdown files.

Usage:
  scripts/convert.py split < exported.jsonl   # write decisions/<slug>/<id>.md + memories/*.md
  scripts/convert.py join                     # emit unified JSONL to stdout

Decision / epic beads live at `decisions/<slug>/<id>.md` with frontmatter +
body sections (`## Description`, `## Design`). The `<slug>` is derived from
the epic's (or standalone decision's) title, so the directory tree mirrors
the parent → child grouping that already exists in the data.

Memories — the tip layer, auto-injected by `bd prime` — live at
`memories/<key>.md` with a tiny frontmatter (key, optional category) and
a short free-form body.

On split, memory lines (`_type: memory`) go to `memories/`. Issue
records with `issue_type` of `decision` or `epic` go to `decisions/`.
Other issue types (tasks, bugs, etc.) are skipped — they belong to the
consuming project's own workflow, not the template. On join, both trees
are walked and their records are merged into one JSONL stream suitable
for `bd import` or `bd init --from-jsonl`.

Personal metadata (owner, created_at, created_by, updated_at) is never
written into the md files — Dolt inside `.beads/` keeps the real history.
"""
from __future__ import annotations

import argparse
import json
import pathlib
import re
import shutil
import sys

import yaml

ROOT = pathlib.Path(__file__).resolve().parent.parent
DECISIONS_DIR = ROOT / "decisions"
MEMORIES_DIR = ROOT / "memories"

H2_LINE_RE = re.compile(r"^## (\S[^\n]*?)\n?$")
SLUG_NONALNUM_RE = re.compile(r"[^a-z0-9]+")


def _slug(title: str) -> str:
    """Lowercase, hyphen-collapsed slug used as the epic subdirectory name."""
    return SLUG_NONALNUM_RE.sub("-", title.lower()).strip("-")


def _split_h2_sections(body: str) -> dict[str, str]:
    """Slice `body` into {section_name_lower: content} on `## ` lines.

    Skips `## ` lines that fall inside triple-backtick fenced code blocks,
    so example markdown embedded in a Design block (e.g. a sample README
    with `## Install`) doesn't get mistaken for a real section header.
    """
    sections: dict[str, str] = {}
    current_name: str | None = None
    current_start: int = 0
    in_fence = False
    pos = 0
    for line in body.splitlines(keepends=True):
        if line.lstrip(" ").startswith("```"):
            in_fence = not in_fence
        elif not in_fence:
            m = H2_LINE_RE.match(line)
            if m:
                if current_name is not None:
                    sections[current_name] = body[current_start:pos].strip("\n")
                current_name = m.group(1).strip().lower()
                current_start = pos + len(line)
        pos += len(line)
    if current_name is not None:
        sections[current_name] = body[current_start:].strip("\n")
    return sections


# ---------- issue (decision/epic) beads ----------

def bead_to_md(bead: dict) -> str:
    """Serialize one issue bead (from `bd export`) as a markdown string."""
    fm: dict = {
        "id": bead["id"],
        "title": bead["title"],
        "type": bead.get("issue_type", "task"),
        "priority": bead.get("priority", 2),
        "status": bead.get("status", "open"),
    }
    for dep in bead.get("dependencies", []) or []:
        if dep.get("type") == "parent-child" and dep.get("depends_on_id"):
            fm["parent"] = dep["depends_on_id"]
            break
    fm["labels"] = sorted(bead.get("labels", []) or [])
    # Personal metadata (owner/created_at/created_by/updated_at) is intentionally
    # omitted — Dolt inside .beads/ preserves real history.

    yaml_block = yaml.safe_dump(
        fm,
        sort_keys=False,
        default_flow_style=False,
        allow_unicode=True,
        width=1000,
    ).rstrip()

    parts = ["---", yaml_block, "---", ""]

    desc = (bead.get("description") or "").rstrip()
    if desc:
        parts += ["## Description", "", desc, ""]

    design = (bead.get("design") or "").rstrip()
    if design:
        parts += ["## Design", "", design, ""]

    return "\n".join(parts) + "\n"


def md_to_bead(text: str) -> dict:
    """Parse a decisions/*.md file back into an issue record for `bd import`."""
    m = re.match(r"^---\n(.*?)\n---\n?", text, re.S)
    if not m:
        raise ValueError("no frontmatter found")
    fm = yaml.safe_load(m.group(1)) or {}
    body = text[m.end():]

    sections = _split_h2_sections(body)

    bead: dict = {
        "id": fm["id"],
        "title": fm["title"],
        "issue_type": fm.get("type", "task"),
        "priority": fm.get("priority", 2),
        "status": fm.get("status", "open"),
        "labels": fm.get("labels", []) or [],
        "description": sections.get("description", ""),
        "design": sections.get("design", ""),
        "dependencies": [],
        "dependency_count": 0,
        "dependent_count": 0,
        "comment_count": 0,
    }
    if fm.get("parent"):
        bead["dependencies"].append({
            "issue_id": fm["id"],
            "depends_on_id": fm["parent"],
            "type": "parent-child",
            "metadata": "{}",
        })
    return bead


# ---------- memories ----------

def memory_to_md(record: dict) -> str:
    """Serialize one memory record as a markdown file."""
    fm = {"key": record["key"]}
    category = record.get("category")
    if category:
        fm["category"] = category
    yaml_block = yaml.safe_dump(
        fm, sort_keys=False, default_flow_style=False, allow_unicode=True
    ).rstrip()
    body = (record.get("value") or "").rstrip()
    return "---\n" + yaml_block + "\n---\n\n" + body + "\n"


def md_to_memory(text: str) -> dict:
    """Parse a memories/*.md file back into a memory record for `bd import`."""
    m = re.match(r"^---\n(.*?)\n---\n?", text, re.S)
    if not m:
        raise ValueError("no frontmatter found")
    fm = yaml.safe_load(m.group(1)) or {}
    body = text[m.end():].strip("\n")
    rec = {
        "_type": "memory",
        "key": fm["key"],
        "value": body,
    }
    # Preserve optional category for human-readable grouping; beads itself
    # doesn't enforce it, but we keep it in frontmatter for our tooling.
    return rec


# ---------- subcommands ----------

def cmd_split(_args) -> None:
    # Wipe the decisions tree wholesale so renamed/removed slug folders
    # don't linger between runs. memories/ stays flat — only purge files.
    shutil.rmtree(DECISIONS_DIR, ignore_errors=True)
    DECISIONS_DIR.mkdir(parents=True, exist_ok=True)
    MEMORIES_DIR.mkdir(exist_ok=True)
    for f in MEMORIES_DIR.glob("*.md"):
        f.unlink()

    # Two-pass: buffer records so we can resolve a child decision's
    # parent title (and hence subdir slug) regardless of stream order.
    issue_recs: list[dict] = []
    memory_recs: list[dict] = []
    skipped_types: dict[str, int] = {}
    for line in sys.stdin:
        line = line.strip()
        if not line:
            continue
        rec = json.loads(line)
        if rec.get("_type") == "memory":
            if rec.get("key"):
                memory_recs.append(rec)
            continue
        if not rec.get("id"):
            continue
        issue_type = rec.get("issue_type")
        if issue_type not in ("decision", "epic"):
            skipped_types[issue_type or "<missing>"] = (
                skipped_types.get(issue_type or "<missing>", 0) + 1
            )
            continue
        issue_recs.append(rec)

    titles_by_id = {r["id"]: r.get("title", "") for r in issue_recs}

    def parent_id(rec: dict) -> str | None:
        for dep in rec.get("dependencies", []) or []:
            if dep.get("type") == "parent-child" and dep.get("depends_on_id"):
                return dep["depends_on_id"]
        return None

    n_decisions = 0
    for rec in issue_recs:
        pid = parent_id(rec)
        slug_source = titles_by_id.get(pid) if pid else rec.get("title", "")
        slug = _slug(slug_source) or rec["id"]
        subdir = DECISIONS_DIR / slug
        subdir.mkdir(parents=True, exist_ok=True)
        (subdir / f"{rec['id']}.md").write_text(
            bead_to_md(rec), encoding="utf-8"
        )
        n_decisions += 1

    n_memories = 0
    for rec in memory_recs:
        (MEMORIES_DIR / f"{rec['key']}.md").write_text(
            memory_to_md(rec), encoding="utf-8"
        )
        n_memories += 1

    print(
        f"Wrote {n_decisions} decision files and {n_memories} memory files",
        file=sys.stderr,
    )
    if skipped_types:
        breakdown = ", ".join(
            f"{k}={v}" for k, v in sorted(skipped_types.items())
        )
        print(
            f"Skipped {sum(skipped_types.values())} non-decision issues ({breakdown})",
            file=sys.stderr,
        )


def cmd_join(_args) -> None:
    for f in sorted(DECISIONS_DIR.rglob("*.md"), key=lambda p: p.as_posix()):
        bead = md_to_bead(f.read_text(encoding="utf-8"))
        print(json.dumps(bead, ensure_ascii=False))
    if MEMORIES_DIR.exists():
        for f in sorted(MEMORIES_DIR.glob("*.md")):
            rec = md_to_memory(f.read_text(encoding="utf-8"))
            print(json.dumps(rec, ensure_ascii=False))


def main() -> None:
    p = argparse.ArgumentParser(description=__doc__)
    sub = p.add_subparsers(dest="cmd", required=True)
    sub.add_parser("split", help="stdin JSONL -> decisions/*.md + memories/*.md")
    sub.add_parser("join", help="decisions/*.md + memories/*.md -> stdout JSONL")
    args = p.parse_args()
    {"split": cmd_split, "join": cmd_join}[args.cmd](args)


if __name__ == "__main__":
    main()
