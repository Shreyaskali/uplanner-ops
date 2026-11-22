# uplanner-ops

A local-first, deterministic career-planning CLI. It combines the degree and
hiring research of **uplanner** with the free learning resources of
**Intersections**, using the agent architecture of
[career-ops](https://github.com/career-ops-hq/career-ops).

**Status: Phase 1 foundation.** Data generation, setup doctor, structured
profile and the Bubble Tea shell are working. Scoring, learning and tracking
modes land in later phases.

## Principles

- **Local-first** — no account or server for the core tool.
- **Deterministic-first** — scoring is Go code with tests, not prompt output.
- **Human-in-the-loop** — never applies, enrols, submits, buys or sends.
- **Free-first** — only free learning resources are generated.
- **One correct datum** — your exact numbers, never averaged or invented.

## Install

Requires Go 1.25+.

```bash
git clone https://github.com/shreyaskali/uplanner-ops.git
cd uplanner-ops
go build -o uplanner ./cmd/uplanner
```

Generate the content from the two source projects:

```bash
go run ./cmd/gen-data \
  --uplanner "/path/to/University Planner" \
  --intersections "/path/to/intersections" \
  --out ./content
```

This produces `content/uplanner.json` with:
- **179** degree programs and **63** hiring reports
- **258** free learning resources across **31** paths
- **37** role profiles (what the job really is, what's screened for, evidence that counts)

## Use

```bash
./uplanner doctor --json       # setup report
./uplanner content             # generated data counts
./uplanner profile --field "Computer Science" --years 3 --skills "SQL, AWS"
./uplanner tui                 # Bubble Tea interface
```

Set `UPLANNER_DATA_DIR` to keep your data outside the install directory. An
optional `UPLANNER_LLM_KEY` enables interpretation modes later; offline modes
work without it. Secrets stay in the environment and are never printed.

## Layout

```
AGENTS.md          the contract — read before any mode
SKILL.md           router for AI coding CLIs
DATA_CONTRACT.md   system vs user layer
modes/*.md         judgment prompts
cmd/gen-data       content generator
cmd/uplanner       CLI + Bubble Tea TUI
internal/data      generated content contract
internal/config    data-root resolution
internal/doctor    setup validation
content/           generated, never hand-edited
```

## Tests

```bash
go test ./...
```

## Roadmap

1. **Foundation** (this) — contract, generator, doctor, profile, TUI shell.
2. **Gap engine** — Go ports of fit/gaps/path with golden parity tests.
3. **Learning** — free resource planning, progress, exam plans.
4. **Degree & execution** — compare, cost, deadlines, tracking, export, weekly plan.
5. **Hardening** — BYOK outbound gate, multi-CLI wrappers, self-update.
