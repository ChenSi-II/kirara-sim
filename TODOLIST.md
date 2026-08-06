# Project TODO

## Missing Character Implementations

- [x] Discover missing characters programmatically from every local data version
- [x] Generate exact local multipliers and complete catalog growth/ascension data
- [x] Replace generic `basicimport` scaffolds with character-specific actions and states
- [x] Run the non-scaffold implementation audit across every imported character
- [x] Run package, action-queue, and targeted character runtime tests
- [ ] Independent user review and frame/hitbox/particle verification

Summary:

- Imported characters: 13
- Review-ready: 13
- Basic-simulation-validated: 12
- Partial-simulation: 1
- Core-support-blocked details: 1

`review-ready` means the package contains real character mechanics and passed local
runtime checks. It does not claim verified live-server frames, hitboxes, ICD, or
particle distributions; those remain explicitly listed below.

## Imported Characters


### Alyosha — 10000148

Status: review-ready
Simulation level: basic-simulation-validated

- [x] Local data, exact-label config, catalog stats, and generated multipliers
- [x] Character-specific normal, charged, and plunge attacks
- [x] Character-specific Skill, Burst, states, passives, and constellations
- [x] Independent non-scaffold audit
- [x] Minimal action-queue simulation

Sources:

- Local: `pipeline/community/data/6.7.54/characters/10000148.json`
- Lunaris fallback: https://lunaris.moe/character/10000148
- Nanoka fallback query: `Alyosha 10000148`

Older implementation/core references:

- `internal/characters/fischl`
- `internal/characters/gorou`

Remaining verification:

- [ ] action frames, hit marks, cancel windows, and hitlag
- [ ] hitbox geometry, targeting, travel time, and multi-target behavior
- [ ] runtime state-machine ordering, snapshot timing, and server-only logic
- [ ] particle observations pending verification; Lunaris is the temporary default

Review TODO:

- [ ] User review of character-specific mechanics
- [ ] Validate final damage values against an independent trusted sample
- [ ] Verify frames, hitboxes, ICD, hitlag, particle behavior, and runtime state ordering

### Iansan — 10000110

Status: review-ready
Simulation level: basic-simulation-validated

- [x] Local data, exact-label config, catalog stats, and generated multipliers
- [x] Character-specific normal, charged, and plunge attacks
- [x] Character-specific Skill, Burst, states, passives, and constellations
- [x] Independent non-scaffold audit
- [x] Minimal action-queue simulation

Sources:

- Local: `pipeline/community/data/6.7.54/characters/10000110.json`
- Lunaris fallback: https://lunaris.moe/character/10000110
- Nanoka fallback query: `Iansan 10000110`

Older implementation/core references:

- `internal/characters/varesa`
- `internal/characters/xilonen`

Remaining verification:

- [ ] action frames, hit marks, cancel windows, and hitlag
- [ ] hitbox geometry, targeting, travel time, and multi-target behavior
- [ ] runtime state-machine ordering, snapshot timing, and server-only logic
- [ ] particle observations pending verification; Lunaris is the temporary default

Review TODO:

- [ ] User review of character-specific mechanics
- [ ] Validate final damage values against an independent trusted sample
- [ ] Verify frames, hitboxes, ICD, hitlag, particle behavior, and runtime state ordering

### Ifa — 10000113

Status: review-ready
Simulation level: basic-simulation-validated

- [x] Local data, exact-label config, catalog stats, and generated multipliers
- [x] Character-specific normal, charged, and plunge attacks
- [x] Character-specific Skill, Burst, states, passives, and constellations
- [x] Independent non-scaffold audit
- [x] Minimal action-queue simulation

Sources:

- Local: `pipeline/community/data/6.7.54/characters/10000113.json`
- Lunaris fallback: https://lunaris.moe/character/10000113
- Nanoka fallback query: `Ifa 10000113`

Older implementation/core references:

- `internal/characters/chasca`
- `internal/characters/sayu`

Remaining verification:

- [ ] action frames, hit marks, cancel windows, and hitlag
- [ ] hitbox geometry, targeting, travel time, and multi-target behavior
- [ ] runtime state-machine ordering, snapshot timing, and server-only logic
- [ ] particle observations pending verification; Lunaris is the temporary default

Review TODO:

- [ ] User review of character-specific mechanics
- [ ] Validate final damage values against an independent trusted sample
- [ ] Verify frames, hitboxes, ICD, hitlag, particle behavior, and runtime state ordering

