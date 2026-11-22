// Package learn maps skill gaps onto free learning resources and paths.
// It is deterministic: selection is a function of the generated content,
// the requested domains, a time budget and recorded progress.
package learn

import (
	"fmt"
	"sort"
	"strings"

	"github.com/shreyaskali/uplanner-ops/internal/data"
)

// domainAliases maps fit dimension groups to Intersections domains.
var domainAliases = map[string][]string{
	"AI / analytics":            {"ai", "data"},
	"Finance / quant":           {"finance", "wm"},
	"Technology / product":      {"product", "cloud", "design"},
	"Supply chain / operations": {"sc", "program"},
	"Economics":                 {"finance", "data"},
	"MBA / management":          {"lead", "entre", "program"},
}

// GapDomains converts a free-text gap into candidate domains.
func GapDomains(gap string) []string {
	lower := strings.ToLower(gap)
	keywords := map[string][]string{
		"ai":      {"ai", "machine learning", "llm", "model", "deep learning", "mlops"},
		"data":    {"sql", "data", "analytics", "dashboard", "statistics", "python", "etl"},
		"finance": {"finance", "valuation", "investment", "markets", "risk", "accounting"},
		"wm":      {"wealth", "portfolio", "advisory", "client"},
		"cloud":   {"cloud", "aws", "azure", "gcp", "infrastructure", "devops", "kubernetes"},
		"product": {"product", "roadmap", "user research", "growth", "ux"},
		"sc":      {"supply chain", "logistics", "operations", "procurement", "inventory"},
		"lead":    {"leadership", "management", "team", "stakeholder", "communication"},
		"entre":   {"founder", "startup", "fundraising", "venture", "entrepreneur"},
		"program": {"delivery", "project", "programme", "program"},
		"design":  {"design", "ux", "interface"},
		"sec":     {"security", "cyber", "threat"},
		"growth":  {"growth", "marketing", "sales"},
	}
	matched := []string{}
	for domain, terms := range keywords {
		for _, term := range terms {
			if strings.Contains(lower, term) {
				matched = append(matched, domain)
				break
			}
		}
	}
	sort.Strings(matched)
	return matched
}

// DomainsForGroup returns the Intersections domains for a fit dimension group.
func DomainsForGroup(group string) []string {
	if domains, ok := domainAliases[group]; ok {
		return append([]string(nil), domains...)
	}
	return nil
}

// Recommendation is one selected free resource with its reason.
type Recommendation struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Source   string  `json:"source"`
	Format   string  `json:"format"`
	Duration string  `json:"duration"`
	URL      string  `json:"url"`
	Note     string  `json:"note"`
	Path     string  `json:"path"`
	Domain   string  `json:"domain"`
	Reason   string  `json:"reason"`
	Score    int     `json:"score"`
	Hours    float64 `json:"hours"`
}

type candidate struct {
	resource data.Resource
	score    int
	domain   string
}

// Select picks a small, coherent set of free resources for the requested
// domains. Completed resources are excluded. It never returns paid content.
func Select(content *data.Content, domains []string, completed map[string]bool, limit int) []Recommendation {
	wanted := map[string]bool{}
	for _, d := range domains {
		wanted[d] = true
	}

	candidates := []candidate{}
	for _, r := range content.Resources {
		if !r.Free || completed[r.ID] {
			continue
		}
		score := 0
		if wanted[r.Domain] {
			score += 20
		}
		if r.Path != "" {
			for _, p := range content.Paths {
				if p.Name != r.Path {
					continue
				}
				// Path bonuses apply only when the path itself matches a
				// requested domain. A free tier alone is not a match.
				pathMatch := false
				for _, pd := range p.Domains {
					if wanted[pd] {
						pathMatch = true
						break
					}
				}
				if pathMatch {
					score += 12
					if p.Tier == "free" {
						score += 6
					}
				}
				break
			}
		}
		if score == 0 {
			continue
		}
		candidates = append(candidates, candidate{resource: r, score: score, domain: r.Domain})
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}
		return candidates[i].resource.ID < candidates[j].resource.ID
	})

	if limit <= 0 {
		limit = 8
	}
	if limit > 20 {
		limit = 20
	}

	out := []Recommendation{}
	seenPath := map[string]int{}
	for _, c := range candidates {
		if len(out) >= limit {
			break
		}
		// Prefer variety within a path: at most four resources per path.
		if seenPath[c.resource.Path] >= 4 {
			continue
		}
		seenPath[c.resource.Path]++
		reason := "Matches your requested domain"
		if wanted[c.domain] {
			reason += " (" + c.domain + ")"
		}
		out = append(out, Recommendation{
			ID: c.resource.ID, Title: c.resource.Title, Source: c.resource.Source,
			Format: c.resource.Format, Duration: c.resource.Duration, URL: c.resource.URL,
			Note: c.resource.Note, Path: c.resource.Path, Domain: c.resource.Domain,
			Reason: reason, Score: c.score, Hours: parseHours(c.resource.Duration),
		})
	}
	return out
}

// MatchingPaths returns paths that touch any of the requested domains.
func MatchingPaths(content *data.Content, domains []string, completed map[string]bool) []data.LearningPath {
	wanted := map[string]bool{}
	for _, d := range domains {
		wanted[d] = true
	}
	out := []data.LearningPath{}
	for _, p := range content.Paths {
		// Free-first: paid tiers are excluded from plans entirely.
		if p.Tier != "free" {
			continue
		}
		match := false
		for _, d := range p.Domains {
			if wanted[d] {
				match = true
			}
		}
		if !match {
			continue
		}
		remaining := 0
		for _, id := range p.ResourceIDs {
			if !completed[id] {
				remaining++
			}
		}
		if remaining > 0 {
			out = append(out, p)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Tier != out[j].Tier {
			return out[i].Tier == "free"
		}
		return out[i].Name < out[j].Name
	})
	return out
}

// parseHours extracts an approximate decimal hour count from a duration
// label. Labels without a number return zero and are not summed.
func parseHours(label string) float64 {
	lower := strings.ToLower(label)
	if strings.Contains(lower, "ongoing") {
		return 0
	}
	fields := strings.Fields(lower)
	for i, f := range fields {
		f = strings.Trim(f, "~+-")
		value := 0.0
		if _, err := fmt.Sscan(f, &value); err == nil {
			if i+1 < len(fields) && strings.HasPrefix(fields[i+1], "h") {
				return value
			}
			if strings.HasSuffix(f, "h") {
				return value
			}
			if strings.HasSuffix(f, "min") {
				return value / 60
			}
			if i+1 < len(fields) && strings.HasPrefix(fields[i+1], "min") {
				return value / 60
			}
		}
	}
	return 0
}
