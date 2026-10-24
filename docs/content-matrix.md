# Community Content Matrix

This matrix is the delivery gate for community-preview content. A row is not
`complete` merely because its package or generated data exists: data,
registration, behavior, tests, evidence, and known deviations must all be
accounted for.

Status meanings:

- `ready`: independent of unimplemented core mechanics and ready for a
  dedicated content PR.
- `mechanics-blocked`: data may be generated, but the passive must not be
  presented as complete until the referenced reaction specification and core
  implementation have landed.
- `existing`: already implemented before this work.

The implementation, research, and review owners must be different people for
reaction-dependent behavior. Unassigned owners are written as `TBD` rather
than silently assigning approval to the author.

| Game ID | Internal key | Type | Data source | Mechanism dependency | Code status | Test status | Evidence | Implementation | Research | Review |
|---:|---|---|---|---|---|---|---|---|---|---|
| 11435 | `hereticsmoltenblade` | sword | community 6.7.54 | movement/path accumulation | ready | local prototype passed | community JSON; Lunaris | TBD | TBD | TBD |
| 11436 | `emberwell` | sword | community 6.7.54 | Stellar Glimmer | mechanics-blocked | non-Stellar portion tested | community JSON; Lunaris | TBD | TBD | TBD |
| 11519 | `lightbearingmoonshard` | sword | live + community comparison | Lunar Crystallize | ready | local prototype passed | live datamine; community JSON | TBD | TBD | TBD |
| 11520 | `whitelakefrostfeather` | sword | community 6.7.54 | Stellar Glimmer | mechanics-blocked | independent ATK stacks tested | community JSON; Lunaris | TBD | TBD | TBD |
| 11521 | `exaiphanesblade` | sword | community 6.7.54 | account resonance count | ready | local prototype passed | community JSON; Lunaris | TBD | TBD | TBD |
| 12432 | `flameforgedinsight` | claymore | live | none | existing | existing baseline | live datamine | existing | TBD | TBD |
| 12435 | `forgedbythegoldenmelody` | claymore | community 6.7.54 | Stellar Glimmer | mechanics-blocked | non-Stellar cycle tested | community JSON; Lunaris | TBD | TBD | TBD |
| 12436 | `bladeofatonement` | claymore | community 6.7.54 | Stellar Glimmer | mechanics-blocked | non-Stellar portion tested | community JSON; Lunaris | TBD | TBD | TBD |
| 12516 | `ateaspoonoftranscendence` | claymore | live + community comparison | Stellar Conduct; Stellar Swirl | mechanics-blocked | permanent ATK tested | live datamine; community JSON | TBD | TBD | TBD |
| 13435 | `frostbreath` | polearm | community 6.7.54 | Cryo/Hydro reaction classification | ready | local prototype passed | community JSON; Lunaris | TBD | TBD | TBD |
| 13436 | `songofthevigil` | polearm | community 6.7.54 | Stellar Glimmer | mechanics-blocked | non-Stellar energy tested | community JSON; Lunaris | TBD | TBD | TBD |
| 13517 | `disasterandremorse` | polearm | live + community comparison | Hexerei classification | ready | local prototype passed | live datamine; community JSON | TBD | TBD | TBD |
| 14435 | `clashofkings` | catalyst | community 6.7.54 | none | ready | local prototype passed | community JSON; Lunaris | TBD | TBD | TBD |
| 14436 | `echoesoftheheart` | catalyst | community 6.7.54 | Stellar Glimmer | mechanics-blocked | non-Stellar portion tested | community JSON; Lunaris | TBD | TBD | TBD |
| 15435 | `jadevista` | bow | community 6.7.54 | party element counting | ready | local prototype passed | community JSON; Lunaris | TBD | TBD | TBD |
| 15436 | `covenantoffrostandsnow` | bow | community 6.7.54 | none | ready | local prototype passed | community JSON; Lunaris | TBD | TBD | TBD |

## Pull request boundaries

The rows above must not be delivered as one content PR.

1. Land the pinned community-data loader and validation tests independently.
2. Land the common reaction-owner test helper independently if it is still
   needed by more than one content PR.
3. Deliver unique 5-star weapons one per PR.
4. Deliver simple 4-star weapons in groups of no more than 3–8 only when their
   trigger and timing semantics are genuinely shared.
5. Keep every `mechanics-blocked` passive out of a ready content PR until its
   reaction specification and core implementation are merged.
