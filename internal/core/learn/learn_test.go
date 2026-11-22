package learn

import (
	"testing"

	"github.com/shreyaskali/uplanner-ops/internal/data"
)

func sampleContent() *data.Content {
	return &data.Content{
		Resources: []data.Resource{
			{ID: "r1", Title: "SQL course", Domain: "data", Path: "AI × Data", Free: true, Cost: "free", Duration: "10h"},
			{ID: "r2", Title: "Paid book", Domain: "data", Path: "AI × Data", Free: false, Cost: "paid"},
			{ID: "r3", Title: "AI video", Domain: "ai", Path: "AI × Finance", Free: true, Cost: "free", Duration: "30 min"},
			{ID: "r4", Title: "Leadership", Domain: "lead", Path: "Leadership × Entrepreneurship", Free: true, Cost: "free", Duration: "Ongoing"},
		},
		Paths: []data.LearningPath{
			{Slug: "p1", Name: "AI × Data", Tier: "free", Domains: []string{"ai", "data"}, ResourceIDs: []string{"r1", "r2"}},
			{Slug: "p2", Name: "Paid Path", Tier: "paid", Domains: []string{"data"}, ResourceIDs: []string{"r3"}},
			{Slug: "p3", Name: "Leadership × Entrepreneurship", Tier: "free", Domains: []string{"lead"}, ResourceIDs: []string{"r4"}},
		},
	}
}

func TestGapDomains(t *testing.T) {
	domains := GapDomains("I need SQL and dashboards")
	found := false
	for _, d := range domains {
		if d == "data" {
			found = true
		}
	}
	if !found {
		t.Fatalf("SQL gap should map to data, got %v", domains)
	}
	if got := GapDomains("something completely unrelated"); len(got) != 0 {
		t.Fatalf("unexpected domains: %v", got)
	}
}

func TestDomainsForGroup(t *testing.T) {
	domains := DomainsForGroup("Technology / product")
	if len(domains) == 0 {
		t.Fatal("expected domains for technology group")
	}
	if got := DomainsForGroup("unknown"); got != nil {
		t.Fatalf("unexpected domains: %v", got)
	}
}

func TestSelectFreeOnlyAndCompleted(t *testing.T) {
	c := sampleContent()
	items := Select(c, []string{"data"}, nil, 10)
	if len(items) != 1 || items[0].ID != "r1" {
		t.Fatalf("expected only free r1, got %+v", items)
	}
	withCompleted := Select(c, []string{"data"}, map[string]bool{"r1": true}, 10)
	if len(withCompleted) != 0 {
		t.Fatalf("completed resource should be excluded, got %+v", withCompleted)
	}
}

func TestSelectRespectsLimit(t *testing.T) {
	c := sampleContent()
	items := Select(c, []string{"data", "ai", "lead"}, nil, 2)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestMatchingPathsFreeOnly(t *testing.T) {
	c := sampleContent()
	paths := MatchingPaths(c, []string{"data"}, nil)
	for _, p := range paths {
		if p.Tier != "free" {
			t.Fatalf("paid path leaked into results: %s", p.Name)
		}
	}
	if len(paths) != 1 || paths[0].Slug != "p1" {
		t.Fatalf("unexpected paths: %+v", paths)
	}
}

func TestParseHours(t *testing.T) {
	cases := map[string]float64{
		"10h":         10,
		"~40h":        40,
		"30 min":      0.5,
		"1h 30m":      1,
		"Ongoing":     0,
		"Per episode": 0,
	}
	for label, want := range cases {
		if got := parseHours(label); got != want {
			t.Errorf("parseHours(%q) = %v, want %v", label, got, want)
		}
	}
}
