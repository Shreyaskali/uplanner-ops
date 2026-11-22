package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/shreyaskali/uplanner-ops/internal/config"
	"github.com/shreyaskali/uplanner-ops/internal/core/learn"
	"github.com/shreyaskali/uplanner-ops/internal/data"
)

func loadContent(root string, out any) error {
	content, err := data.Load(root)
	if err != nil {
		return err
	}
	switch target := out.(type) {
	case **data.Content:
		*target = content
	}
	return nil
}

func runResources(root string, cfg config.Config, args []string, jsonOut bool) error {
	fs := flag.NewFlagSet("resources", flag.ContinueOnError)
	domain := fs.String("domain", "", "domain to filter by (repeatable, comma separated)")
	limit := fs.Int("limit", 8, "maximum resources")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	var content *data.Content
	if err := loadContent(root, &content); err != nil {
		return err
	}
	domains := splitList(*domain)
	completed := readProgress(cfg)
	items := learn.Select(content, domains, completed, *limit)
	if jsonOut {
		return json.NewEncoder(os.Stdout).Encode(items)
	}
	if len(items) == 0 {
		fmt.Println("No free resources matched. Try a domain such as ai, data, finance, cloud or lead.")
		return nil
	}
	fmt.Printf("Free resources (%d):\n\n", len(items))
	for _, item := range items {
		fmt.Printf("  %-32s %-10s %-12s %s\n", truncateRunes(item.Title, 32), item.Format, item.Duration, item.URL)
		fmt.Printf("    %s · %s\n", item.Path, item.Reason)
	}
	return nil
}

func runLearnPaths(root string, cfg config.Config, args []string, jsonOut bool) error {
	fs := flag.NewFlagSet("learn-paths", flag.ContinueOnError)
	domain := fs.String("domain", "", "domain to filter by (comma separated)")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	var content *data.Content
	if err := loadContent(root, &content); err != nil {
		return err
	}
	domains := splitList(*domain)
	paths := learn.MatchingPaths(content, domains, readProgress(cfg))
	if jsonOut {
		return json.NewEncoder(os.Stdout).Encode(paths)
	}
	if len(paths) == 0 {
		fmt.Println("No learning paths matched those domains.")
		return nil
	}
	fmt.Printf("Learning paths (%d):\n\n", len(paths))
	for _, p := range paths {
		fmt.Printf("  %-34s %-6s %3d resources  %s\n", p.Name, p.Tier, len(p.ResourceIDs), p.Tagline)
	}
	return nil
}

func runPlanLearning(root string, cfg config.Config, args []string, jsonOut bool) error {
	fs := flag.NewFlagSet("plan-learning", flag.ContinueOnError)
	gap := fs.String("gap", "", "describe the gap, e.g. 'SQL and dashboards'")
	domain := fs.String("domain", "", "domains directly (comma separated)")
	limit := fs.Int("limit", 6, "maximum resources")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	domains := splitList(*domain)
	if *gap != "" {
		domains = mergeUnique(domains, learn.GapDomains(*gap))
	}
	if len(domains) == 0 {
		return fmt.Errorf("supply --gap or --domain")
	}
	var content *data.Content
	if err := loadContent(root, &content); err != nil {
		return err
	}
	completed := readProgress(cfg)
	items := learn.Select(content, domains, completed, *limit)
	paths := learn.MatchingPaths(content, domains, completed)
	hours := 0.0
	for _, item := range items {
		hours += item.Hours
	}
	result := map[string]any{
		"gap": *gap, "domains": domains, "resources": items,
		"paths": paths, "estimatedHours": hours,
		"note": "Free resources only. Durations are approximate; 'Ongoing' items are not summed.",
	}
	if jsonOut {
		return json.NewEncoder(os.Stdout).Encode(result)
	}
	fmt.Printf("Learning plan for %q\n\n", *gap)
	fmt.Printf("Domains: %s\n", strings.Join(domains, ", "))
	for _, item := range items {
		fmt.Printf("  %-32s %-10s %-12s %s\n", truncateRunes(item.Title, 32), item.Format, item.Duration, item.URL)
	}
	if hours > 0 {
		fmt.Printf("\nApproximate tracked hours: %.1f\n", hours)
	}
	return nil
}

func runProgress(cfg config.Config, args []string, jsonOut bool) error {
	fs := flag.NewFlagSet("progress", flag.ContinueOnError)
	done := fs.String("done", "", "mark a resource id as done (gated)")
	confirm := fs.Bool("confirm", false, "confirm a gated write")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}

	state := readProgress(cfg)
	if *done == "" {
		ids := make([]string, 0, len(state))
		for id, value := range state {
			if value {
				ids = append(ids, id)
			}
		}
		sort.Strings(ids)
		if jsonOut {
			return json.NewEncoder(os.Stdout).Encode(map[string]any{"completed": ids})
		}
		fmt.Printf("Completed resources: %d\n", len(ids))
		for _, id := range ids {
			fmt.Println(" ", id)
		}
		return nil
	}

	if !*confirm {
		token := fmt.Sprintf("needs_confirmation: learning_write mark %s done - rerun with --confirm", *done)
		if jsonOut {
			return json.NewEncoder(os.Stdout).Encode(map[string]string{"needs_confirmation": token})
		}
		fmt.Println(token)
		return nil
	}
	state[*done] = true
	if err := writeProgress(cfg, state); err != nil {
		return err
	}
	if jsonOut {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"done": *done, "completed": len(state)})
	}
	fmt.Println("Marked done:", *done)
	return nil
}

// readProgress parses the Markdown table in data/learning.md.
func readProgress(cfg config.Config) map[string]bool {
	state := map[string]bool{}
	raw, err := os.ReadFile(cfg.Paths()["learning"])
	if err != nil {
		return state
	}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") || strings.Contains(line, "---") || strings.Contains(line, "resource") {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			continue
		}
		id := strings.TrimSpace(parts[1])
		value := strings.ToLower(strings.TrimSpace(parts[2]))
		if id != "" && (value == "done" || value == "complete" || value == "yes") {
			state[id] = true
		}
	}
	return state
}

func writeProgress(cfg config.Config, state map[string]bool) error {
	ids := make([]string, 0, len(state))
	for id, done := range state {
		if done {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	var b strings.Builder
	b.WriteString("# Learning progress\n\n| resource | state | note |\n|---|---|---|\n")
	for _, id := range ids {
		fmt.Fprintf(&b, "| %s | done |  |\n", id)
	}
	return os.WriteFile(cfg.Paths()["learning"], []byte(b.String()), 0o644)
}

func splitList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := []string{}
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func mergeUnique(a, b []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, item := range append(append([]string{}, a...), b...) {
		if !seen[item] {
			seen[item] = true
			out = append(out, item)
		}
	}
	sort.Strings(out)
	return out
}

func truncateRunes(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}
