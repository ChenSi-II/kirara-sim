# Stellar Reactions — Executable Specification Gate

Status: **research required; implementation blocked**

This document covers the two pending core reactions currently referred to as
Stellar Conduct and Stellar Swirl. Stellar Glimmer is treated as a shared
weapon/content dependency until research determines whether it is an umbrella
classification, a distinct reaction, or a damage-category label.

No placeholder formula may be merged into `pkg/reactable` or weapon passives.

## Shared questions

- Triggering elements and forward/reverse forms.
- Aura coexistence, consumption, residual aura, decay, refresh, and
  conversion.
- Priority inside `Reactable.React`, including competition with existing
  reactions.
- Enabling rules and whether a character flag is required.
- Trigger owner, damage owner, and behavior after swap or death.
- Trigger event versus damage event, including exact event arguments.
- Whether damage can crit and whether DEF, RES, DMG bonus, EM, level, and
  reaction bonus apply.
- ICD/GCD, first damage frame, later ticks, hitlag, and snapshot behavior.
- Single/multi-target selection, range, chain rules, and target death.
- Interactions with existing characters, weapons, artifacts, resonance, and
  statistics.

Each resolved constant must record source, game version, confidence, and
reviewer.

## Stellar Conduct

Formula: **TBD — user/research team will provide**

Required test vectors:

- forward and reverse trigger;
- zero, exactly consumed, and partially remaining aura;
- level/EM/reaction-bonus table;
- crit/non-crit if applicable;
- one, two, and three targets at boundary distances;
- owner attribution, swap, refresh, and target death;
- event emitted exactly once.

## Stellar Swirl

Formula: **TBD — user/research team will provide**

The specification must additionally answer:

- which source element is spread;
- whether spread damage can trigger another reaction;
- target selection order and maximum target count;
- self-hit or repeated-hit prevention;
- interaction with ordinary Swirl bonuses and Viridescent Venerer;
- whether its damage is a new `AttackTag` or a modifier of an existing event.

Required tests mirror Stellar Conduct and add chained-target ordering and
ordinary-Swirl regression coverage.

## Merge sequence

1. Research evidence and formula tables.
2. Specification review by at least two non-authors.
3. Test fixtures that fail before implementation.
4. Core reaction implementation.
5. Existing-reaction and fixed-seed regression comparison.
6. Only then enable dependent weapon, artifact, and character clauses.
