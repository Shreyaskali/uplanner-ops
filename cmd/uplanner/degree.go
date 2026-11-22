package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/shreyaskali/uplanner-ops/internal/config"
	"github.com/shreyaskali/uplanner-ops/internal/core/degree"
	"github.com/shreyaskali/uplanner-ops/internal/data"
	"github.com/shreyaskali/uplanner-ops/internal/tracker"
)

func todayIST() string {
	return time.Now().In(mustIST()).Format("2006-01-02")
}

func mustIST() *time.Location {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		return time.UTC
	}
	return loc
}

func runDegreePlan(root string, args []string, jsonOut bool) error {
	fs := flag.NewFlagSet("degree-plan", flag.ContinueOnError)
	country := fs.String("country", "", "filter by country")
	group := fs.String("group", "", "filter by degree area")
	limit := fs.Int("limit", 12, "maximum results")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	content, err := data.Load(root)
	if err != nil {
		return err
	}
	items := degree.Rank(content, *country, *group, todayIST(), *limit)
	if jsonOut {
		return json.NewEncoder(os.Stdout).Encode(items)
	}
	if len(items) == 0 {
		fmt.Println("No programs matched those filters.")
		return nil
	}
	fmt.Printf("Programs (%d):\n\n", len(items))
	for _, item := range items {
		fmt.Printf("  %-38s %-5s %3d  %s\n", truncateRunes(item.Institution, 38), item.Priority, item.Score, item.Evidence)
		fmt.Printf("    %s · %s · next %s\n", truncateRunes(item.Program, 60), item.Country, dash(item.NextRound))
	}
	return nil
}

func runCompare(root string, args []string, jsonOut bool) error {
	content, err := data.Load(root)
	if err != nil {
		return err
	}
	ids := []string{}
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			continue
		}
		ids = append(ids, a)
	}
	if len(ids) != 2 {
		return fmt.Errorf("usage: uplanner compare <program-id-a> <program-id-b>")
	}
	result, err := degree.Compare(content, ids[0], ids[1])
	if err != nil {
		return err
	}
	if jsonOut {
		return json.NewEncoder(os.Stdout).Encode(result)
	}
	for _, key := range []string{"a", "b"} {
		v := result[key].(map[string]any)
		fmt.Printf("%s — %s\n", v["institution"], v["program"])
		for _, field := range []string{"country", "group", "priority", "studyLength", "entryGate", "tuition", "funding", "salary", "evidence", "caveat", "verifiedAsOf"} {
			fmt.Printf("  %-14s %s\n", field, fmt.Sprint(v[field]))
		}
		fmt.Println()
	}
	fmt.Println(result["note"])
	return nil
}

func runCost(args []string, jsonOut bool) error {
	fs := flag.NewFlagSet("cost", flag.ContinueOnError)
	months := fs.Float64("months", 24, "months away from paid work")
	pay := fs.Float64("pay", 18, "alternative annual gross pay in ₹ lakh")
	tuition := fs.Float64("tuition", 0, "total tuition and living cost in ₹ lakh")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	forgone := *pay * *months / 12
	result := map[string]any{
		"months": *months, "annualPayLakh": *pay, "tuitionLakh": *tuition,
		"forgoneLakh": forgone, "totalLakh": forgone + *tuition,
		"note": "Planning scenario before tax, raises and scholarships. Indian CTC, MEP and foreign base salary are not interchangeable.",
	}
	if jsonOut {
		return json.NewEncoder(os.Stdout).Encode(result)
	}
	fmt.Printf("Gross pay forgone: ₹%.1fL\n", forgone)
	fmt.Printf("With study costs:  ₹%.1fL\n", forgone+*tuition)
	fmt.Println(result["note"])
	return nil
}

func runDeadlines(root string, cfg config.Config, args []string, jsonOut bool) error {
	fs := flag.NewFlagSet("deadlines", flag.ContinueOnError)
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
	items := []dueItem{}
	for _, rec := range records {
		if rec.Status == "submitted" || rec.Deadline == "" {
			continue
		}
		p := content.ProgramByID(rec.ID)
		name, degreeName, label := rec.ID, "", "tracker deadline"
		date := rec.Deadline
		if p != nil {
			name, degreeName = p.Institution, p.Program
			if next := degreeNextRound(p, today); next != "" {
				label = next
			}
		}
		days := daysUntil(today, date)
		items = append(items, dueItem{ID: rec.ID, Institution: name, Program: degreeName, Date: date, Label: label, Status: rec.Status, Days: days})
	}
	sortDue(items)
	if jsonOut {
		return json.NewEncoder(os.Stdout).Encode(items)
	}
	if len(items) == 0 {
		fmt.Println("No upcoming application deadlines in your tracker.")
		return nil
	}
	for _, item := range items {
		state := fmt.Sprintf("in %d days", item.Days)
		if item.Days < 0 {
			state = fmt.Sprintf("%d days past", -item.Days)
		}
		if item.Days == 0 {
			state = "today"
		}
		fmt.Printf("  %-11s %-32s %-36s %s\n", item.Date, truncateRunes(item.Institution, 32), truncateRunes(item.Program, 36), state)
	}
	return nil
}

