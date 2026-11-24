// Package fit is the deterministic job-description fit engine.
// It is a faithful Go port of compass-core.js. No model, no randomness:
// the same inputs always produce the same analysis.
package fit

import (
	"regexp"
	"sort"
	"strings"

	"github.com/shreyaskali/uplanner-ops/internal/profile"
)

type dimension struct {
	ID       string
	Label    string
	Group    string
	Strength int
	Roles    []string
	Evidence string
	Keywords []string
}

var dimensions = []dimension{
	{"analytics", "Data & analytics", "AI / analytics", 72,
		[]string{"Data analyst", "Business analyst", "Analytics manager", "Product analyst"},
		"Your degree and analytical coursework support metric-driven, reporting and dashboard work.",
		[]string{"analytics", "analytical", "data", "data analysis", "data-driven", "sql", "python", "pandas", "excel", "dashboard", "dashboards", "metrics", "kpi", "kpis", "reporting", "forecasting", "statistics", "statistical", "segmentation", "tableau", "power bi", "bigquery", "etl", "data pipeline"}},
	{"ai", "AI & machine learning", "AI / analytics", 70,
		[]string{"AI product analyst", "Applied AI associate", "Data scientist", "AI governance analyst"},
		"Quantitative preparation and any modeling coursework are relevant foundations for AI work.",
		[]string{"artificial intelligence", "machine learning", "deep learning", "generative ai", "genai", "llm", "llms", "nlp", "computer vision", "model deployment", "mlops", "prompt engineering", "ai governance", "ai observability", "model monitoring", "ai"}},
	{"finance", "Finance & markets", "Finance / quant", 70,
		[]string{"Finance analyst", "Strategy and finance associate", "Risk analyst", "Fintech strategy associate"},
		"Quantitative background and analytical rigor suit finance and markets roles.",
		[]string{"finance", "financial", "financial modeling", "valuation", "dcf", "investment banking", "m&a", "equity research", "portfolio management", "asset management", "wealth management", "private equity", "venture capital", "trading", "risk management", "credit risk", "treasury", "fp&a", "accounting", "audit", "fintech", "payments", "lending", "banking", "capital markets"}},
	{"consulting", "Strategy & consulting", "MBA / management", 70,
		[]string{"Strategy analyst", "Consultant", "Business transformation associate", "Corporate strategy associate"},
		"Structured problem-solving and communication suit strategy and consulting work.",
		[]string{"consulting", "consultant", "strategy", "strategic", "client engagement", "client-facing", "stakeholder management", "due diligence", "market entry", "operating model", "business case", "problem solving", "recommendations", "transformation", "change management"}},
	{"technology", "Cloud & technology", "Technology / product", 74,
		[]string{"Cloud strategy analyst", "Platform strategy associate", "Technology consultant", "Solutions analyst"},
		"Technical coursework and any engineering background suit product, platform and technology roles.",
		[]string{"software", "software engineering", "cloud", "aws", "azure", "gcp", "google cloud", "kubernetes", "docker", "devops", "platform", "architecture", "api", "apis", "microservices", "observability", "monitoring", "infrastructure", "saas", "security", "data engineering", "site reliability", "terraform", "java", "javascript", "golang"}},
	{"product", "Product & growth", "Technology / product", 70,
		[]string{"Product analyst", "Associate product manager", "Product strategy associate", "Growth analyst"},
		"Product thinking and any build or launch experience suit product and growth roles.",
		[]string{"product manager", "product management", "product owner", "roadmap", "product strategy", "user research", "customer discovery", "go-to-market", "gtm", "feature", "features", "agile", "scrum", "backlog", "ux", "product analytics", "a/b testing", "launch"}},
	{"operations", "Operations & supply chain", "Supply chain / operations", 70,
		[]string{"Operations analyst", "Supply chain analyst", "Process improvement analyst", "Planning associate"},
		"Engineering and systems thinking suit operations, logistics and supply chain roles.",
		[]string{"operations", "operational", "supply chain", "logistics", "procurement", "inventory", "manufacturing", "production", "process improvement", "lean", "six sigma", "quality", "warehouse", "planning", "s&op", "capacity planning"}},
	{"economics", "Economics & policy", "Economics", 68,
		[]string{"Economic analyst", "Policy analyst", "Pricing analyst", "Research associate"},
		"Quantitative training supports economics and policy analysis.",
		[]string{"economics", "economic", "microeconomics", "macroeconomics", "econometrics", "pricing", "market design", "public policy", "policy analysis", "regulation", "cost-benefit"}},
	{"management", "Management & leadership", "MBA / management", 66,
		[]string{"Management associate", "Strategy manager", "Operations manager", "General management associate"},
		"Any leadership or ownership experience is relevant to general management and people-management roles.",
		[]string{"people management", "team management", "line management", "direct reports", "p&l", "profit and loss", "budget ownership", "general management", "leadership", "lead a team", "manage a team", "hiring", "performance management"}},
	{"entrepreneurship", "Founding & venture", "MBA / management", 66,
		[]string{"Founder", "Venture analyst", "New ventures associate", "Chief of staff"},
		"Any 0→1 build or ownership experience suits founding, venture and new-venture roles.",
		[]string{"founder", "founding", "startup", "start-up", "zero to one", "venture", "fundraising", "bootstrapped", "entrepreneur", "ownership", "build from scratch"}},
}

