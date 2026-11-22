# Mode: plan-learning

Turn a named gap into a small, free learning plan.

```bash
uplanner plan-learning --gap "SQL and dashboards" --limit 6 --json
# or: uplanner plan-learning --domain data,ai --limit 6 --json
```

The Go core maps the gap text onto domains, excludes resources already marked
done, and returns a bounded set with approximate hours. Explain the order and
why each item is there. Durations are approximate; "Ongoing" resources are not
summed into the hours estimate.

Keep the plan small. A 60-hour course plus two short articles is a better
first week than twenty bookmarks. Offer `progress` to mark items done.
