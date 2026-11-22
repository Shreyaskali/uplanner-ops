package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/shreyaskali/uplanner-ops/internal/config"
	"github.com/shreyaskali/uplanner-ops/internal/core/learn"
	"github.com/shreyaskali/uplanner-ops/internal/data"
	"github.com/shreyaskali/uplanner-ops/internal/profile"
	"github.com/shreyaskali/uplanner-ops/internal/tracker"
)

func runGaps(cfg config.Config, args []string, jsonOut bool) error {
	fs := flag.NewFlagSet("gaps", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	dir := filepath.Join(cfg.DataDir, "data", "fit")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			entries = nil
		} else {
			return err
		}
	}
	type gapCount struct {
		Label   string   `json:"label"`
		Status  string   `json:"status"`
		Count   int      `json:"count"`
		Details []string `json:"details"`
	}
	counts := map[string]*gapCount{}
	re := regexp.MustCompile(`^- \[(met|partial|gap|note)\] ([^—]+) — (.*)$`)
	files := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		files++
		for _, line := range strings.Split(string(raw), "\n") {
			m := re.FindStringSubmatch(strings.TrimSpace(line))
			if len(m) < 4 || m[1] == "met" {
				continue
			}
			label := strings.TrimSpace(m[2])
			key := m[1] + "|" + label
			if counts[key] == nil {
				counts[key] = &gapCount{Label: label, Status: m[1], Details: []string{}}
			}
			counts[key].Count++
			counts[key].Details = append(counts[key].Details, strings.TrimSpace(m[3]))
		}
	}
	items := []*gapCount{}
	for _, item := range counts {
		items = append(items, item)
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Count != items[j].Count {
			return items[i].Count > items[j].Count
		}
		return items[i].Label < items[j].Label
	})
	result := map[string]any{"reports": files, "gaps": items}
	if jsonOut {
		return json.NewEncoder(os.Stdout).Encode(result)
	}
	fmt.Printf("Recurring gaps across %d saved fit reports:\n\n", files)
	if len(items) == 0 {
		fmt.Println("  No requirement gaps found. Save a fit report first with `fit --save --confirm`.")
	}
	for _, item := range items {
		fmt.Printf("  %2d×  [%-7s] %s\n", item.Count, item.Status, item.Label)
	}
	return nil
}

func runMarket(root string, args []string, jsonOut bool) error {
	fs := flag.NewFlagSet("market", flag.ContinueOnError)
	domain := fs.String("domain", "", "filter by domain (comma separated)")
	limit := fs.Int("limit", 8, "maximum roles")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	content, err := data.Load(root)
	if err != nil {
		return err
	}
	wanted := map[string]bool{}
	for _, d := range splitList(*domain) {
		wanted[d] = true
	}
	roles := []data.Role{}
	for _, role := range content.Roles {
		if len(wanted) > 0 {
			match := false
			for _, d := range role.Domains {
				if wanted[d] {
					match = true
				}
			}
			if !match {
				continue
			}
		}
		roles = append(roles, role)
	}
	if len(roles) > *limit {
		roles = roles[:*limit]
	}
	if jsonOut {
		return json.NewEncoder(os.Stdout).Encode(roles)
	}
	for _, role := range roles {
		fmt.Printf("%s — %s\n", role.Title, strings.Join(role.Domains, ", "))
		fmt.Println("  Reality:", truncateRunes(role.Reality, 180))
		if len(role.CommonGaps) > 0 {
			fmt.Println("  Common gaps:", truncateRunes(strings.Join(role.CommonGaps, "; "), 160))
		}
		fmt.Println()
	}
	return nil
}

