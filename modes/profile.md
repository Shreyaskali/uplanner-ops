# Mode: profile

Populate the structured profile from a conversation, then write it with the
CLI. All fields are optional.

Ask only for what is missing. Use the user's exact numbers. Do not infer a
score, CGPA or year from anything else.

```bash
uplanner profile --degree "B.Tech" --field "Production Engineering" \
  --years 2 --role "Cloud strategy analyst" --skills "SQL, AWS" --json
```

Show the resulting fields and ask before overwriting an existing value.
