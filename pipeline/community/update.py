#!/usr/bin/env python3
"""Fetch and normalize preview combat data from Nanoka and Lunaris.

The resulting files are implementation inputs, not active simulator entities. They
intentionally omit lore, voice lines, materials, and images while retaining the
combat-facing fields needed when an implementation is promoted into gcsim.
"""

from __future__ import annotations

import argparse
import concurrent.futures
import hashlib
import json
import re
import sys
import urllib.error
import urllib.request
from pathlib import Path
from typing import Any


HERE = Path(__file__).resolve().parent
REPO_ROOT = HERE.parents[1]
TARGETS_PATH = HERE / "targets.json"
USER_AGENT = "gcsim-community-data/1"

CHAR_NANOKA_FIELDS = (
    "name",
    "weapon",
    "rarity",
    "element",
    "icon",
    "stamina_recovery",
    "base_hp",
    "base_atk",
    "base_def",
    "crit_rate",
    "crit_dmg",
    "elemental_mastery",
    "stats_modifier",
    "skills",
    "passives",
    "constellations",
    "attack",
    "energy",
)
CHAR_LUNARIS_FIELDS = (
    "info",
    "skills",
    "passives",
    "constellations",
    "hyperlinks",
    "energy",
    "attacks",
    "endure",
)
WEAPON_NANOKA_FIELDS = (
    "name",
    "weapon_type",
    "weapon_prop",
    "rarity",
    "icon",
    "stats_modifier",
    "refinement",
)
WEAPON_LUNARIS_FIELDS = (
    "name",
    "passive",
    "qualityType",
    "stats",
    "weaponIcon",
    "weaponType",
)
ARTIFACT_NANOKA_FIELDS = ("id", "icon", "need", "rank", "affix")
ARTIFACT_LUNARIS_FIELDS = ("info",)

UNRESOLVED = {
    "character": [
        "action frames, hit marks, cancel windows, and hitlag",
        "hitbox geometry, targeting, travel time, and multi-target behavior",
        "runtime state-machine ordering, snapshot timing, and server-only logic",
        "particle observations pending verification; Lunaris is the temporary default",
    ],
    "weapon": [
        "trigger ordering and target ownership where the passive text is ambiguous",
        "whether duration and cooldown timers are extended by hitlag",
    ],
    "artifact": [
        "event ordering and off-field behavior not explicitly encoded by the source",
        "the reaction event/tag implementation required by the 4-piece effect",
    ],
    "reaction": [
        "exact base coefficient and level-scaling formula",
        "eligibility conditions represented only as 'certain conditions'",
        "field/vortex duration, radius, trigger count, stored-energy thresholds, and caps",
        "aura consumption, ICD, ownership, snapshot, and multi-target rules",
    ],
}


class SourceError(RuntimeError):
    pass


def load_json(path: Path) -> Any:
    return json.loads(path.read_text(encoding="utf-8"))


def canonical_json(value: Any) -> bytes:
    return (
        json.dumps(value, ensure_ascii=False, indent=2, sort_keys=True) + "\n"
    ).encode("utf-8")


def fetch_json(url: str) -> tuple[Any, str]:
    req = urllib.request.Request(url, headers={"User-Agent": USER_AGENT})
    try:
        with urllib.request.urlopen(req, timeout=60) as response:
            raw = response.read()
    except (urllib.error.HTTPError, urllib.error.URLError) as exc:
        raise SourceError(f"failed to fetch {url}: {exc}") from exc
    try:
        value = json.loads(raw)
    except json.JSONDecodeError as exc:
        raise SourceError(f"{url} did not return JSON") from exc
    return value, hashlib.sha256(raw).hexdigest()


def select(value: dict[str, Any], fields: tuple[str, ...]) -> dict[str, Any]:
    return {key: value[key] for key in fields if key in value}


def slugify(value: str) -> str:
    return "".join(char.lower() for char in value if char.isalnum())


