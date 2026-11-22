# Modes: cost, deadlines, track, export

## cost

```bash
uplanner cost --months 24 --pay 18 --tuition 20 --json
```

A planning scenario before tax, raises and scholarships. Use the user's exact
numbers. Never convert Indian CTC, MEP and foreign base salary as if they were
the same measure.

## deadlines

```bash
uplanner deadlines --json
```

Lists upcoming application dates from the tracker, soonest first. Past dates
remain visible as overdue until the user changes the status. Deadline times
are in the university's timezone, not IST; verify before relying on them.

## track

```bash
uplanner track --json
uplanner track --id <id> --status preparing --json      # gated
uplanner track --id <id> --status preparing --confirm --json
```

Never write without an explicit confirmation. The canonical tracker is
`data/applications.md`; the write is atomic. Status vocabulary: not started,
preparing, ready, submitted, interview, offer, rejected, withdrawn.

## export

```bash
uplanner export --out plan.md --json
```

Produces a Markdown application plan from the canonical tracker. This is a
review artifact, not a second source of truth. Verify every date on the
official programme page before relying on it.