func runPath(root string, cfg config.Config, args []string, jsonOut bool) error {
	fs := flag.NewFlagSet("path", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	content, err := data.Load(root)
	if err != nil {
		return err
	}
	p, err := profile.Load(cfg.Paths()["profile"])
	if err != nil {
		return err
	}
	domains := mergeUnique(learn.GapDomains(p.Field+" "+p.Skills+" "+p.Role), nil)
	roles := []data.Role{}
	for _, role := range content.Roles {
		for _, d := range role.Domains {
			if containsString(domains, d) {
				roles = append(roles, role)
				break
			}
		}
	}
	result := map[string]any{"domains": domains, "roles": roles,
		"note": "Role fit is a planning signal from your recorded field, skills and role. It is not an offer probability."}
	if jsonOut {
		return json.NewEncoder(os.Stdout).Encode(result)
	}
	fmt.Printf("Domains from your profile: %s\n\n", strings.Join(domains, ", "))
	for _, role := range roles {
		fmt.Printf("  %-32s %s\n", role.Title, strings.Join(role.Domains, ", "))
	}
	return nil
}

func containsString(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

type exam struct {
	Key    string   `json:"key"`
	Name   string   `json:"name"`
	Window string   `json:"window"`
	Phases []string `json:"phases"`
	Note   string   `json:"note"`
}

var exams = []exam{
	{"cat", "CAT 2026", "Typical: November; exact date unverified", []string{"T-12 weeks: sectional fundamentals", "T-8 weeks: timed sets and error log", "T-4 weeks: full mocks", "T-1 week: revision only"}, "Verify the official date. No score is invented here."},
	{"gre", "GRE", "Year-round; book your own slot", []string{"T-8 weeks: vocabulary and quant fundamentals", "T-4 weeks: timed practice", "T-2 weeks: full-length tests", "T-1 week: review only"}, "Verbal minimums can be binding. Check program policy before retaking."},
	{"gate", "GATE", "Typical: February; exact date unverified", []string{"T-16 weeks: core subjects", "T-8 weeks: previous papers", "T-4 weeks: mock tests", "T-1 week: revision"}, "Coaching or self-study; verify the official notification."},
	{"cfa-l1", "CFA Level I", "Typical windows vary by year", []string{"T-20 weeks: core readings", "T-10 weeks: practice questions", "T-4 weeks: mocks", "T-1 week: review"}, "Confirm the current exam window and fee with the provider."},
	{"frm-p1", "FRM Part I", "Typical windows vary by year", []string{"T-16 weeks: foundations", "T-8 weeks: practice", "T-3 weeks: mocks", "T-1 week: review"}, "Confirm the current window with GARP."},
}

func runCert(args []string, jsonOut bool) error {
	fs := flag.NewFlagSet("cert", flag.ContinueOnError)
	key := fs.String("exam", "", "exam key (cat, gre, gate, cfa-l1, frm-p1)")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	items := exams
	if *key != "" {
		items = []exam{}
		for _, e := range exams {
			if e.Key == *key {
				items = append(items, e)
			}
		}
		if len(items) == 0 {
			return fmt.Errorf("unknown exam %q", *key)
		}
	}
	if jsonOut {
		return json.NewEncoder(os.Stdout).Encode(items)
	}
	for _, e := range items {
		fmt.Printf("%s — %s\n", e.Name, e.Window)
		for _, phase := range e.Phases {
			fmt.Println("  ", phase)
		}
		fmt.Println("  Note:", e.Note)
		fmt.Println()
	}
	return nil
}

func runPlan(root string, cfg config.Config, args []string, jsonOut bool) error {
	fs := flag.NewFlagSet("plan", flag.ContinueOnError)
	days := fs.Int("days", 7, "planning horizon in days")
	limit := fs.Int("limit", 2, "learning resources per deadline")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	content, err := data.Load(root)
	if err != nil {
		return err
	}
	records, err := tracker.Read(cfg.Paths()["applications"])
	if err != nil {
		return err
	}
	today := todayIST()
	type block struct {
		ID          string                 `json:"id"`
		Institution string                 `json:"institution"`
		Date        string                 `json:"date"`
		Days        int                    `json:"days"`
		Status      string                 `json:"status"`
		Learn       []learn.Recommendation `json:"learn"`
	}
	blocks := []block{}
	for _, rec := range records {
		if rec.Status == "submitted" || rec.Deadline == "" {
			continue
		}
		until := daysUntil(today, rec.Deadline)
		if until < 0 || until > *days {
			continue
		}
		p := content.ProgramByID(rec.ID)
		name := rec.ID
		domains := []string{}
		if p != nil {
			name = p.Institution
			domains = learn.DomainsForGroup(p.Group)
		}
		resources := learn.Select(content, domains, readProgress(cfg), *limit)
		blocks = append(blocks, block{ID: rec.ID, Institution: name, Date: rec.Deadline, Days: until, Status: rec.Status, Learn: resources})
	}
	sort.SliceStable(blocks, func(i, j int) bool { return blocks[i].Date < blocks[j].Date })
	result := map[string]any{
		"today": today, "horizonDays": *days, "blocks": blocks,
		"note": "Capacity-aware: verify every date; learning items are free only.",
	}
	if jsonOut {
		return json.NewEncoder(os.Stdout).Encode(result)
	}
	fmt.Printf("Plan for %s (%d days):\n\n", today, *days)
	if len(blocks) == 0 {
		fmt.Println("  No application deadlines inside the horizon. Set one with `track --id ... --status ...`.")
	}
	for _, b := range blocks {
		fmt.Printf("  %s · %s · in %d days · %s\n", b.Date, b.Institution, b.Days, b.Status)
		for _, r := range b.Learn {
			fmt.Printf("    - %s (%s, %s)\n", r.Title, r.Format, r.Duration)
		}
	}
	return nil
}
