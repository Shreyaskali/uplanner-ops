# Mode: fit

Score a job description against the structured profile. The Go core is
deterministic: never change its score, band or requirement statuses.

```bash
uplanner fit --text-file <path> --json
# or: uplanner fit --text '<job description>' --json
# or: piped stdin
```

The result contains a score, band, matched dimensions, explicit requirement
checks, gaps and next moves. Explain it in the user's language. Say plainly
when a requirement is a hard gap versus a wording issue. Never raise the
score and never claim a credential the profile does not record.

## Saving

Writes to the user layer are gated. Without `--confirm` the command returns:

```
needs_confirmation: fit_write <n> characters - rerun with --confirm
```

Show the result, ask the user, and only rerun with `--confirm` after an
explicit yes. Reports land in `data/fit/` as Markdown and are review
artifacts, not a second source of truth.

If two or more evaluations exist, offer `gaps` to find the recurring pattern.
Never apply on the user's behalf.
