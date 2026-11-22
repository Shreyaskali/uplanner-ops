// gen-data converts uplanner's recruitment-data.js and Intersections'
// TypeScript data into one generated JSON contract under content/.
//
// It never writes to the user layer. Run from anywhere:
//
//	go run ./cmd/gen-data --uplanner /path/to/University\ Planner \
//	  --intersections /path/to/intersections --out ./content
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/shreyaskali/uplanner-ops/internal/data"
)

func main() {
	uplanner := flag.String("uplanner", "", "path to the University Planner directory")
	intersections := flag.String("intersections", "", "path to the Intersections directory")
	out := flag.String("out", "content", "output directory for generated JSON")
	flag.Parse()

	if *uplanner == "" || *intersections == "" {
		fmt.Fprintln(os.Stderr, "usage: gen-data --uplanner DIR --intersections DIR [--out DIR]")
		os.Exit(2)
	}

	content, err := build(*uplanner, *intersections)
	if err != nil {
		fmt.Fprintln(os.Stderr, "gen-data:", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "gen-data:", err)
		os.Exit(1)
	}
	raw, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "gen-data:", err)
		os.Exit(1)
	}
	path := filepath.Join(*out, "uplanner.json")
	if err := os.WriteFile(path, append(raw, '\n'), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "gen-data:", err)
		os.Exit(1)
	}
	fmt.Printf("Generated %s: %d programs, %d reports, %d free resources, %d paths, %d roles\n",
		path, len(content.Programs), len(content.Reports), len(content.Resources), len(content.Paths), len(content.Roles))
}

func build(uplannerDir, intersectionsDir string) (data.Content, error) {
	out := data.Content{
		Manifest: data.Manifest{
			GeneratedAt:         time.Now().UTC().Format(time.RFC3339),
			SourceUplanner:      uplannerDir,
			SourceIntersections: intersectionsDir,
			FreeOnly:            true,
		},
	}

	programs, reports, err := parseUplanner(filepath.Join(uplannerDir, "recruitment-data.js"))
	if err != nil {
		return out, err
	}
	out.Programs, out.Reports = programs, reports

	paths, resources, err := parseIntersections(
		filepath.Join(intersectionsDir, "src", "lib", "data.ts"),
		filepath.Join(intersectionsDir, "src", "lib", "compass", "roles.ts"),
	)
	if err != nil {
		return out, err
	}
	out.Paths, out.Resources = paths, resources

	roles, err := parseRoles(filepath.Join(intersectionsDir, "src", "lib", "compass", "roles.ts"))
	if err != nil {
		return out, err
	}
	out.Roles = roles

	out.Manifest.Programs = len(out.Programs)
	out.Manifest.Reports = len(out.Reports)
	out.Manifest.Resources = len(out.Resources)
	out.Manifest.Paths = len(out.Paths)
	out.Manifest.Roles = len(out.Roles)
	return out, nil
}

// ---------- uplanner recruitment-data.js ----------

type rawData struct {
	Programs []struct {
		ID               string `json:"id"`
		Institution      string `json:"institution"`
		Program          string `json:"program"`
		Country          string `json:"country"`
		Region           string `json:"region"`
		Pathway          string `json:"pathway"`
		Priority         string `json:"priority"`
		VerifiedAsOf     string `json:"verifiedAsOf"`
		SourceURL        string `json:"sourceUrl"`
		DeadlineURL      string `json:"deadlineUrl"`
		FundingURL       string `json:"fundingUrl"`
		EntryGate        string `json:"entryGate"`
		Caveat           string `json:"caveat"`
		PublishedTuition string `json:"publishedTuition"`
		FundingNote      string `json:"fundingNote"`
		StudyLength      string `json:"studyLength"`
		Outcome          string `json:"outcome"`
		OutcomeURL       string `json:"outcomeUrl"`
		ReportID         string `json:"reportId"`
		Group            string `json:"group"`
		Milestones       []struct {
			ID         string `json:"id"`
			Label      string `json:"label"`
			Date       string `json:"date"`
			Type       string `json:"type"`
			Confidence string `json:"confidence"`
			SourceURL  string `json:"sourceUrl"`
			IsFinal    bool   `json:"isFinal"`
		} `json:"milestones"`
		Requirements []struct {
			ID          string `json:"id"`
			Label       string `json:"label"`
			Kind        string `json:"kind"`
			DueOffset   int    `json:"dueOffsetDays"`
			DueDate     string `json:"dueDate"`
			Confidence  string `json:"confirmation"`
			SourceURL   string `json:"sourceUrl"`
			DateKind    string `json:"dateKind"`
			MilestoneID string `json:"milestoneId"`
		} `json:"requirements"`
	} `json:"programs"`
	Reports []struct {
		ID           string          `json:"id"`
		Institution  string          `json:"institution"`
		ProgramNames []string        `json:"program_names"`
		ReportYear   json.RawMessage `json:"report_year"`
		CohortScope  string          `json:"cohort_scope"`
		EvidenceKind string          `json:"evidence_kind"`
		Roles        []string        `json:"roles"`
		Employers    []struct {
			Name        string `json:"name"`
			DisplayName string `json:"display_name"`
			Industry    string `json:"industry"`
			Evidence    string `json:"evidence"`
			Note        string `json:"note"`
		} `json:"employers"`
		Industries []struct {
			Name         string  `json:"name"`
			Share        float64 `json:"share_percent"`
			Basis        string  `json:"share_basis"`
			Dimension    string  `json:"dimension"`
			Note         string  `json:"note"`
			DisplayLabel string  `json:"display_label"`
		} `json:"industries"`
		Salary       json.RawMessage `json:"salary"`
		Employment   json.RawMessage `json:"employment"`
		SourceURL    string          `json:"source_url"`
		VerifiedAsOf string          `json:"verified_as_of"`
	} `json:"reports"`
}

