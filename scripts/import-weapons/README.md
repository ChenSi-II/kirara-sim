# Weapon JSON importer

This adapter audits the per-ID files in `Weapons/` against gcsim's weapon
packages and converts their base stats and ascension data to
`data_gen.textproto`.

Audit without writing:

```powershell
go run ./scripts/import-weapons
```

Create packages listed in `weapons.json` and generate their static data:

```powershell
go run ./scripts/import-weapons -scaffold -write -ids 11435,11436
```

Both write modes require an explicit ID list. `-scaffold` refuses to touch an
existing package directory, while `-write` only operates on IDs that already
have a `config.yml`. Passive implementations, shared keys, shortcuts, and
simulation imports remain reviewed Go changes rather than inferred output.