### Illuga — 10000127

Status: review-ready
Simulation level: basic-simulation-validated

- [x] Local data, exact-label config, catalog stats, and generated multipliers
- [x] Character-specific normal, charged, and plunge attacks
- [x] Character-specific Skill, Burst, states, passives, and constellations
- [x] Independent non-scaffold audit
- [x] Minimal action-queue simulation

Sources:

- Local: `pipeline/community/data/6.7.54/characters/10000127.json`
- Lunaris fallback: https://lunaris.moe/character/10000127
- Nanoka fallback query: `Illuga 10000127`

Older implementation/core references:

- `internal/characters/albedo`
- `internal/characters/gorou`

Remaining verification:

- [ ] action frames, hit marks, cancel windows, and hitlag
- [ ] hitbox geometry, targeting, travel time, and multi-target behavior
- [ ] runtime state-machine ordering, snapshot timing, and server-only logic
- [ ] particle observations pending verification; Lunaris is the temporary default

Review TODO:

- [ ] User review of character-specific mechanics
- [ ] Validate final damage values against an independent trusted sample
- [ ] Verify frames, hitboxes, ICD, hitlag, particle behavior, and runtime state ordering

### Jahoda — 10000124

Status: review-ready
Simulation level: basic-simulation-validated

- [x] Local data, exact-label config, catalog stats, and generated multipliers
- [x] Character-specific normal, charged, and plunge attacks
- [x] Character-specific Skill, Burst, states, passives, and constellations
- [x] Independent non-scaffold audit
- [x] Minimal action-queue simulation

Sources:

- Local: `pipeline/community/data/6.7.54/characters/10000124.json`
- Lunaris fallback: https://lunaris.moe/character/10000124
- Nanoka fallback query: `Jahoda 10000124`

Older implementation/core references:

- `internal/characters/heizou`
- `internal/characters/sayu`

Remaining verification:

- [ ] action frames, hit marks, cancel windows, and hitlag
- [ ] hitbox geometry, targeting, travel time, and multi-target behavior
- [ ] runtime state-machine ordering, snapshot timing, and server-only logic
- [ ] particle observations pending verification; Lunaris is the temporary default

Review TODO:

- [ ] User review of character-specific mechanics
- [ ] Validate final damage values against an independent trusted sample
- [ ] Verify frames, hitboxes, ICD, hitlag, particle behavior, and runtime state ordering

### Kachina — 10000100

Status: review-ready
Simulation level: basic-simulation-validated

- [x] Local data, exact-label config, catalog stats, and generated multipliers
- [x] Character-specific normal, charged, and plunge attacks
- [x] Character-specific Skill, Burst, states, passives, and constellations
- [x] Independent non-scaffold audit
- [x] Minimal action-queue simulation

Sources:

- Local: `pipeline/community/data/6.7.54/characters/10000100.json`
- Lunaris fallback: https://lunaris.moe/character/10000100
- Nanoka fallback query: `Kachina 10000100`

Older implementation/core references:

- `internal/characters/albedo`
- `internal/characters/gorou`

Remaining verification:

- [ ] action frames, hit marks, cancel windows, and hitlag
- [ ] hitbox geometry, targeting, travel time, and multi-target behavior
- [ ] runtime state-machine ordering, snapshot timing, and server-only logic
- [ ] particle observations pending verification; Lunaris is the temporary default

Review TODO:

- [ ] User review of character-specific mechanics
- [ ] Validate final damage values against an independent trusted sample
- [ ] Verify frames, hitboxes, ICD, hitlag, particle behavior, and runtime state ordering

### Linnea — 10000130

Status: review-ready
Simulation level: basic-simulation-validated

- [x] Local data, exact-label config, catalog stats, and generated multipliers
- [x] Character-specific normal, charged, and plunge attacks
- [x] Character-specific Skill, Burst, states, passives, and constellations
- [x] Independent non-scaffold audit
- [x] Minimal action-queue simulation

Sources:

- Local: `pipeline/community/data/6.7.54/characters/10000130.json`
- Lunaris fallback: https://lunaris.moe/character/10000130
- Nanoka fallback query: `Linnea 10000130`

Older implementation/core references:

- `internal/characters/columbina`
- `internal/characters/flins`
- `internal/characters/gorou`
- `internal/characters/kokomi`
- `pkg/reactable/lunarcrystallize.go`

