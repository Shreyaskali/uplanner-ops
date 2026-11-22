// Package degree ranks programs from generated content using evidence and
// cost signals. It never invents a deadline or a salary figure.
package degree

import (
	"sort"
	"strings"

	"github.com/shreyaskali/uplanner-ops/internal/data"
)

// Candidate is one ranked program with the reasons for its position.
type Candidate struct {
	ID           string   `json:"id"`
	Institution  string   `json:"institution"`
	Program      string   `json:"program"`
	Country      string   `json:"country"`
	Group        string   `json:"group"`
	Priority     string   `json:"priority"`
	Evidence     string   `json:"evidence"`
	Salary       string   `json:"salary"`
	Outcome      string   `json:"outcome"`
	NextRound    string   `json:"nextRound"`
	VerifiedAsOf string   `json:"verifiedAsOf"`
	SourceURL    string   `json:"sourceUrl"`
	Score        int      `json:"score"`
	Reasons      []string `json:"reasons"`
}

func evidenceRank(kind string) int {
	switch kind {
	case "placement_report":
		return 4
	case "recruiter_list":
		return 3
	case "alumni":
		return 2
	case "no_report":
		return 1
	}
	return 0
}

func priorityRank(p string) int {
	switch p {
	case "P1":
		return 3
	case "P2":
		return 2
	case "P3":
		return 1
	}
	return 0
}

// nextRound returns the soonest application round with its label, or "".
func nextRound(p data.Program, today string) string {
	bestDate, bestLabel := "", ""
	for _, m := range p.Milestones {
		if m.Type != "application" || len(m.Date) != 10 {
			continue
		}
		if m.Date < today {
			continue
		}
		if bestDate == "" || m.Date < bestDate {
			bestDate, bestLabel = m.Date, m.Label
		}
	}
	if bestDate == "" {
		for _, m := range p.Milestones {
			if m.Type == "application" && len(m.Date) == 10 && (bestDate == "" || m.Date > bestDate) {
				bestDate, bestLabel = m.Date, m.Label
			}
		}
	}
	if bestDate == "" {
		return ""
	}
	return bestDate + " · " + bestLabel
}

// Rank scores programs for a shortlist. Filters are optional; an empty string
// means "all". The score is transparent and every reason is named.
func Rank(content *data.Content, country, group, today string, limit int) []Candidate {
	candidates := []Candidate{}
	for _, p := range content.Programs {
		if country != "" && !strings.EqualFold(p.Country, country) {
			continue
		}
		if group != "" && !strings.EqualFold(p.Group, group) {
			continue
		}
		report := content.ReportByID(p.ReportID)
		kind := ""
		if report != nil {
			kind = report.EvidenceKind
		}
		score := 40 + evidenceRank(kind)*8 + priorityRank(p.Priority)*4
		reasons := []string{}
		if report != nil {
			reasons = append(reasons, "Reviewed source: "+kind)
			if report.Salary != "" {
				score += 6
				reasons = append(reasons, "Salary evidence recorded")
			}
		} else {
			reasons = append(reasons, "Not reviewed yet")
		}
		if p.Priority != "" {
			reasons = append(reasons, "Research tier "+p.Priority)
		}
		if next := nextRound(p, today); next != "" {
			reasons = append(reasons, "Next recorded round "+next)
		}
		candidates = append(candidates, Candidate{
			ID: p.ID, Institution: p.Institution, Program: p.Program,
			Country: p.Country, Group: p.Group, Priority: p.Priority,
			Evidence: evidenceLabel(kind), Salary: salaryOf(report),
			Outcome: p.Outcome, NextRound: nextRound(p, today),
			VerifiedAsOf: p.VerifiedAsOf, SourceURL: p.SourceURL,
			Score: score, Reasons: reasons,
		})
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Score != candidates[j].Score {
			return candidates[i].Score > candidates[j].Score
		}
		return candidates[i].Institution < candidates[j].Institution
	})
	if limit > 0 && len(candidates) > limit {
		candidates = candidates[:limit]
	}
	return candidates
}

func evidenceLabel(kind string) string {
	switch kind {
	case "placement_report":
		return "Published outcomes"
	case "recruiter_list":
		return "Recruiter list"
	case "alumni":
		return "Alumni evidence"
	case "no_report":
		return "No report found"
	}
	return "Not reviewed yet"
}

func salaryOf(report *data.Report) string {
	if report == nil {
		return ""
	}
	return report.Salary
}

// Compare builds a side-by-side view of two programs using only generated
// facts. Missing facts are labelled, never inferred.
func Compare(content *data.Content, idA, idB string) (map[string]any, error) {
	a := content.ProgramByID(idA)
	b := content.ProgramByID(idB)
	if a == nil {
		return nil, errNotFound(idA)
	}
	if b == nil {
		return nil, errNotFound(idB)
	}
	return map[string]any{
		"a":    programView(content, a),
		"b":    programView(content, b),
		"note": "CTC, MEP, base and guaranteed compensation are not interchangeable. Verify the 2027 cycle on the official page.",
	}, nil
}

func programView(content *data.Content, p *data.Program) map[string]any {
	report := content.ReportByID(p.ReportID)
	rounds := []string{}
	for _, m := range p.Milestones {
		if m.Type == "application" && len(m.Date) == 10 {
			rounds = append(rounds, m.Date+" · "+m.Label)
		}
	}
	view := map[string]any{
		"id": p.ID, "institution": p.Institution, "program": p.Program,
		"country": p.Country, "group": p.Group, "priority": p.Priority,
		"studyLength": p.StudyLength, "entryGate": p.EntryGate,
		"caveat": p.Caveat, "tuition": p.PublishedTuition,
		"funding": p.FundingNote, "verifiedAsOf": p.VerifiedAsOf,
		"sourceUrl": p.SourceURL, "deadlineUrl": p.DeadlineURL,
		"rounds": rounds, "outcome": p.Outcome,
		"evidence": evidenceLabel(func() string {
			if report != nil {
				return report.EvidenceKind
			}
			return ""
		}()),
		"salary": salaryOf(report), "cohort": "", "roles": []string{},
	}
	if report != nil {
		view["cohort"] = report.CohortScope
		view["roles"] = report.Roles
		view["sourceUrl"] = report.SourceURL
	}
	if view["tuition"] == "" {
		view["tuition"] = "Not verified"
	}
	if view["studyLength"] == "" {
		view["studyLength"] = "Verify on program page"
	}
	return view
}

type notFoundError string

func (e notFoundError) Error() string { return "program not found: " + string(e) }

func errNotFound(id string) error { return notFoundError(id) }
