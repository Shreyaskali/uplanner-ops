// Package update manages safe self-updates of the system layer only.
// The user layer is never read for writing, never deleted and never replaced.
package update

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// SystemPaths is the allowlist of files an update may replace. Anything not
// listed here is untouchable. Generated data and user data are not updated
// here: content is rebuilt by cmd/gen-data into its own directory.
var SystemPaths = []string{
	"AGENTS.md", "SKILL.md", "DATA_CONTRACT.md", "README.md",
	"CLAUDE.md", "CODEX.md", "GEMINI.md", "QWEN.md",
	"Makefile",
	"cmd/gen-data/main.go", "cmd/gen-data/main_test.go",
	"cmd/uplanner/main.go", "cmd/uplanner/main_test.go",
	"cmd/uplanner/fit.go", "cmd/uplanner/learn.go", "cmd/uplanner/degree.go",
	"cmd/uplanner/more.go", "cmd/uplanner/tui.go", "cmd/uplanner/tui_test.go",
	"cmd/uplanner/more_test.go", "cmd/parity-go/main.go",
	"internal/config/config.go",
	"internal/data/data.go", "internal/data/data_test.go",
	"internal/profile/profile.go",
	"internal/doctor/doctor.go",
	"internal/core/fit/fit.go", "internal/core/fit/fit_test.go", "internal/core/fit/requirements.go",
	"internal/core/learn/learn.go", "internal/core/learn/learn_test.go",
	"internal/core/degree/degree.go", "internal/core/degree/degree_test.go",
	"internal/outbound/outbound.go", "internal/outbound/outbound_test.go",
	"internal/tracker/tracker.go", "internal/tracker/tracker_test.go",
	"modes/_shared.md", "modes/doctor.md", "modes/profile.md", "modes/fit.md",
	"modes/resources.md", "modes/learn-paths.md", "modes/plan-learning.md",
	"modes/progress.md", "modes/degree-plan.md", "modes/compare.md",
	"modes/degree-ops.md", "modes/review-modes.md",
	"scripts/check-parity.sh", "parity/js-engine.mjs",
	"go.mod", "go.sum", ".gitignore",
}

type Change struct {
	Path   string `json:"path"`
	Action string `json:"action"` // add | update | unchanged | blocked
	Detail string `json:"detail"`
}

// Plan compares the local install with a source checkout and lists changes.
// It never writes anything and refuses paths outside the allowlist.
func Plan(root, source string) ([]Change, error) {
	allowed := map[string]bool{}
	for _, p := range SystemPaths {
		allowed[clean(p)] = true
	}

	changes := []Change{}
	for _, rel := range SystemPaths {
		rel = clean(rel)
		src := filepath.Join(source, rel)
		local := filepath.Join(root, rel)
		want, err := os.ReadFile(src)
		if err != nil {
			if os.IsNotExist(err) {
				changes = append(changes, Change{Path: rel, Action: "unchanged", Detail: "not present in source"})
			} else {
				return nil, err
			}
			continue
		}
		have, err := os.ReadFile(local)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		if err == nil && equalHash(have, want) {
			changes = append(changes, Change{Path: rel, Action: "unchanged"})
			continue
		}
		action := "update"
		if err != nil {
			action = "add"
		}
		changes = append(changes, Change{Path: rel, Action: action, Detail: "system-owned"})
	}

	// Report any source file that looks like user data and would be rejected.
	for _, rel := range []string{"profile.yml", "resume.md", "evidence.md", "data/applications.md", "data/learning.md", ".env"} {
		if _, err := os.Stat(filepath.Join(source, rel)); err == nil {
			changes = append(changes, Change{
				Path: rel, Action: "blocked",
				Detail: "user layer or secret — never updated",
			})
		}
	}
	sort.SliceStable(changes, func(i, j int) bool { return changes[i].Path < changes[j].Path })
	return changes, nil
}

// Apply copies only allowlisted changed files. Local modifications are kept
// as .bak and block that file's replacement.
func Apply(root, source string) ([]Change, error) {
	changes, err := Plan(root, source)
	if err != nil {
		return nil, err
	}
	applied := []Change{}
	for _, change := range changes {
		if change.Action != "add" && change.Action != "update" {
			continue
		}
		rel := clean(change.Path)
		if !isAllowed(rel) {
			return nil, fmt.Errorf("refusing to update non-system path %s", rel)
		}
		src := filepath.Join(source, rel)
		local := filepath.Join(root, rel)
		want, err := os.ReadFile(src)
		if err != nil {
			return nil, err
		}
		have, err := os.ReadFile(local)
		if err == nil && !equalHash(have, want) {
			// Preserve a local edit rather than silently overwriting it.
			if err := os.WriteFile(local+".bak", have, 0o644); err != nil {
				return nil, err
			}
			applied = append(applied, Change{Path: rel, Action: "blocked", Detail: "local edit preserved as .bak; not replaced"})
			continue
		}
		if err := os.MkdirAll(filepath.Dir(local), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(local, want, 0o644); err != nil {
			return nil, err
		}
		applied = append(applied, Change{Path: rel, Action: change.Action, Detail: "updated"})
	}
	return applied, nil
}

func clean(p string) string { return strings.Trim(filepath.ToSlash(p), "/") }

func isAllowed(p string) bool {
	for _, candidate := range SystemPaths {
		if clean(candidate) == p {
			return true
		}
	}
	return false
}

func equalHash(a, b []byte) bool {
	ha, hb := sha256.Sum256(a), sha256.Sum256(b)
	return hex.EncodeToString(ha[:]) == hex.EncodeToString(hb[:])
}
