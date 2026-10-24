# Community Preview Data Snapshot

Community-preview generation must be reproducible and must not fetch mutable
network data during normal pipeline runs.

## Pinned snapshot

- Data track: `preview`
- Data version: `6.7.54`
- Schema version: `1`
- Intended repository path:
  `pipeline/community/data/6.7.54`
- Source manifest SHA-256:
  `317c05ced66cd7401729d028187b31e41623eac557b1861b817e528012e90110`
- Target repository baseline:
  `2a38d66e29231b57f09e04446da2905d809e51b8`
- Common upstream baseline referenced by `TODOLIST.md`:
  `45585d29d1b548b0e9ddcc78fafbd8c57d09afe6`

The snapshot records evidence originating from Lunaris and Nanoka. Network
access is allowed only while curating a new snapshot. Once reviewed, the
normalized source JSON and its manifest must be committed; generators read
only those fixed local files.

## Loader requirements

The community loader must reject a record unless all of the following hold:

- `kind == "weapon"` and the schema version is supported.
- `identity.id`, configured override ID, and config-seed override ID agree.
- Config name, config-seed name, and gcsim slug agree.
- Weapon type is recognized.
- Base stats, weapon properties, and R1–R5 refinement parameters are complete.
- A live/community collision is identical in critical identity and base-stat
  fields; otherwise generation stops with a conflict.

Live datamine remains the default. Community fallback is allowed only for an
explicitly recognized live not-found result and a single pinned local version.

## Regeneration record

Every generated-data PR must state:

- snapshot version and manifest hash;
- exact generator command;
- Go version;
- whether live/community comparison was performed;
- why every generated diff outside the intended content is expected.