Remaining verification:

- [ ] action frames, hit marks, cancel windows, and hitlag
- [ ] hitbox geometry, targeting, travel time, and multi-target behavior
- [ ] runtime state-machine ordering, snapshot timing, and server-only logic
- [ ] particle observations pending verification; Lunaris is the temporary default

Review TODO:

- [ ] User review of character-specific mechanics
- [ ] Validate final damage values against an independent trusted sample
- [ ] Verify frames, hitboxes, ICD, hitlag, particle behavior, and runtime state ordering

### Lohen — 10000129

Status: review-ready
Simulation level: basic-simulation-validated

- [x] Local data, exact-label config, catalog stats, and generated multipliers
- [x] Character-specific normal, charged, and plunge attacks
- [x] Character-specific Skill, Burst, states, passives, and constellations
- [x] Independent non-scaffold audit
- [x] Minimal action-queue simulation

Sources:

- Local: `pipeline/community/data/6.7.54/characters/10000129.json`
- Lunaris fallback: https://lunaris.moe/character/10000129
- Nanoka fallback query: `Lohen 10000129`

Older implementation/core references:

- `internal/characters/ayaka`
- `internal/characters/tartaglia`

Remaining verification:

- [ ] action frames, hit marks, cancel windows, and hitlag
- [ ] hitbox geometry, targeting, travel time, and multi-target behavior
- [ ] runtime state-machine ordering, snapshot timing, and server-only logic
- [ ] particle observations pending verification; Lunaris is the temporary default

Review TODO:

- [ ] User review of character-specific mechanics
- [ ] Validate final damage values against an independent trusted sample
- [ ] Verify frames, hitboxes, ICD, hitlag, particle behavior, and runtime state ordering

### Nefer — 10000122

Status: review-ready
Simulation level: partial-simulation

- [x] Local data, exact-label config, catalog stats, and generated multipliers
- [x] Character-specific normal, charged, and plunge attacks
- [x] Character-specific Skill, Burst, states, passives, and constellations
- [x] Independent non-scaffold audit
- [x] Minimal action-queue simulation

Sources:

- Local: `pipeline/community/data/6.7.54/characters/10000122.json`
- Lunaris fallback: https://lunaris.moe/character/10000122
- Nanoka fallback query: `Nefer 10000122`

Older implementation/core references:

- `internal/characters/columbina`
- `internal/characters/nahida`
- `pkg/reactable/bloom.go`

Remaining verification:

- [ ] action frames, hit marks, cancel windows, and hitlag
- [ ] hitbox geometry, targeting, travel time, and multi-target behavior
- [ ] runtime state-machine ordering, snapshot timing, and server-only logic
- [ ] particle observations pending verification; Lunaris is the temporary default

Core support still required:

- [ ] Seed of Deceit entity

Review TODO:

- [ ] User review of character-specific mechanics
- [ ] Validate final damage values against an independent trusted sample
- [ ] Verify frames, hitboxes, ICD, hitlag, particle behavior, and runtime state ordering
- [ ] Implement Seed of Deceit entity

### Odette — 10000150

Status: review-ready
Simulation level: basic-simulation-validated

- [x] Local data, exact-label config, catalog stats, and generated multipliers
- [x] Character-specific normal, charged, and plunge attacks
- [x] Character-specific Skill, Burst, states, passives, and constellations
- [x] Independent non-scaffold audit
- [x] Minimal action-queue simulation

Sources:

- Local: `pipeline/community/data/6.7.54/characters/10000150.json`
- Lunaris fallback: https://lunaris.moe/character/10000150
- Nanoka fallback query: `Odette 10000150`

Older implementation/core references:

- `internal/characters/fischl`
- `pkg/core/star_reaction.go`

Remaining verification:

- [ ] action frames, hit marks, cancel windows, and hitlag
- [ ] hitbox geometry, targeting, travel time, and multi-target behavior
- [ ] runtime state-machine ordering, snapshot timing, and server-only logic
- [ ] particle observations pending verification; Lunaris is the temporary default

Review TODO:

- [ ] User review of character-specific mechanics
- [ ] Validate final damage values against an independent trusted sample
- [ ] Verify frames, hitboxes, ICD, hitlag, particle behavior, and runtime state ordering

### Prune — 10000132

