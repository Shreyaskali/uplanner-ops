# uplanner-ops

A local-first, deterministic career-planning CLI. It combines the degree and
hiring research of **uplanner** with the free learning resources of
**Intersections**, using the agent architecture of
[career-ops](https://github.com/career-ops-hq/career-ops).

**Status: complete.** All planned modes work locally: profile, fit, gaps,
market, path, free learning, cert, degree planning, deadlines, tracking,
export and a weekly plan. The single outbound boundary has a fail-closed PII
gate. The Go fit engine is byte-for-byte equivalent to the web tool's
`compass-core.js` across the parity fixtures. Multi-CLI wrappers and a
system-layer-only self-update are included.

## Principles

- **Local-first** — no account or server for the core tool.
- **Deterministic-first** — scoring is Go code with tests, not prompt output.
- **Human-in-the-loop** — never applies, enrols, submits, buys or sends.
- **Free-first** — only free learning resources are generated.
- **One correct datum** — your exact numbers, never averaged or invented.

## Install

Prebuilt binaries for macOS, Linux and Windows are on the
[Releases page](https://github.com/Shreyaskali/uplanner-ops/releases).

```bash
# macOS / Linux
curl -L https://github.com/Shreyaskali/uplanner-ops/releases/latest/download/uplanner-ops_1.0.0_darwin_arm64.tar.gz | tar xz
./uplanner doctor

# Or with Go 1.25+
go install github.com/shreyaskali/uplanner-ops/cmd/uplanner@latest
```

From source:

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
./uplanner fit --text-file job.md          # score a JD against your profile
./uplanner fit --text-file job.md --save --confirm   # save a gated report
./uplanner resources --domain ai,data      # free learning resources
./uplanner plan-learning --gap "SQL"       # a small free study plan
./uplanner progress --done <id> --confirm  # mark a resource done
./uplanner degree-plan --country India     # rank programs by evidence
./uplanner compare <id-a> <id-b>           # side-by-side comparison
./uplanner deadlines                       # upcoming application dates
./uplanner track --id <id> --status preparing --confirm
./uplanner export --out plan.md            # markdown application plan
./uplanner tui                             # Bubble Tea interface
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
internal/profile   structured profile loader
internal/core/fit  deterministic JD fit engine (Go port of compass-core.js)
internal/core/learn free-resource selector and gap→domain mapping
internal/core/degree program ranking and side-by-side comparison
internal/tracker    canonical applications.md with atomic writes
internal/outbound   single network boundary with a fail-closed PII gate
content/           generated, never hand-edited
```

## Tests and parity

```bash
go test ./...
./scripts/check-parity.sh "/path/to/University Planner"
```

The parity script runs four fixtures through both the Go engine
(`cmd/parity-go`) and the web tool's `compass-core.js`
(`parity/js-engine.mjs`) and compares the parsed results. Score, band,
confidence, dimensions, requirement statuses and matched keywords must match.

## Self-update

```bash
uplanner update --source /path/to/uplanner-ops --json   # plan only
uplanner update --source /path/to/uplanner-ops --apply   # apply
```

Only files in `internal/update.SystemPaths` may be replaced. User data and
secrets are never touched; local edits are preserved as `.bak` and skipped.

## Roadmap

1. **Foundation** — contract, generator, doctor, profile, TUI shell. ✓
2. **Gap engine** — deterministic fit engine with profile-driven checks. ✓
   Remaining: `gaps`, `market`, `path`, plus Go↔JS parity fixtures.
3. **Learning** — free resource planning, paths and progress. ✓
   Remaining: `cert` exam plans.
4. **Degree & execution** — degree-plan, compare, cost, deadlines, track, export. ✓
   Remaining: weekly `plan`.
5. **Hardening** — gaps/market/path/cert/plan, BYOK gate, Go↔JS parity,
   multi-CLI wrappers, system-layer-only self-update. ✓
