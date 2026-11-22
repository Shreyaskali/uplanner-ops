# Modes: gaps, market, path, cert, plan

## gaps

```bash
uplanner gaps --json
```

Aggregates requirement statuses from saved fit reports and names the
recurring ones. A single appearance is a caveat; three appearances are a
pattern worth acting on. Never change a status to make a pattern look better.

## market

```bash
uplanner market --domain ai --json
```

Shows what a role actually involves, what a hiring manager screens for, the
evidence that counts and the common gaps. The demand note is an editorial
estimate, not measured data — say so when quoting it.

## path

```bash
uplanner path --json
```

Maps the recorded field, skills and role onto role profiles. It is a planning
signal from the user's own evidence, not an offer probability. If the profile
is empty, ask for it instead of guessing.

## cert

```bash
uplanner cert --exam cat --json
```

Gives relative preparation phases. Exam windows are marked unverified where
they are not confirmed. Never present a guessed date as fact and never
invent a score.

## plan

```bash
uplanner plan --days 7 --limit 2 --json
```

Combines upcoming tracker deadlines with a small free-resource set per
deadline. Keep it capacity-aware and never add more work than the user can
do. Verify every date on the official programme page.