def read_registry_names(path: Path, marker: str) -> set[str]:
    text = path.read_text(encoding="utf-8")
    match = re.search(
        rf"var _{re.escape(marker)}Names = \[\.\.\.\]string\{{(?P<body>.*?)\n\}}",
        text,
        flags=re.DOTALL,
    )
    if match is None:
        raise RuntimeError(f"could not parse generated registry {path}")
    return set(re.findall(r'"([^"]*)"', match.group("body")))


def source_record(url: str, payload: Any, sha256: str) -> dict[str, Any]:
    return {"url": url, "sha256": sha256, "payload": payload}


def float_lists_agree(left: Any, right: Any, tolerance: float = 1e-6) -> bool:
    if not isinstance(left, list) or not isinstance(right, list):
        return False
    if len(left) != len(right):
        return False
    return all(
        isinstance(a, (int, float))
        and isinstance(b, (int, float))
        and abs(float(a) - float(b)) <= tolerance
        for a, b in zip(left, right, strict=True)
    )


def numeric_text(value: Any) -> float | None:
    if isinstance(value, (int, float)):
        return float(value)
    if not isinstance(value, str):
        return None
    match = re.search(r"-?\d+(?:\.\d+)?", value)
    return float(match.group()) if match else None


def energy_comparison(
    nanoka_reports: list[dict[str, Any]],
    lunaris_reports: list[dict[str, Any]],
) -> dict[str, Any]:
    remaining = list(lunaris_reports)
    matched: list[dict[str, Any]] = []
    unmatched_nanoka: list[dict[str, Any]] = []
    for report in nanoka_reports:
        name = str(report.get("name", "")).removeprefix("UNIQUE_")
        index = next(
            (
                i
                for i, candidate in enumerate(remaining)
                if str(candidate.get("source", "")).removeprefix("UNIQUE_")
                == name
            ),
            None,
        )
        if index is None:
            unmatched_nanoka.append(report)
            continue
        other = remaining.pop(index)
        n_cd = numeric_text(report.get("cd"))
        l_cd = numeric_text(other.get("cd"))
        n_chance = numeric_text(report.get("chance"))
        l_chance = numeric_text(other.get("chance"))
        matched.append(
            {
                "source": name,
                "nanoka": report,
                "lunaris": other,
                "cd_agrees": n_cd is not None
                and l_cd is not None
                and abs(n_cd - l_cd) <= 1e-6,
                "chance_agrees_after_display_rounding": n_chance is not None
                and l_chance is not None
                and abs(n_chance - l_chance) <= 1.0,
            }
        )
    return {
        "note": (
            "Nanoka per_drop/applications and Lunaris particles are not treated "
            "as equivalent fields. Particle counts still require verification."
        ),
        "matched": matched,
        "unmatched_nanoka": unmatched_nanoka,
        "unmatched_lunaris": remaining,
    }


def fetch_detail(
    kind: str,
    target: dict[str, Any],
    targets: dict[str, Any],
) -> tuple[dict[str, Any], dict[str, Any]]:
    nanoka_cfg = targets["sources"]["nanoka"]
    lunaris_cfg = targets["sources"]["lunaris"]
    nanoka_kind = {"character": "character", "weapon": "weapon", "artifact": "artifact"}[
        kind
    ]
    lunaris_kind = {"character": "char", "weapon": "weapon", "artifact": "artifact"}[
        kind
    ]
    nanoka_fields = {
        "character": CHAR_NANOKA_FIELDS,
        "weapon": WEAPON_NANOKA_FIELDS,
        "artifact": ARTIFACT_NANOKA_FIELDS,
    }[kind]
    lunaris_fields = {
        "character": CHAR_LUNARIS_FIELDS,
        "weapon": WEAPON_LUNARIS_FIELDS,
        "artifact": ARTIFACT_LUNARIS_FIELDS,
    }[kind]

    nanoka: dict[str, Any] = {"version": nanoka_cfg["preview_version"]}
    lunaris: dict[str, Any] = {"version": lunaris_cfg["preview_version"]}
    lunaris_id = target.get("lunaris_id", target["id"])
    for language, nanoka_language, lunaris_language in (
        ("en", "en", "en"),
        ("zh", "zh", "chs"),
    ):
        n_url = (
            f"{nanoka_cfg['base_url']}/{nanoka_cfg['preview_version']}/"
            f"{nanoka_language}/{nanoka_kind}/{target['id']}.json"
        )
        l_url = (
            f"{lunaris_cfg['base_url']}/{lunaris_cfg['preview_version']}/"
            f"{lunaris_language}/{lunaris_kind}/{lunaris_id}.json"
        )
        n_value, n_hash = fetch_json(n_url)
        l_value, l_hash = fetch_json(l_url)
        nanoka[language] = source_record(
            n_url, select(n_value, nanoka_fields), n_hash
        )
        lunaris[language] = source_record(
            l_url, select(l_value, lunaris_fields), l_hash
        )
    return nanoka, lunaris