func degreeNextRound(p *data.Program, today string) string {
	best := ""
	for _, m := range p.Milestones {
		if m.Type != "application" || len(m.Date) != 10 || m.Date < today {
			continue
		}
		if best == "" || m.Date < best {
			best = m.Date
		}
	}
	return best
}

func daysUntil(today, date string) int {
	a, err1 := time.Parse("2006-01-02", today)
	b, err2 := time.Parse("2006-01-02", date)
	if err1 != nil || err2 != nil {
		return 0
	}
	return int(b.Sub(a).Hours() / 24)
}

type dueItem struct {
	ID          string `json:"id"`
	Institution string `json:"institution"`
	Program     string `json:"program"`
	Date        string `json:"date"`
	Label       string `json:"label"`
	Status      string `json:"status"`
	Days        int    `json:"days"`
}

func sortDue(items []dueItem) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Date != items[j].Date {
			if items[i].Date == "" {
				return false
			}
			if items[j].Date == "" {
				return true
			}
			return items[i].Date < items[j].Date
		}
		return items[i].ID < items[j].ID
	})
}

func runTrack(cfg config.Config, args []string, jsonOut bool) error {
	fs := flag.NewFlagSet("track", flag.ContinueOnError)
	id := fs.String("id", "", "program id to update")
	status := fs.String("status", "", "new status")
	confirm := fs.Bool("confirm", false, "confirm a gated write")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	path := cfg.Paths()["applications"]
	records, err := tracker.Read(path)
	if err != nil {
		return err
	}
	if *id == "" {
		if jsonOut {
			return json.NewEncoder(os.Stdout).Encode(records)
		}
		for _, r := range records {
			fmt.Printf("  %-42s %-12s %-11s %s\n", r.ID, r.Status, dash(r.Deadline), r.Documents)
		}
		return nil
	}
	if *status == "" {
		return fmt.Errorf("--status is required with --id")
	}
	if !tracker.IsValidStatus(*status) {
		return fmt.Errorf("unsupported status %q", *status)
	}
	if !*confirm {
		token := fmt.Sprintf("needs_confirmation: track_write set %s to %s - rerun with --confirm", *id, *status)
		if jsonOut {
			return json.NewEncoder(os.Stdout).Encode(map[string]string{"needs_confirmation": token})
		}
		fmt.Println(token)
		return nil
	}
	found := false
	for i := range records {
		if records[i].ID == *id {
			records[i].Status = *status
			found = true
		}
	}
	if !found {
		records = append(records, tracker.Record{ID: *id, Status: *status})
	}
	if err := tracker.Write(path, records); err != nil {
		return err
	}
	if jsonOut {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"id": *id, "status": *status})
	}
	fmt.Printf("Updated %s → %s\n", *id, *status)
	return nil
}

func runExport(cfg config.Config, args []string, jsonOut bool) error {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	out := fs.String("out", "", "write markdown to a file instead of stdout")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	records, err := tracker.Read(cfg.Paths()["applications"])
	if err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("# University Planner — Application Export\n\n")
	fmt.Fprintf(&b, "Generated: %s\n\n", todayIST())
	b.WriteString("## Applications\n\n")
	if len(records) == 0 {
		b.WriteString("(No programs tracked yet.)\n")
	} else {
		b.WriteString("| id | status | deadline | internal target | documents |\n|---|---|---|---|---|\n")
		for _, r := range records {
			fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n", r.ID, r.Status, dash(r.Deadline), dash(r.Target), dash(r.Documents))
		}
	}
	b.WriteString("\nVerify every 2027 date on the official programme page before relying on it.\n")
	if *out != "" {
		if err := os.WriteFile(*out, []byte(b.String()), 0o644); err != nil {
			return err
		}
		if jsonOut {
			return json.NewEncoder(os.Stdout).Encode(map[string]string{"written": *out})
		}
		fmt.Println("Wrote", *out)
		return nil
	}
	if jsonOut {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"markdown": b.String()})
	}
	fmt.Print(b.String())
	return nil
}

func dash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "—"
	}
	return value
}