func parseUplanner(path string) ([]data.Program, []data.Report, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	text := string(raw)
	const marker = "window.RECRUITMENT_DATA = "
	idx := strings.Index(text, marker)
	if idx < 0 {
		return nil, nil, errors.New("recruitment-data.js has no window.RECRUITMENT_DATA assignment")
	}
	body := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(text[idx+len(marker):]), ";"))
	var src rawData
	if err := json.Unmarshal([]byte(body), &src); err != nil {
		return nil, nil, fmt.Errorf("parse recruitment-data.js: %w", err)
	}

	programs := make([]data.Program, 0, len(src.Programs))
	for _, p := range src.Programs {
		program := data.Program{
			ID: p.ID, Institution: p.Institution, Program: p.Program, Country: p.Country,
			Region: p.Region, Group: p.Group, Pathway: p.Pathway, Priority: p.Priority,
			VerifiedAsOf: p.VerifiedAsOf, SourceURL: p.SourceURL, DeadlineURL: p.DeadlineURL,
			FundingURL: p.FundingURL, EntryGate: p.EntryGate, Caveat: p.Caveat,
			PublishedTuition: p.PublishedTuition, FundingNote: p.FundingNote,
			StudyLength: p.StudyLength, Outcome: p.Outcome, OutcomeURL: p.OutcomeURL,
			ReportID: p.ReportID, Milestones: []data.Milestone{}, Requirements: []data.Requirement{},
		}
		for _, m := range p.Milestones {
			program.Milestones = append(program.Milestones, data.Milestone{
				ID: m.ID, Label: m.Label, Date: m.Date, Type: m.Type,
				Confidence: m.Confidence, SourceURL: m.SourceURL, IsFinal: m.IsFinal,
			})
		}
		for _, r := range p.Requirements {
			program.Requirements = append(program.Requirements, data.Requirement{
				ID: r.ID, Label: r.Label, Kind: r.Kind, DueOffset: r.DueOffset,
				DueDate: r.DueDate, Confidence: r.Confidence, SourceURL: r.SourceURL,
				DateKind: r.DateKind, MilestoneID: r.MilestoneID,
			})
		}
		programs = append(programs, program)
	}

	reports := make([]data.Report, 0, len(src.Reports))
	for _, r := range src.Reports {
		report := data.Report{
			ID: r.ID, Institution: r.Institution, ProgramNames: r.ProgramNames,
			ReportYear: stringValue(r.ReportYear), CohortScope: r.CohortScope, EvidenceKind: r.EvidenceKind,
			Roles: r.Roles, Salary: readableMetric(r.Salary), Employment: readableMetric(r.Employment),
			SourceURL: r.SourceURL, VerifiedAsOf: r.VerifiedAsOf,
			Employers: []data.Employer{}, Industries: []data.IndustryShare{},
		}
		for _, e := range r.Employers {
			report.Employers = append(report.Employers, data.Employer{
				Name: e.Name, DisplayName: e.DisplayName, Industry: e.Industry,
				Evidence: e.Evidence, Note: e.Note,
			})
		}
		for _, i := range r.Industries {
			report.Industries = append(report.Industries, data.IndustryShare{
				Name: i.Name, Share: i.Share, Basis: i.Basis, Dimension: i.Dimension,
				Note: i.Note, DisplayLabel: i.DisplayLabel,
			})
		}
		reports = append(reports, report)
	}
	return programs, reports, nil
}

