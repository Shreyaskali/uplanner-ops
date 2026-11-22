# Mode: doctor

Run the setup check and explain anything that is not OK.

```bash
uplanner doctor --json
```

Report each check as OK / WARN / FAIL. Never print a key value. If generated
content is missing, tell the user to run `go run ./cmd/gen-data` with the two
source paths. If the LLM key is unset, say that offline modes still work.