var fieldSignals = []struct {
	group string
	terms []string
}{
	{"AI / analytics", []string{"analytics", "data", "statistics", "statistical", "mathematics", "computer science", "computer", "computing", "actuarial", "quantitative", "econometrics", "ai", "ml"}},
	{"Finance / quant", []string{"finance", "financial", "accounting", "economics", "business", "commerce", "banking", "quant"}},
	{"Technology / product", []string{"computer", "computing", "software", "information technology", "electronics", "electrical", "engineering", "technology", "product", "design"}},
	{"Supply chain / operations", []string{"operations", "operations research", "supply chain", "logistics", "industrial", "production", "manufacturing", "mechanical", "civil", "process"}},
	{"Economics", []string{"economics", "econometric", "policy"}},
	{"MBA / management", []string{"mba", "management", "business", "commerce", "bba", "administration"}},
}

// Dimension is one matched area with its evidence fit.
type Dimension struct {
	ID       string   `json:"id"`
	Label    string   `json:"label"`
	Group    string   `json:"group"`
	Fit      int      `json:"fit"`
	Demand   int      `json:"demand"`
	Evidence string   `json:"evidence"`
	Roles    []string `json:"roles"`
	Terms    []string `json:"terms"`
}

// Check is one explicit requirement with a status and evidence-based detail.
type Check struct {
	Label  string `json:"label"`
	Status string `json:"status"` // met | partial | gap | note
	Detail string `json:"detail"`
	Weight int    `json:"weight"`
}

