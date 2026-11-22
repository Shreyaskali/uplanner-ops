# Mode: progress

Show or update learning progress.

```bash
uplanner progress --json                  # list completed resource ids
uplanner progress --done <id> --json      # gated: returns needs_confirmation
uplanner progress --done <id> --confirm --json
```

Never write without an explicit confirmation. Show the gated token to the
user, ask, and only rerun with `--confirm` after a yes. `progress` writes only
`data/learning.md` in the user layer.
