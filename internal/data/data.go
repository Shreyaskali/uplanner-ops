// Package data defines the generated content contract shared by every mode.
// User data never lives here: see DATA_CONTRACT.md.
package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Milestone is one dated step for a program.
type Milestone struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Date       string `json:"date"`
	Type       string `json:"type"`
	Confidence string `json:"confidence"`
	SourceURL  string `json:"sourceUrl"`
	IsFinal    bool   `json:"isFinal"`
}

// Requirement is an eligibility or preparation check.
type Requirement struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Kind        string `json:"kind"`
	DueOffset   int    `json:"dueOffsetDays"`
	DueDate     string `json:"dueDate"`
	Confidence  string `json:"confirmation"`
	SourceURL   string `json:"sourceUrl"`
	DateKind    string `json:"dateKind"`
	MilestoneID string `json:"milestoneId"`
}

// Program is one degree entry from the uplanner directory.
type Program struct {
	ID               string        `json:"id"`
	Institution      string        `json:"institution"`
	Program          string        `json:"program"`
	Country          string        `json:"country"`
	Region           string        `json:"region"`
	Group            string        `json:"group"`
	Pathway          string        `json:"pathway"`
	Priority         string        `json:"priority"`
	VerifiedAsOf     string        `json:"verifiedAsOf"`
	SourceURL        string        `json:"sourceUrl"`
	DeadlineURL      string        `json:"deadlineUrl"`
	FundingURL       string        `json:"fundingUrl"`
	EntryGate        string        `json:"entryGate"`
	Caveat           string        `json:"caveat"`
	PublishedTuition string        `json:"publishedTuition"`
	FundingNote      string        `json:"fundingNote"`
	StudyLength      string        `json:"studyLength"`
	Outcome          string        `json:"outcome"`
	OutcomeURL       string        `json:"outcomeUrl"`
	ReportID         string        `json:"reportId"`
	Milestones       []Milestone   `json:"milestones"`
	Requirements     []Requirement `json:"requirements"`
}

// Employer is a named employer in a hiring report.
type Employer struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Industry    string `json:"industry"`
	Evidence    string `json:"evidence"`
	Note        string `json:"note"`
}

// IndustryShare is a reported share with its own denominator.
type IndustryShare struct {
	Name         string  `json:"name"`
	Share        float64 `json:"sharePercent"`
	Basis        string  `json:"shareBasis"`
	Dimension    string  `json:"dimension"`
	Note         string  `json:"note"`
	DisplayLabel string  `json:"displayLabel"`
}

// Report is a reviewed hiring/outcomes source.
type Report struct {
	ID           string          `json:"id"`
	Institution  string          `json:"institution"`
	ProgramNames []string        `json:"programNames"`
	ReportYear   string          `json:"reportYear"`
	CohortScope  string          `json:"cohortScope"`
	EvidenceKind string          `json:"evidenceKind"`
	Roles        []string        `json:"roles"`
	Employers    []Employer      `json:"employers"`
	Industries   []IndustryShare `json:"industries"`
	Salary       string          `json:"salary"`
	Employment   string          `json:"employment"`
	SourceURL    string          `json:"sourceUrl"`
	VerifiedAsOf string          `json:"verifiedAsOf"`
}

// Resource is a free learning resource. Cost is always "free" in generated
// content; paid resources are dropped by the generator.
type Resource struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Source   string `json:"source"`
	Format   string `json:"format"`
	Duration string `json:"duration"`
	URL      string `json:"url"`
	Note     string `json:"note"`
	Domain   string `json:"domain"`
	Path     string `json:"path"`
	Cost     string `json:"cost"`
	Free     bool   `json:"free"`
}

// LearningPath is an Intersections domain intersection.
type LearningPath struct {
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	Tier        string   `json:"tier"`
	Tagline     string   `json:"tagline"`
	Domains     []string `json:"domains"`
	Hours       float64  `json:"hours"`
	Count       int      `json:"count"`
	ResourceIDs []string `json:"resourceIds"`
}

// Role is a market role profile: what the job really is and what counts.
type Role struct {
	Key        string   `json:"key"`
	Title      string   `json:"title"`
	Aliases    []string `json:"aliases"`
	Domains    []string `json:"domains"`
	Reality    string   `json:"reality"`
	ScreensFor []string `json:"screensFor"`
	Evidence   []string `json:"evidenceThatCounts"`
	CommonGaps []string `json:"commonGaps"`
	Signals    []string `json:"signals"`
	Leverage   int      `json:"leverage"`
	DemandNote string   `json:"demandNote"`
}

// Manifest records what generated content is available.
type Manifest struct {
	GeneratedAt         string `json:"generatedAt"`
	SourceUplanner      string `json:"sourceUplanner"`
	SourceIntersections string `json:"sourceIntersections"`
	Programs            int    `json:"programs"`
	Reports             int    `json:"reports"`
	Resources           int    `json:"resources"`
	Paths               int    `json:"paths"`
	Roles               int    `json:"roles"`
	FreeOnly            bool   `json:"freeOnly"`
}

// Content is the full generated dataset.
type Content struct {
	Manifest  Manifest       `json:"manifest"`
	Programs  []Program      `json:"programs"`
	Reports   []Report       `json:"reports"`
	Resources []Resource     `json:"resources"`
	Paths     []LearningPath `json:"paths"`
	Roles     []Role         `json:"roles"`
}

// Load reads generated content from a root directory.
func Load(root string) (*Content, error) {
	path := filepath.Join(root, "content", "uplanner.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var c Content
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if len(c.Programs) == 0 {
		return nil, errors.New("generated content has no programs; run `go run ./cmd/gen-data`")
	}
	return &c, nil
}

// ProgramByID returns a program by id, or nil.
func (c *Content) ProgramByID(id string) *Program {
	for i := range c.Programs {
		if c.Programs[i].ID == id {
			return &c.Programs[i]
		}
	}
	return nil
}

// ReportByID returns a report by id, or nil.
func (c *Content) ReportByID(id string) *Report {
	for i := range c.Reports {
		if c.Reports[i].ID == id {
			return &c.Reports[i]
		}
	}
	return nil
}