// stringValue converts a JSON scalar to text regardless of its original type.
func stringValue(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var value any
	if json.Unmarshal(raw, &value) != nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	}
	return ""
}

// readableMetric turns {display:"...", ...} into the display string.
func readableMetric(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var obj map[string]any
	if json.Unmarshal(raw, &obj) != nil {
		return ""
	}
	if v, ok := obj["display"].(string); ok {
		return v
	}
	return ""
}

// ---------- Intersections data.ts ----------

var (
	reSlug    = regexp.MustCompile(`slug:\s*"([^"]+)"`)
	reName    = regexp.MustCompile(`name:\s*"([^"]+)"`)
	reTier    = regexp.MustCompile(`tier:\s*"([^"]+)"`)
	reTagline = regexp.MustCompile(`tagline:\s*"([^"]+)"`)
	reDomains = regexp.MustCompile(`domains:\s*\[([^\]]*)\]`)
	reHours   = regexp.MustCompile(`hours:\s*([0-9.]+)`)
	reCount   = regexp.MustCompile(`count:\s*([0-9]+)`)
	reQuoted  = regexp.MustCompile(`"([^"]+)"|'([^']+)'`)
	reEntry   = regexp.MustCompile(`\{([^{}]*)\}`)
	reKey     = regexp.MustCompile(`([a-zA-Z]+):\s*`)
)

var knownDomains = map[string]bool{
	"ai": true, "finance": true, "cloud": true, "arch": true, "lead": true,
	"entre": true, "wm": true, "sc": true, "product": true, "program": true,
	"data": true, "sec": true, "design": true, "growth": true,
}

func parseIntersections(dataPath, rolesPath string) ([]data.LearningPath, []data.Resource, error) {
	raw, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, nil, err
	}
	text := string(raw)

	slugs := reSlug.FindAllStringSubmatchIndex(text, -1)
	paths := make([]data.LearningPath, 0, len(slugs))
	resources := []data.Resource{}

	for i, loc := range slugs {
		end := len(text)
		if i+1 < len(slugs) {
			end = slugs[i+1][0]
		}
		chunk := text[loc[0]:end]
		slug := text[loc[2]:loc[3]]
		path := data.LearningPath{
			Slug: slug, Tier: fieldString(reTier, chunk), Tagline: fieldString(reTagline, chunk),
			Domains: []string{}, ResourceIDs: []string{},
		}
		if m := reName.FindStringSubmatch(chunk); len(m) > 1 {
			path.Name = m[1]
		}
		if m := reHours.FindStringSubmatch(chunk); len(m) > 1 {
			path.Hours, _ = strconv.ParseFloat(m[1], 64)
		}
		if m := reCount.FindStringSubmatch(chunk); len(m) > 1 {
			path.Count, _ = strconv.Atoi(m[1])
		}
		if m := reDomains.FindStringSubmatch(chunk); len(m) > 1 {
			for _, value := range reQuoted.FindAllStringSubmatch(m[1], -1) {
				domain := value[1]
				if domain == "" {
					domain = value[2]
				}
				if knownDomains[domain] {
					path.Domains = append(path.Domains, domain)
				}
			}
		}

		domain := "general"
		for _, d := range path.Domains {
			if knownDomains[d] {
				domain = d
				break
			}
		}

		resStart := strings.Index(chunk, "resources: r([")
		if resStart >= 0 {
			tail := chunk[resStart:]
			resEnd := endOfArray(tail, "resources: r([")
			if resEnd > 0 {
				block := tail[:resEnd]
				for _, entry := range reEntry.FindAllStringSubmatch(block, -1) {
					fields := parseFields(entry[1])
					title, url := fields["title"], fields["url"]
					if title == "" || url == "" {
						continue
					}
					format := fields["format"]
					if format == "" {
						format = "Article"
					}
					if !isFree(url, format) {
						continue
					}
					id := fmt.Sprintf("%s-%02d", slug, len(resources)+1)
					res := data.Resource{
						ID: id, Title: title, Source: fields["source"], Format: format,
						Duration: fields["duration"], URL: url, Note: fields["note"],
						Domain: domain, Path: path.Name, Cost: "free", Free: true,
					}
					resources = append(resources, res)
					path.ResourceIDs = append(path.ResourceIDs, id)
				}
			}
		}
		if len(path.ResourceIDs) > 0 {
			paths = append(paths, path)
		}
	}
	return paths, resources, nil
}

