// Package doctor validates the install without touching user data.
package doctor

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/shreyaskali/uplanner-ops/internal/config"
	"github.com/shreyaskali/uplanner-ops/internal/data"
)

// Check is one validation result.
type Check struct {
	Name   string `json:"name"`
	Status string `json:"status"` // ok | warn | fail
	Detail string `json:"detail"`
}

// Report is the full doctor result.
type Report struct {
	Checks []Check `json:"checks"`
	OK     bool    `json:"ok"`
}

func add(c *Check, status, format string, args ...any) {
	c.Status = status
	c.Detail = fmt.Sprintf(format, args...)
}

// Run checks the system layer, generated content and the user data root.
func Run(root string, cfg config.Config) Report {
	report := Report{Checks: []Check{}}

	check := Check{Name: "go"}
	add(&check, "ok", "%s", runtime.Version())
	report.Checks = append(report.Checks, check)

	// Contract files.
	for _, name := range []string{"AGENTS.md", "SKILL.md", "DATA_CONTRACT.md"} {
		check = Check{Name: "contract:" + name}
		if _, err := os.Stat(filepath.Join(root, name)); err != nil {
			add(&check, "fail", "missing %s", name)
		} else {
			add(&check, "ok", "present")
		}
		report.Checks = append(report.Checks, check)
	}

	// Modes.
	check = Check{Name: "modes"}
	entries, err := os.ReadDir(filepath.Join(root, "modes"))
	if err != nil || len(entries) == 0 {
		add(&check, "fail", "no modes/ directory or no mode files")
	} else {
		add(&check, "ok", "%d prompt files", len(entries))
	}
	report.Checks = append(report.Checks, check)

	// Generated content.
	check = Check{Name: "content"}
	content, err := data.Load(root)
	if err != nil {
		add(&check, "fail", "%v", err)
	} else {
		add(&check, "ok", "%d programs, %d reports, %d free resources, %d paths, %d roles",
			len(content.Programs), len(content.Reports), len(content.Resources), len(content.Paths), len(content.Roles))
	}
	report.Checks = append(report.Checks, check)

	// Data root.
	check = Check{Name: "data-root"}
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		add(&check, "fail", "cannot create %s: %v", cfg.DataDir, err)
	} else {
		add(&check, "ok", "%s", cfg.DataDir)
	}
	report.Checks = append(report.Checks, check)

	// The generated content must never live inside the user data root.
	check = Check{Name: "boundary"}
	absRoot, err1 := filepath.Abs(root)
	absData, err2 := filepath.Abs(cfg.DataDir)
	if err1 == nil && err2 == nil && strings.HasPrefix(absData, absRoot+string(filepath.Separator)+"content") {
		add(&check, "fail", "data root resolves inside the system content directory")
	} else {
		add(&check, "ok", "system and user layers are separate")
	}
	report.Checks = append(report.Checks, check)

	// Scaffold user files without overwriting.
	check = Check{Name: "scaffold"}
	if err := cfg.Ensure(); err != nil {
		add(&check, "fail", "%v", err)
	} else {
		add(&check, "ok", "user-layer directories and scaffolds ready")
	}
	report.Checks = append(report.Checks, check)

	// Optional LLM key: report presence only, never the value.
	check = Check{Name: "llm-key"}
	if os.Getenv("UPLANNER_LLM_KEY") != "" {
		add(&check, "ok", "configured (value hidden)")
	} else {
		add(&check, "warn", "not set; offline modes still work")
	}
	report.Checks = append(report.Checks, check)

	report.OK = true
	for _, c := range report.Checks {
		if c.Status == "fail" {
			report.OK = false
		}
	}
	return report
}

// Text renders a human-readable report.
func (r Report) Text() string {
	var b strings.Builder
	for _, c := range r.Checks {
		mark := "OK  "
		if c.Status == "warn" {
			mark = "WARN"
		}
		if c.Status == "fail" {
			mark = "FAIL"
		}
		fmt.Fprintf(&b, "%s  %-20s %s\n", mark, c.Name, c.Detail)
	}
	return b.String()
}
