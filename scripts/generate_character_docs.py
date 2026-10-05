#!/usr/bin/env python3
"""Regenerate character script documentation without downloading datamine.

Requires PyYAML. Uses the same Names/Actions/Params/Fields/Issues JSON schemas
as pipeline/tmpl.go, with internal/characters/**/config.yml as the source.
This does not regenerate talent tables or mark mechanics as verified.
"""

import argparse
import json
import re
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parents[1]
OUTPUT = ROOT / "ui/packages/docs/src/components"
ACTIONS = (
    "attack", "aim", "charge", "low_plunge", "high_plunge", "skill", "burst",
    "dash", "jump", "walk", "swap",
)
DEFAULT_ACTIONS = {"dash", "jump", "walk", "swap"}
ABILITY_ORDER = (
    "default", "attack", "aim", "charge", "plunge", "low_plunge", "high_plunge",
    "skill", "burst", "asc", "a1", "a4", "cons", "dash", "jump", "walk", "swap",
    "c1", "c2", "c3", "c4", "c5", "c6", "set1", "set2", "set4", "passive",
)


def generate():
    outputs = {name: {} for name in ("Names", "Actions", "Params", "Fields", "Issues")}
    names = set()
    for path in sorted((ROOT / "internal/characters").rglob("config.yml")):
        config = yaml.safe_load(path.read_text(encoding="utf-8"))
        if config.get("use") != "pipeline" or config.get("kind") != "character":
            continue
        name = config["name"]
        if name in names:
            raise ValueError(f"Duplicate character name: {name}")
        names.add(name)
        abilities = config.get("abilities", [])
        if len({a["name"] for a in abilities}) != len(abilities):
            raise ValueError(f"Duplicate ability in {path}")
        abilities = sorted(abilities, key=lambda a: ABILITY_ORDER.index(a["name"]))
        by_action = {a["name"]: a for a in abilities}
        aliases = sorted({re.sub(r"[^a-z0-9]", "", s.lower()) for s in config.get("shortcuts", [])})
        if aliases:
            outputs["Names"][name] = aliases
        outputs["Actions"][name] = []
        for action in ACTIONS:
            entry = {"ability": action}
            note = by_action.get(action, {}).get("note")
            if note:
                entry["note"] = note
            entry["invalid"] = action not in by_action and action not in DEFAULT_ACTIONS
            outputs["Actions"][name].append(entry)
        params = [
            {"ability": "-" if a["name"] == "default" else a["name"], "param": p, "desc": desc}
            for a in abilities
            for p, desc in sorted(a.get("params", {}).items())
        ]
        if params:
            outputs["Params"][name] = params
        docs = config.get("docs", {})
        if docs.get("issues"):
            outputs["Issues"][name] = docs["issues"]
        if docs.get("fields"):
            outputs["Fields"][name] = [
                {"field": field, "desc": desc}
                for field, desc in sorted(docs["fields"].items())
            ]
    if not names:
        raise ValueError("No character configs found")
    return outputs, len(names)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true", help="fail if generated files are stale")
    args = parser.parse_args()
    outputs, count = generate()
    stale = []
    for component, data in outputs.items():
        path = OUTPUT / component / "character.dm.json"
        # Sort only the outer map; retain pipeline's field order within records.
        text = json.dumps(dict(sorted(data.items())), ensure_ascii=False, indent=2) + "\n"
        if path.exists() and path.read_text(encoding="utf-8") == text:
            continue
        stale.append(str(path.relative_to(ROOT)))
        if not args.check:
            path.write_text(text, encoding="utf-8")
            print(f"generated {path.relative_to(ROOT)}")
    if args.check and stale:
        raise SystemExit("Stale character documentation:\n" + "\n".join(stale))
    print(f"{count} character configs; script documentation {'checked' if args.check else 'updated'}")


if __name__ == "__main__":
    main()