def identity_comparison(
    kind: str,
    target: dict[str, Any],
    nanoka: dict[str, Any],
    lunaris: dict[str, Any],
) -> dict[str, Any]:
    n = nanoka["en"]["payload"]
    l = lunaris["en"]["payload"]
    if kind == "character":
        li = l["info"]
        values = {
            "name": [n.get("name"), li.get("name")],
            "element": [n.get("element"), li.get("element")],
            "weapon_type": [n.get("weapon"), li.get("weapon")],
            "rarity": [n.get("rarity"), li.get("rarity")],
        }
        n_attack_names = {item.get("name") for item in n.get("attack", [])}
        l_attack_names = {item.get("name") for item in l.get("attacks", [])}
        combat = {
            "attack_records": {
                "nanoka_count": len(n.get("attack", [])),
                "lunaris_count": len(l.get("attacks", [])),
                "only_nanoka": sorted(n_attack_names - l_attack_names),
                "only_lunaris": sorted(l_attack_names - n_attack_names),
            },
            "energy_reports": energy_comparison(
                n.get("energy", []), l.get("energy", [])
            ),
        }
    elif kind == "weapon":
        values = {
            "name": [n.get("name"), l.get("name")],
            "weapon_type": [n.get("weapon_type"), l.get("weaponType")],
            "rarity": [
                f"QUALITY_{'ORANGE' if n.get('rarity') == 5 else 'PURPLE'}",
                l.get("qualityType"),
            ],
        }
        n_refinements = {
            refine: value.get("param_list", [])
            for refine, value in n.get("refinement", {}).items()
        }
        l_refinements = {
            refine: value.get("params", [])
            for refine, value in l.get("passive", {}).get("refinements", {}).items()
        }
        combat = {
            "refinement_params": {
                "nanoka": n_refinements,
                "lunaris": l_refinements,
                "agrees": n_refinements.keys() == l_refinements.keys()
                and all(
                    float_lists_agree(values, l_refinements[refine])
                    for refine, values in n_refinements.items()
                ),
            }
        }
    else:
        n_affix = n.get("affix", [])
        n_name = n_affix[0].get("name") if n_affix else None
        l_info = l.get("info", {})
        values = {"name": [n_name, l_info.get("setName")]}
        n_bonuses = {
            f"{2 if value.get('level') == 0 else 4}pc": value.get(
                "param_list", []
            )
            for value in n_affix
        }
        l_bonuses = {
            key: value.get("params", [])
            for key, value in l_info.get("setBonuses", {}).items()
        }
        combat = {
            "set_bonus_params": {
                "nanoka": n_bonuses,
                "lunaris": l_bonuses,
                "agrees": n_bonuses.keys() == l_bonuses.keys()
                and all(
                    float_lists_agree(values, l_bonuses[key])
                    for key, values in n_bonuses.items()
                ),
            }
        }
    identity = {
        key: {
            "nanoka": pair[0],
            "lunaris": pair[1],
            "agrees": pair[0] == pair[1],
        }
        for key, pair in values.items()
    }
    return {"identity": identity, "combat": combat}


