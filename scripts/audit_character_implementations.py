#!/usr/bin/env python3
"""Audit imported character packages for scaffold-only implementations."""

from __future__ import annotations

import argparse
import json
import re
import sys
from collections import defaultdict
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
MANIFEST = ROOT / "character_imports/manifest.json"
TODO = ROOT / "TODOLIST.md"
HANDWRITTEN = ("attack.go", "charge.go", "plunge.go", "skill.go", "burst.go", "asc.go", "cons.go")


def imported_entries(character: str | None = None) -> list[dict]:
    doc = json.loads(MANIFEST.read_text())
    entries = [entry for entry in doc["characters"] if entry["status"] != "existing"]
    if character is not None:
        entries = [entry for entry in entries if entry["name"] == character]
        if not entries:
            raise SystemExit(f"unknown imported character: {character}")
    return entries


def has_logic(text: str) -> bool:
    body = re.sub(r"(?m)^\s*(package\s+\w+|import\s+.*|//.*)$", "", text)
    return bool(re.search(r"\bfunc\b|\btype\b|\bvar\b|\bconst\b", body))


def audit(entry: dict) -> list[str]:
    slug = entry["name"]
    directory = ROOT / "internal/characters" / slug
    issues: list[str] = []
    texts = {}
    for name in HANDWRITTEN:
        path = directory / name
        if not path.is_file():
            issues.append(f"missing {name}")
            continue
        texts[name] = path.read_text()
        if not has_logic(texts[name]):
            issues.append(f"{name} is TODO/comment-only")
    base = (directory / f"{slug}.go").read_text()
    if "basicimport.Character" in base:
        issues.append("base character delegates every action to generic basicimport")
    for name in ("skill.go", "burst.go"):
        text = texts.get(name, "")
        if "QueueAttack" not in text and "Player.Heal" not in text:
            issues.append(f"{name} has no character-specific attack or healing")
        if not re.search(r"AddStatus|AddStatMod|Subscribe|QueueCharTask|SetNumCharges|type\s+\w+", text):
            issues.append(f"{name} has no character-specific state/resource logic")
    for name in ("asc.go", "cons.go"):
        if not re.search(r"\bfunc\b", texts.get(name, "")):
            issues.append(f"{name} has no implementation functions")
    generated = (directory / f"zz_{slug}.dm.go").read_text()
    for variable in re.findall(r"^\s*(\w+):\s+basicimport\.Scaling", generated, re.M):
        if variable not in "".join(texts.values()):
            issues.append(f"generated multiplier {variable} is not referenced by handwritten code")
    return issues


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--character", help="audit one imported character by internal name")
    args = parser.parse_args()
    entries = imported_entries(args.character)
    failures = defaultdict(list)
    for entry in entries:
        failures[entry["name"]].extend(audit(entry))
    todo = TODO.read_text() if TODO.exists() else ""
    for entry in entries:
        slug = entry["name"]
        display = entry["display_name"]
        block_match = re.search(
            rf"(?ms)^###\s+{re.escape(display)}\s+—.*?(?=^###\s+|\Z)", todo
        )
        block = block_match.group(0) if block_match else ""
        if not block:
            failures[slug].append("TODOLIST character block missing")
        elif failures[slug] and "Status: compile-ready" in block:
            failures[slug].append("TODOLIST incorrectly marks scaffold as compile-ready")
        if failures[slug] and entry["status"] == "compile-ready":
            failures[slug].append("manifest incorrectly marks scaffold as compile-ready")
    if any(failures.values()):
        for slug, issues in failures.items():
            if not issues:
                continue
            print(f"{slug}:")
            for issue in sorted(set(issues)):
                print(f"  - {issue}")
        sys.exit(1)
    print(f"audited={len(entries)}")


if __name__ == "__main__":
    main()
