# Community preview data

This directory stores combat-facing source data that is known before a complete
gcsim implementation exists. It currently compares:

- [原神数据库 — nanoka.cc](https://gi.nanoka.cc/)
- [Lunaris](https://lunaris.moe/)

The committed `data/6.7.54` snapshots use Nanoka `6.7.54` and Lunaris
`6.7.54.2`. They are preview data, not a claim that the content is live or
stable.

## What is stored

Each entity file contains:

- the exact source URLs and SHA-256 hashes;
- English and Simplified Chinese combat payloads from both sources;
- identity fields on which the two sources agree or disagree;
- base stats, talent/refinement/set parameters, attack metadata, ICD/gauge
  metadata, energy metadata, passives, and constellations when available;
- a list of runtime fields that still require measurement or a documented
  simulator assumption.

For character particle generation, Lunaris is the temporary implementation
default. Nanoka values remain in each file as a secondary reference and are not
discarded.

Lore, voice lines, materials, and image assets are intentionally omitted.

`manifest.json` is the review entry point. It also records corrections to the
manual gap list, excluded internal/special entities, and lower-priority legacy
artifact gaps.

These files are **implementation inputs only**. They deliberately do not create
active character, weapon, artifact, or reaction registrations. Registering a
placeholder with guessed frames, hitboxes, reaction coefficients, or event
ordering would produce plausible-looking but incorrect simulations.

## Refresh and verify

Use the project-specific Conda environment:

```sh
conda run -n genshin python pipeline/community/update.py
conda run -n genshin python pipeline/community/update.py --check
```

`--check` downloads the same versioned endpoints and fails when a source payload
or generated file has changed. Preview endpoints can legitimately change; review
the diff before accepting a refresh.

## Promotion into the simulator

For a character, promote the `config_seed` into the appropriate
`internal/characters` directory only after adding:

1. action implementations and state transitions;
2. measured frames, hit marks, cancel windows, hitlag, and geometry;
3. verified particle generation and multi-target behavior;
4. tests for talents, constellations, ICD, aura, and snapshot behavior.

Weapons and artifacts need their passive/set implementation and focused tests.
Stellar-Conduct and Stellar Swirl must first gain core reaction types, attack
tags, events, ownership rules, and verified formulas. The two new artifact sets
depend on those reaction primitives, so their 4-piece effects should not be
activated earlier.