// Gap is a user-facing requirement gap.
type Gap struct {
	Label  string `json:"label"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

// Analysis is the full deterministic result.
type Analysis struct {
	OK              bool        `json:"ok"`
	Title           string      `json:"title"`
	Score           int         `json:"score"`
	Band            string      `json:"band"`
	Tone            string      `json:"tone"`
	Headline        string      `json:"headline"`
	Summary         string      `json:"summary"`
	Confidence      string      `json:"confidence"`
	Words           int         `json:"words"`
	SignalCount     int         `json:"signalCount"`
	MatchedKeywords []string    `json:"matchedKeywords"`
	Dimensions      []Dimension `json:"dimensions"`
	Strengths       []Dimension `json:"strengths"`
	Requirements    []Check     `json:"requirements"`
	Gaps            []Gap       `json:"gaps"`
	NextMoves       []string    `json:"nextMoves"`
	Caveat          string      `json:"caveat"`
}

func norm(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "\u2013", "-")
	s = strings.ReplaceAll(s, "\u2014", "-")
	return strings.Join(strings.Fields(s), " ")
}

// countTerm mirrors compass-core.js: word-boundary matching for plain
// alphanumeric terms, substring counting for anything containing punctuation.
func countTerm(text, term string) int {
	needle := norm(term)
	if needle == "" {
		return 0
	}
	for _, r := range needle {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			return strings.Count(norm(text), needle)
		}
	}
	re := regexp.MustCompile(`\b` + regexp.QuoteMeta(needle) + `\b`)
	return len(re.FindAllString(norm(text), -1))
}

// fieldGroups maps the user's field and skills onto dimension groups.
func fieldGroups(p profile.Profile) map[string]bool {
	groups := map[string]bool{}
	haystack := norm(p.Field + " " + p.Skills)
	for _, signal := range fieldSignals {
		matched := false
		for _, term := range signal.terms {
			// JS uses substring matching here and stops at the FIRST group.
			if strings.Contains(haystack, term) {
				groups[signal.group] = true
				matched = true
				break
			}
		}
		if matched {
			break
		}
	}
	return groups
}

func reWord(term string) *regexp.Regexp {
	return regexp.MustCompile(`\b` + regexp.QuoteMeta(term) + `\b`)
}

func hasField(p profile.Profile, terms ...string) bool {
	haystack := norm(p.Field + " " + p.Skills + " " + p.Role)
	for _, term := range terms {
		if reWord(term).MatchString(haystack) {
			return true
		}
	}
	return false
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

type demandScore struct {
	raw       float64
	intensity float64
	fit       int
}

// Analyse scores a pasted job description against the structured profile.
func Analyse(text string, p profile.Profile) Analysis {
	source := strings.TrimSpace(text)
	words := len(strings.Fields(norm(source)))
	if words < 12 {
		return Analysis{OK: false, Summary: "Paste at least a few lines of the job description — role, responsibilities and requirements."}
	}

	lower := norm(source)
	head := lower
	if len(head) > 220 {
		head = head[:220]
	}

	demand := map[string]*demandScore{}
	matchedSet := map[string]bool{}
	matchedOrder := []string{}
	maxRaw := 0.0
	for _, d := range dimensions {
		raw := 0.0
		for _, keyword := range d.Keywords {
			n := countTerm(source, keyword)
			if n == 0 {
				continue
			}
			// JS compass-core.js: multi-word or punctuated keywords carry 1.4,
			// a keyword in the first 220 characters gets a 1.6 head bonus.
			weight := 1.0
			if strings.Contains(keyword, " ") || strings.ContainsAny(keyword, "&/-+") {
				weight = 1.4
			}
			raw += float64(min(n, 3)) * weight
			// JS uses containsTerm (word boundary), not substring: "platform"
			// inside "platforms" does not earn the head bonus.
			if countTerm(head, keyword) > 0 {
				raw += 1.6
			}
			if !matchedSet[keyword] {
				matchedSet[keyword] = true
				matchedOrder = append(matchedOrder, keyword)
			}
		}
		demand[d.ID] = &demandScore{raw: raw}
		if raw > maxRaw {
			maxRaw = raw
		}
	}

	groups := fieldGroups(p)
	matched := matchedOrder

	if maxRaw > 0 {
		for _, d := range dimensions {
			ds := demand[d.ID]
			ds.intensity = ds.raw / maxRaw
			// The displayed fit is strength-only in JS; the field-alignment
			// bonus affects strength sorting and the score base, not this field.
			ds.fit = clamp(floatToInt(float64(d.Strength)*(0.6+0.4*ds.intensity)), 5, 97)
		}
	} else {
		for _, d := range dimensions {
			demand[d.ID].fit = d.Strength
		}
	}

	checks := requirementChecks(lower, p)
	adjustment := 0
	for _, c := range checks {
		adjustment += c.Weight
	}
	adjustment = clamp(adjustment, -36, 5)

	base := 50.0
	if maxRaw > 0 {
		total := 0.0
		weighted := 0.0
		for _, d := range dimensions {
			ds := demand[d.ID]
			total += ds.raw
			align := 0
			if groups[d.Group] {
				align = 8
			}
			weighted += ds.raw * float64(d.Strength+align)
		}
		if total > 0 {
			base = weighted / total
		}
	}
	score := clamp(floatToInt(base)+adjustment, 5, 97)

	band, tone := bandFor(score)
	confidence := "Low"
	if words >= 170 && len(matched) >= 6 {
		confidence = "High"
	} else if words >= 70 && len(matched) >= 3 {
		confidence = "Medium"
	}

	ranked := make([]dimension, 0, len(dimensions))
	ranked = append(ranked, dimensions...)
	sort.SliceStable(ranked, func(i, j int) bool {
		a, b := demand[ranked[i].ID], demand[ranked[j].ID]
		if a.raw != b.raw {
			return a.raw > b.raw
		}
		if a.fit != b.fit {
			return a.fit > b.fit
		}
		return ranked[i].Label < ranked[j].Label
	})

	// Strengths use the field-aligned score, matching JS's (fit + align) * intensity.
	aligned := make([]dimension, len(ranked))
	copy(aligned, ranked)
	sort.SliceStable(aligned, func(i, j int) bool {
		ai, aj := aligned[i], aligned[j]
		bi, bj := demand[ai.ID], demand[aj.ID]
		vi := float64(bi.fit+alignOf(ai, groups)) * bi.intensity
		vj := float64(bj.fit+alignOf(aj, groups)) * bj.intensity
		if vi != vj {
			return vi > vj
		}
		if bi.fit != bj.fit {
			return bi.fit > bj.fit
		}
		return ai.Label < aj.Label
	})

	dims := []Dimension{}
	strengths := []Dimension{}
	for _, d := range ranked {
		if demand[d.ID].raw == 0 {
			continue
		}
		terms := []string{}
		for _, k := range d.Keywords {
			if countTerm(source, k) > 0 {
				terms = append(terms, k)
			}
		}
		item := Dimension{
			ID: d.ID, Label: d.Label, Group: d.Group, Fit: demand[d.ID].fit,
			Demand:   floatToInt(demand[d.ID].intensity * 100),
			Evidence: evidenceFor(d, p, groups), Roles: d.Roles,
			Terms: first(terms, 6),
		}
		dims = append(dims, item)
	}
	dims = first(dims, 3)
	for _, d := range aligned {
		if demand[d.ID].raw == 0 {
			continue
		}
		strengths = append(strengths, Dimension{
			ID: d.ID, Label: d.Label, Group: d.Group, Fit: demand[d.ID].fit,
			Evidence: evidenceFor(d, p, groups), Roles: d.Roles,
		})
		if len(strengths) == 3 {
			break
		}
	}

	gaps := []Gap{}
	for _, c := range checks {
		if c.Status == "gap" || c.Status == "partial" {
			gaps = append(gaps, Gap{Label: c.Label, Status: c.Status, Detail: c.Detail})
		}
	}

	topLabel := "the role"
	if len(dims) > 0 {
		topLabel = strings.ToLower(dims[0].Label)
	}
	strongest := "your recorded evidence"
	if len(strengths) > 0 {
		strongest = strings.ToLower(strengths[0].Label)
	}
	gapText := " No hard requirement gaps were detected; verify the original posting."
	if len(gaps) > 0 {
		gapText = " Main caveat: " + strings.ToLower(gaps[0].Label) + " — " + gaps[0].Detail
	}
	summary := "The description leans on " + topLabel + ". Your strongest overlap is " + strongest +
		", and the score weighs that evidence against the stated requirements." + gapText

	next := []string{}
	if len(strengths) > 0 {
		next = append(next, "Lead with "+strings.ToLower(strengths[0].Label)+": "+strengths[0].Evidence)
	}
	if len(matched) > 0 {
		next = append(next, "Mirror the posting's language where it is true: "+strings.Join(first(matched, 5), ", ")+".")
	}
	if len(gaps) > 0 {
		next = append(next, "Address "+strings.ToLower(gaps[0].Label)+" before the first screen: "+gaps[0].Detail)
	} else {
		next = append(next, "Keep the evidence specific: one concrete outcome and one measurable result beat a list of tools.")
	}
	if p.Field != "" {
		next = append(next, "Connect the posting back to your "+p.Field+" training: name the coursework or project that maps to "+topLabel+".")
	}

	return Analysis{
		OK: true, Title: extractTitle(source), Score: score, Band: band, Tone: tone,
		Headline: headlineFor(band), Summary: summary, Confidence: confidence,
		Words: words, SignalCount: len(matched), MatchedKeywords: matched,
		Dimensions: dims, Strengths: first(strengths, 3), Requirements: checks,
		Gaps: gaps, NextMoves: next,
		Caveat: "Fit is a positioning aid built from structured profile evidence and keyword rules. It is not an ATS score, hiring probability or a substitute for the original posting.",
	}
}

func alignOf(d dimension, groups map[string]bool) int {
	if groups[d.Group] {
		return 8
	}
	return 0
}

func evidenceFor(d dimension, p profile.Profile, groups map[string]bool) string {
	if !groups[d.Group] || p.Field == "" {
		return d.Evidence
	}
	lead := "Your " + p.Field + " training"
	parts := []string{}
	if p.Role != "" {
		parts = append(parts, p.Role)
	}
	if p.Company != "" {
		parts = append(parts, p.Company)
	}
	if len(parts) > 0 {
		lead += " and work as " + strings.Join(parts, " at ")
	}
	return lead + " support this area. " + d.Evidence
}

func bandFor(score int) (string, string) {
	switch {
	case score >= 80:
		return "Strong match", "strong"
	case score >= 65:
		return "Good match", "medium"
	case score >= 50:
		return "Stretch", "medium"
	default:
		return "Long shot", "low"
	}
}

func headlineFor(band string) string {
	switch band {
	case "Strong match":
		return "The evidence lines up."
	case "Good match":
		return "A credible match with one story to sharpen."
	case "Stretch":
		return "Reachable, but not on today's evidence alone."
	default:
		return "This is a longer bridge than it looks."
	}
}

func extractTitle(source string) string {
	firstLine := strings.SplitN(source, "\n", 2)[0]
	return truncate(firstLine, 90)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func first[T any](s []T, n int) []T {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func floatToInt(f float64) int {
	return int(f + 0.5)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
