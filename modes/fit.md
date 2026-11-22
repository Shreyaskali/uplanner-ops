# Mode: fit

Score a job description against the structured profile.

```bash
uplanner fit --json --text-file <path>
```

The Go core returns a score, requirement checks, strengths, gaps and next
moves. Explain the result in the user's language. Say plainly when a
requirement is a hard gap versus a wording issue. Never raise the score.

If two or more evaluations exist, offer `gaps` to find the recurring pattern.
Never apply on the user's behalf.