// endOfArray returns the index just before the matching "]" for a marker that
// opens with "[". It counts nested brackets but ignores quoted strings.
func endOfArray(s, marker string) int {
	start := strings.Index(s, marker)
	if start < 0 {
		return -1
	}
	depth := 0
	inQuote := byte(0)
	for i := start + len(marker) - 1; i < len(s); i++ {
		c := s[i]
		if inQuote != 0 {
			if c == inQuote && (i == 0 || s[i-1] != '\\') {
				inQuote = 0
			}
			continue
		}
		switch c {
		case '"', '\'':
			inQuote = c
		case '[':
			depth++
		case ']':
			if depth == 0 {
				return i
			}
			depth--
			if depth == 0 {
				return i + 1
			}
		}
	}
	return -1
}

func parseFields(body string) map[string]string {
	out := map[string]string{}
	for _, loc := range reKey.FindAllStringSubmatchIndex(body, -1) {
		key := body[loc[2]:loc[3]]
		rest := strings.TrimSpace(body[loc[1]:])
		if len(rest) == 0 || (rest[0] != '"' && rest[0] != '\'') {
			continue
		}
		quote := rest[0]
		idx := 1
		for idx < len(rest) {
			if rest[idx] == quote && rest[idx-1] != '\\' {
				break
			}
			idx++
		}
		if idx >= len(rest) {
			continue
		}
		value := rest[1:idx]
		if quote == '\'' {
			value = strings.ReplaceAll(value, `\'`, `'`)
		}
		out[key] = value
	}
	return out
}

func fieldString(re *regexp.Regexp, s string) string {
	if m := re.FindStringSubmatch(s); len(m) > 1 {
		return m[1]
	}
	return ""
}

// isFree is deliberately conservative: books and retail links are paid; open
// courses, papers, videos and independent writing are free.
func isFree(url, format string) bool {
	lower := strings.ToLower(url)
	if strings.Contains(lower, "amzn.") || strings.Contains(lower, "amazon.") ||
		strings.Contains(lower, "kindle") {
		return false
	}
	if strings.EqualFold(format, "Book") {
		return false
	}
	if strings.Contains(lower, "garp.org/frm") || strings.Contains(lower, "cfainstitute.org/programs") {
		return false
	}
	return true
}

// ---------- roles.ts ----------

func parseRoles(path string) ([]data.Role, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	text := string(raw)
	start := strings.Index(text, "ROLES")
	if start < 0 {
		return nil, errors.New("roles.ts has no ROLES export")
	}
	text = text[start:]

	chunks := regexp.MustCompile(`(?m)^\s*\{\s*$`).Split(text, -1)
	roles := []data.Role{}
	for _, chunk := range chunks {
		if !strings.Contains(chunk, `key:`) || !strings.Contains(chunk, `title:`) {
			continue
		}
		role := data.Role{
			Key:        fieldString(regexp.MustCompile(`key:\s*"([^"]+)"`), chunk),
			Title:      fieldString(regexp.MustCompile(`title:\s*"([^"]+)"`), chunk),
			Reality:    fieldString(regexp.MustCompile(`reality:\s*"([^"]*)"[,]?`), chunk),
			DemandNote: fieldString(regexp.MustCompile(`demandNote:\s*"([^"]*)"[,]?`), chunk),
			Aliases:    []string{}, Domains: []string{}, ScreensFor: []string{},
			Evidence: []string{}, CommonGaps: []string{}, Signals: []string{},
		}
		if m := regexp.MustCompile(`leverage:\s*([0-9])`).FindStringSubmatch(chunk); len(m) > 1 {
			role.Leverage, _ = strconv.Atoi(m[1])
		}
		for key, target := range map[string]*[]string{
			"aliases": &role.Aliases, "domains": &role.Domains,
			"screensFor": &role.ScreensFor, "evidenceThatCounts": &role.Evidence,
			"commonGaps": &role.CommonGaps, "signals": &role.Signals,
		} {
			re := regexp.MustCompile(key + `:\s*\[([\s\S]*?)\]`)
			if m := re.FindStringSubmatch(chunk); len(m) > 1 {
				for _, value := range reQuoted.FindAllStringSubmatch(m[1], -1) {
					item := value[1]
					if item == "" {
						item = value[2]
					}
					*target = append(*target, item)
				}
			}
		}
		if role.Key != "" && role.Title != "" {
			roles = append(roles, role)
		}
	}
	return roles, nil
}