def confirmed_inputs(
    kind: str, nanoka: dict[str, Any], lunaris: dict[str, Any]
) -> dict[str, Any]:
    payload = nanoka["en"]["payload"]
    if kind == "character":
        lunaris_payload = lunaris["en"]["payload"]
        return {
            "identity": {
                key: payload.get(key)
                for key in ("name", "weapon", "rarity", "element", "icon")
            },
            "base_stats": {
                key: payload.get(key)
                for key in (
                    "base_hp",
                    "base_atk",
                    "base_def",
                    "crit_rate",
                    "crit_dmg",
                    "elemental_mastery",
                    "stats_modifier",
                )
            },
            "combat_tables": {
                "skills": payload.get("skills", []),
                "passives": payload.get("passives", []),
                "constellations": payload.get("constellations", []),
                "attacks": payload.get("attack", []),
                "energy": lunaris_payload.get("energy", []),
                "energy_source": "lunaris",
                "energy_secondary_reference_nanoka": payload.get("energy", []),
            },
        }
    if kind == "weapon":
        return {
            "identity": {
                key: payload.get(key)
                for key in ("name", "weapon_type", "rarity", "icon")
            },
            "base_stats": {
                "weapon_prop": payload.get("weapon_prop", []),
                "stats_modifier": payload.get("stats_modifier", {}),
            },
            "refinements": payload.get("refinement", {}),
        }
    return {
        "identity": {
            "id": payload.get("id"),
            "icon": payload.get("icon"),
            "rank": payload.get("rank"),
        },
        "set_bonuses": payload.get("affix", []),
    }


def build_entity(
    kind: str,
    target: dict[str, Any],
    targets: dict[str, Any],
    registry: set[str],
) -> dict[str, Any]:
    nanoka, lunaris = fetch_detail(kind, target, targets)
    registered = target["slug"] in registry
    expected_registered = target["gcsim_status"] != "missing"
    if registered != expected_registered:
        raise RuntimeError(
            f"{kind} {target['id']} registry state changed: "
            f"expected registered={expected_registered}, got {registered}"
        )
    return {
        "schema_version": 1,
        "kind": kind,
        "identity": {
            "id": target["id"],
            "lunaris_id": target.get("lunaris_id", target["id"]),
            "gcsim_slug": target["slug"],
        },
        "gcsim": {
            "status": target["gcsim_status"],
            "registered": registered,
            "config_seed": {
                "use": "pipeline",
                "kind": kind,
                "name": target["slug"],
                "override_id": target["id"].split("-", 1)[0],
            },
        },
        "release_track": target.get("release_track", "live"),
        "source_comparison": identity_comparison(kind, target, nanoka, lunaris),
        "implementation_inputs": {
            "confirmed": confirmed_inputs(kind, nanoka, lunaris),
            "unresolved": UNRESOLVED[kind],
        },
        "sources": {"nanoka": nanoka, "lunaris": lunaris},
    }


def build_reaction(
    target: dict[str, Any],
    targets: dict[str, Any],
    nanoka_hyperlinks: dict[str, dict[str, Any]],
    nanoka_hyperlink_urls: dict[str, str],
    nanoka_hyperlink_hashes: dict[str, str],
) -> dict[str, Any]:
    lunaris_cfg = targets["sources"]["lunaris"]
    lunaris_sources: dict[str, Any] = {"version": lunaris_cfg["preview_version"]}
    nanoka_sources: dict[str, Any] = {
        "version": targets["sources"]["nanoka"]["preview_version"]
    }
    for language, lunaris_language in (("en", "en"), ("zh", "chs")):
        keywords = target["keywords"][language]
        matching = {
            key: value
            for key, value in nanoka_hyperlinks[language].items()
            if any(
                keyword.casefold()
                in json.dumps(value, ensure_ascii=False).casefold()
                for keyword in keywords
            )
        }
        nanoka_sources[language] = source_record(
            nanoka_hyperlink_urls[language],
            matching,
            nanoka_hyperlink_hashes[language],
        )
        url = (
            f"{lunaris_cfg['base_url']}/{lunaris_cfg['preview_version']}/"
            f"{lunaris_language}/tutorial/{target['tutorial_id']}.json"
        )
        value, sha256 = fetch_json(url)
        lunaris_sources[language] = source_record(url, value, sha256)
    return {
        "schema_version": 1,
        "kind": "reaction",
        "identity": {
            "tutorial_id": target["tutorial_id"],
            "gcsim_slug": target["slug"],
        },
        "gcsim": {"status": target["gcsim_status"], "registered": False},
        "implementation_inputs": {
            "confirmed": {
                "tutorial": lunaris_sources["en"]["payload"],
                "related_hyperlinks": nanoka_sources["en"]["payload"],
            },
            "unresolved": UNRESOLVED["reaction"],
        },
        "sources": {"nanoka": nanoka_sources, "lunaris": lunaris_sources},
    }