Status: review-ready
Simulation level: basic-simulation-validated

- [x] Local data, exact-label config, catalog stats, and generated multipliers
- [x] Character-specific normal, charged, and plunge attacks
- [x] Character-specific Skill, Burst, states, passives, and constellations
- [x] Independent non-scaffold audit
- [x] Minimal action-queue simulation

Sources:

- Local: `pipeline/community/data/6.7.54/characters/10000132.json`
- Lunaris fallback: https://lunaris.moe/character/10000132
- Nanoka fallback query: `Prune 10000132`

Older implementation/core references:

- `internal/characters/heizou`
- `internal/characters/sayu`

Remaining verification:

- [ ] action frames, hit marks, cancel windows, and hitlag
- [ ] hitbox geometry, targeting, travel time, and multi-target behavior
- [ ] runtime state-machine ordering, snapshot timing, and server-only logic
- [ ] particle observations pending verification; Lunaris is the temporary default

Review TODO:

- [ ] User review of character-specific mechanics
- [ ] Validate final damage values against an independent trusted sample
- [ ] Verify frames, hitboxes, ICD, hitlag, particle behavior, and runtime state ordering

### Sandrone — 10000133

Status: review-ready
Simulation level: basic-simulation-validated

- [x] Local data, exact-label config, catalog stats, and generated multipliers
- [x] Character-specific normal, charged, and plunge attacks
- [x] Character-specific Skill, Burst, states, passives, and constellations
- [x] Independent non-scaffold audit
- [x] Minimal action-queue simulation

Sources:

- Local: `pipeline/community/data/6.7.54/characters/10000133.json`
- Lunaris fallback: https://lunaris.moe/character/10000133
- Nanoka fallback query: `Sandrone 10000133`

Older implementation/core references:

- `internal/characters/fischl`
- `pkg/core/star_reaction.go`

Remaining verification:

- [ ] action frames, hit marks, cancel windows, and hitlag
- [ ] hitbox geometry, targeting, travel time, and multi-target behavior
- [ ] runtime state-machine ordering, snapshot timing, and server-only logic
- [ ] particle observations pending verification; Lunaris is the temporary default

Review TODO:

- [ ] User review of character-specific mechanics
- [ ] Validate final damage values against an independent trusted sample
- [ ] Verify frames, hitboxes, ICD, hitlag, particle behavior, and runtime state ordering

### Zibai — 10000126

Status: review-ready
Simulation level: basic-simulation-validated

- [x] Local data, exact-label config, catalog stats, and generated multipliers
- [x] Character-specific normal, charged, and plunge attacks
- [x] Character-specific Skill, Burst, states, passives, and constellations
- [x] Independent non-scaffold audit
- [x] Minimal action-queue simulation

Sources:

- Local: `pipeline/community/data/6.7.54/characters/10000126.json`
- Lunaris fallback: https://lunaris.moe/character/10000126
- Nanoka fallback query: `Zibai 10000126`

Older implementation/core references:

- `internal/characters/albedo`
- `pkg/reactable/lunarcrystallize.go`

Remaining verification:

- [ ] action frames, hit marks, cancel windows, and hitlag
- [ ] hitbox geometry, targeting, travel time, and multi-target behavior
- [ ] runtime state-machine ordering, snapshot timing, and server-only logic
- [ ] particle observations pending verification; Lunaris is the temporary default

Review TODO:

- [ ] User review of character-specific mechanics
- [ ] Validate final damage values against an independent trusted sample
- [ ] Verify frames, hitboxes, ICD, hitlag, particle behavior, and runtime state ordering

## Kirara Non-Character Import

Source: `ChenSi-II/kirara-sim` commits `cf027d533` and `560609444`.

- [x] Star Superconduct core reaction, domain state, contribution damage, events, and tests
- [x] Star Diffusion core reaction, vortex state, contribution damage, events, and tests
- [x] Scroll of the Hero of Cinder City and Thundering Fury reaction integration
- [x] Import Kirara artifact and weapon basic-data snapshots
- [x] Adapt Star reaction character detection to the current catalog API
- [x] Core, reactable, combat, and affected artifact package tests
- [x] Full `internal/...` regression suite (backend-only Docker tests excluded)

Excluded from this import: Kirara character implementations, character basic data, character mechanism spreadsheet, and Kirara's project TODOLIST.


## Remaining Core Support

- [ ] Seed of Deceit entity
