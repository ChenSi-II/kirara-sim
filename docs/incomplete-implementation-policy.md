# Incomplete Implementation Policy

An implementation being loadable is not the same as it being mechanically
complete.

## Merge-blocking omissions

A content PR is not ready when any missing behavior can change which damage
formula, event ordering, owner, target, attack type, or stack state applies.
This includes:

- a missing reaction event or reaction damage tag;
- an unknown reaction formula or aura-consumption rule;
- ambiguous ownership that could let teammate actions trigger the passive;
- unknown hitlag, snapshot, or trigger ordering that changes the result;
- an unavailable engine mechanism required by the main passive.

Such content is marked `mechanics-blocked` in the content matrix. Data and
research scaffolding may be reviewed separately, but the weapon must not be
documented as fully implemented.

## Allowed documented deviations

A deviation may be accepted only when it does not select a materially
different mechanic and all of the following are present:

- an explicit docs issue describing the approximation;
- evidence and game version;
- a parameter or isolated function that can be replaced without rewriting the
  package;
- a focused test locking the current convention;
- reviewer agreement that the deviation is non-blocking.

## Stellar reaction dependency

Stellar Glimmer, Stellar Conduct, and Stellar Swirl references are currently
mechanics-blocked because the repository has neither their approved executable
specifications nor their final event/tag/formula implementation.

Weapon packages may implement independent clauses such as permanent stats,
ordinary elemental-reaction buffs, or non-Stellar stacks. Those clauses must
be described as partial behavior and must not use an existing reaction as a
silent substitute.