def describe_artifact_gap(
    artifact_id: str, targets: dict[str, Any]
) -> str:
    target_ids = {item["id"] for item in targets["artifacts"]}
    if artifact_id in target_ids:
        return "current_target"
    if artifact_id in targets["legacy_artifact_gaps"]:
        return "legacy_low_priority"
    if artifact_id == "15004":
        return "excluded_unreleased"
    return "unclassified"


def build_outputs(targets: dict[str, Any]) -> dict[Path, bytes]:
    nanoka_cfg = targets["sources"]["nanoka"]
    lunaris_cfg = targets["sources"]["lunaris"]
    version_dir = HERE / "data" / nanoka_cfg["preview_version"]
    outputs: dict[Path, bytes] = {}
    entity_values: dict[tuple[str, str], dict[str, Any]] = {}

    registries = {
        "character": read_registry_names(
            REPO_ROOT / "pkg/core/keys/character.dm.go", "Char"
        ),
        "weapon": read_registry_names(
            REPO_ROOT / "pkg/core/keys/weapon.dm.go", "Weapon"
        ),
        "artifact": read_registry_names(
            REPO_ROOT / "pkg/core/keys/artifact.dm.go", "Set"
        ),
    }

    jobs: list[
        tuple[
            str,
            str,
            dict[str, Any],
            concurrent.futures.Future[dict[str, Any]],
        ]
    ] = []
    with concurrent.futures.ThreadPoolExecutor(max_workers=6) as executor:
        for plural, kind in (
            ("characters", "character"),
            ("weapons", "weapon"),
            ("artifacts", "artifact"),
        ):
            for target in targets[plural]:
                future = executor.submit(
                    build_entity, kind, target, targets, registries[kind]
                )
                jobs.append((plural, kind, target, future))
        for plural, _kind, target, future in jobs:
            value = future.result()
            entity_values[(_kind, target["id"])] = value
            outputs[version_dir / plural / f"{target['id']}.json"] = canonical_json(
                value
            )

    nanoka_hyperlinks: dict[str, dict[str, Any]] = {}
    nanoka_hyperlink_urls: dict[str, str] = {}
    nanoka_hyperlink_hashes: dict[str, str] = {}
    for language in ("en", "zh"):
        url = (
            f"{nanoka_cfg['base_url']}/{nanoka_cfg['preview_version']}/"
            f"{language}/hyperlink.json"
        )
        value, sha256 = fetch_json(url)
        nanoka_hyperlinks[language] = value
        nanoka_hyperlink_urls[language] = url
        nanoka_hyperlink_hashes[language] = sha256
    for target in targets["reactions"]:
        value = build_reaction(
            target,
            targets,
            nanoka_hyperlinks,
            nanoka_hyperlink_urls,
            nanoka_hyperlink_hashes,
        )
        outputs[
            version_dir / "reactions" / f"{target['slug']}.json"
        ] = canonical_json(value)

    indexes: dict[str, Any] = {}
    index_sources: dict[str, Any] = {}
    for kind in ("character", "weapon", "artifact"):
        url = (
            f"{nanoka_cfg['base_url']}/{nanoka_cfg['preview_version']}/{kind}.json"
        )
        value, sha256 = fetch_json(url)
        indexes[kind] = value
        index_sources[kind] = {
            "url": url,
            "sha256": sha256,
            "count": len(value),
        }

    version_url = f"{lunaris_cfg['base_url']}/version.json"
    lunaris_version, lunaris_version_hash = fetch_json(version_url)
    artifact_gaps = []
    for artifact_id, value in indexes["artifact"].items():
        first = next(iter(value["set"].values()))
        slug = slugify(first["name"]["en"])
        if slug in registries["artifact"]:
            continue
        artifact_gaps.append(
            {
                "id": artifact_id,
                "name": first["name"],
                "classification": describe_artifact_gap(artifact_id, targets),
            }
        )
    energy_timing_conflicts = []
    for (kind, entity_id), value in entity_values.items():
        if kind != "character":
            continue
        reports = value["source_comparison"]["combat"]["energy_reports"]
        conflicts = [
            {
                "source": item["source"],
                "nanoka_cd": item["nanoka"].get("cd"),
                "lunaris_cd": item["lunaris"].get("cd"),
            }
            for item in reports["matched"]
            if not item["cd_agrees"]
        ]
        if conflicts or reports["unmatched_nanoka"] or reports["unmatched_lunaris"]:
            energy_timing_conflicts.append(
                {
                    "id": entity_id,
                    "conflicts": conflicts,
                    "unmatched_nanoka": reports["unmatched_nanoka"],
                    "unmatched_lunaris": reports["unmatched_lunaris"],
                }
            )

    manifest = {
        "schema_version": 1,
        "data_track": "preview",
        "source_policy": {
            "character_energy_default": "lunaris",
            "note": (
                "Use Lunaris particle count, chance, and cooldown until verified "
                "measurements are supplied. Keep Nanoka as a secondary reference."
            ),
        },
        "versions": {
            "nanoka": nanoka_cfg,
            "lunaris": lunaris_cfg,
            "lunaris_reported": {
                "url": version_url,
                "sha256": lunaris_version_hash,
                "version": lunaris_version.get("version"),
                "timestamp": lunaris_version.get("timestamp"),
            },
        },
        "source_indexes": index_sources,
        "targets": {
            plural: [
                {
                    **{
                        key: value
                        for key, value in target.items()
                        if key not in ("keywords",)
                    },
                    "registered": target["slug"] in registries[kind],
                }
                for target in targets[plural]
            ]
            for plural, kind in (
                ("characters", "character"),
                ("weapons", "weapon"),
                ("artifacts", "artifact"),
            )
        },
        "reactions": targets["reactions"],
        "supplemental_findings": {
            "missing_cryo_traveler_counterpart": "10000005-5 (Aether)",
            "already_implemented": [
                "10000128 Varka (registered, explicitly incomplete)",
                "12432 Flame-Forged Insight",
            ],
            "source_conflicts": {
                "character_energy_timing": energy_timing_conflicts
            },
            "artifact_catalog_gaps": artifact_gaps,
            "excluded_entities": targets["excluded_entities"],
        },
        "promotion_rule": (
            "Do not promote a snapshot to an active simulation implementation "
            "until every unresolved runtime field used by that implementation "
            "has a cited measurement, test, or explicit documented assumption."
        ),
    }
    outputs[version_dir / "manifest.json"] = canonical_json(manifest)
    return outputs


def apply_outputs(outputs: dict[Path, bytes], check: bool) -> int:
    changed: list[Path] = []
    for path, expected in sorted(outputs.items(), key=lambda item: str(item[0])):
        actual = path.read_bytes() if path.exists() else None
        if actual == expected:
            continue
        changed.append(path)
        if not check:
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_bytes(expected)
    if changed:
        action = "would update" if check else "updated"
        for path in changed:
            print(f"{action}: {path.relative_to(REPO_ROOT)}")
    if check and changed:
        return 1
    print(f"validated {len(outputs)} generated community-data files")
    return 0


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--check",
        action="store_true",
        help="compare freshly fetched data with the committed snapshots",
    )
    args = parser.parse_args()
    targets = load_json(TARGETS_PATH)
    try:
        outputs = build_outputs(targets)
    except (RuntimeError, SourceError) as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 2
    return apply_outputs(outputs, args.check)


if __name__ == "__main__":
    raise SystemExit(main())
