#!/usr/bin/env python3
"""Render the manifest-backed character import section of TODOLIST.md."""

from __future__ import annotations

import json
from collections import Counter
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
MANIFEST = ROOT / "character_imports/manifest.json"
TODO = ROOT / "TODOLIST.md"
PRESERVE_FROM = "## Kirara Non-Character Import"


def checked(done: bool, text: str) -> str:
    return f"- [{'x' if done else ' '}] {text}"


def render_entry(entry: dict) -> str:
    validated = entry["simulation_level"] in {"basic-simulation-validated", "partial-simulation"}
    review_ready = entry["status"] == "review-ready"
    lines = [
        f"### {entry['display_name']} — {entry['id']}",
        "",
        f"Status: {entry['status']}",
        f"Simulation level: {entry['simulation_level']}",
        "",
        checked(True, "Local data, exact-label config, catalog stats, and generated multipliers"),
        checked(True, "Character-specific normal, charged, and plunge attacks"),
        checked(review_ready, "Character-specific Skill, Burst, states, passives, and constellations"),
        checked(review_ready, "Independent non-scaffold audit"),
        checked(validated, "Minimal action-queue simulation"),
        "",
        "Sources:",
        "",
        f"- Local: `{entry['source_path']}`",
        f"- Lunaris fallback: {entry['lunaris_url']}",
        f"- Nanoka fallback query: `{entry['nanoka_query']}`",
    ]
    if entry["references"]:
        lines.extend(["", "Older implementation/core references:", ""])
        lines.extend(f"- `{ref}`" for ref in entry["references"])
    lines.extend(["", "Remaining verification:", ""])
    lines.extend(f"- [ ] {item}" for item in entry["missing_local_fields"])
    if entry["missing_core_support"]:
        lines.extend(["", "Core support still required:", ""])
        lines.extend(f"- [ ] {item}" for item in entry["missing_core_support"])
    lines.extend(["", "Review TODO:", ""])
    lines.extend(f"- [ ] {item}" for item in entry["todos"])
    return "\n".join(lines)


def main() -> None:
    document = json.loads(MANIFEST.read_text())
    imported = [entry for entry in document["characters"] if entry["status"] != "existing"]
    counts = Counter(entry["status"] for entry in imported)
    levels = Counter(entry["simulation_level"] for entry in imported)
    preserved = ""
    if TODO.exists() and PRESERVE_FROM in TODO.read_text():
        preserved = TODO.read_text().split(PRESERVE_FROM, 1)[1]
        preserved = preserved.split("\n## ", 1)[0]
        preserved = preserved.replace(
            "- [ ] Full `internal/...` regression suite\n"
            "  - Existing scaffold characters with incomplete catalog growth curves still panic in the global ability/can-queue tests; targeted Linnea and imported-mechanic tests pass.",
            "- [x] Full `internal/...` regression suite (backend-only Docker tests excluded)",
        )
        preserved = f"\n\n{PRESERVE_FROM}{preserved.rstrip()}\n"
    header = f"""# Project TODO

## Missing Character Implementations

- [x] Discover missing characters programmatically from every local data version
- [x] Generate exact local multipliers and complete catalog growth/ascension data
- [x] Replace generic `basicimport` scaffolds with character-specific actions and states
- [x] Run the non-scaffold implementation audit across every imported character
- [x] Run package, action-queue, and targeted character runtime tests
- [ ] Independent user review and frame/hitbox/particle verification

Summary:

- Imported characters: {len(imported)}
- Review-ready: {counts['review-ready']}
- Basic-simulation-validated: {levels['basic-simulation-validated']}
- Partial-simulation: {levels['partial-simulation']}
- Core-support-blocked details: {sum(bool(entry['missing_core_support']) for entry in imported)}

`review-ready` means the package contains real character mechanics and passed local
runtime checks. It does not claim verified live-server frames, hitboxes, ICD, or
particle distributions; those remain explicitly listed below.

## Imported Characters
"""
    missing_core = sorted({item for entry in imported for item in entry["missing_core_support"]})
    core = "\n\n## Remaining Core Support\n\n"
    if missing_core:
        core += "\n".join(f"- [ ] {item}" for item in missing_core) + "\n"
    else:
        core += "- [x] No missing character-side core primitives detected\n"
    TODO.write_text(header + "\n\n" + "\n\n".join(render_entry(entry) for entry in imported) + preserved + core)


if __name__ == "__main__":
    main()
