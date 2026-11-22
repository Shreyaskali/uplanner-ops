package degree

import (
	"testing"

	"github.com/shreyaskali/uplanner-ops/internal/data"
)

func content() *data.Content {
	return &data.Content{
		Programs: []data.Program{
			{ID: "a", Institution: "A", Program: "MSc", Country: "UK", Group: "AI / analytics", Priority: "P2", ReportID: "r1", Milestones: []data.Milestone{{Type: "application", Date: "2027-01-10", Label: "Round 1"}}},
			{ID: "b", Institution: "B", Program: "MBA", Country: "India", Group: "MBA / management", Priority: "P1", Milestones: []data.Milestone{}},
			{ID: "c", Institution: "C", Program: "MSc", Country: "UK", Group: "Finance / quant", Priority: "P3", ReportID: "r2"},
		},
		Reports: []data.Report{
			{ID: "r1", EvidenceKind: "placement_report", Salary: "₹30L"},
			{ID: "r2", EvidenceKind: "no_report"},
		},
	}
}

func TestRankEvidenceFirst(t *testing.T) {
	items := Rank(content(), "", "", "2026-11-22", 10)
	if len(items) != 3 {
		t.Fatalf("expected 3, got %d", len(items))
	}
	if items[0].ID != "a" {
		t.Fatalf("placement_report should rank first, got %s", items[0].ID)
	}
	if items[0].Evidence != "Published outcomes" || items[0].Salary != "₹30L" {
		t.Fatalf("evidence not surfaced: %+v", items[0])
	}
}

func TestRankFilters(t *testing.T) {
	items := Rank(content(), "UK", "", "2026-11-22", 10)
	if len(items) != 2 {
		t.Fatalf("expected 2 UK programs, got %d", len(items))
	}
	items = Rank(content(), "", "MBA / management", "2026-11-22", 10)
	if len(items) != 1 || items[0].ID != "b" {
		t.Fatalf("group filter failed: %+v", items)
	}
}

func TestRankLimitAndNextRound(t *testing.T) {
	items := Rank(content(), "", "", "2026-11-22", 1)
	if len(items) != 1 {
		t.Fatalf("limit failed: %d", len(items))
	}
	if items[0].NextRound != "2027-01-10 · Round 1" {
		t.Fatalf("next round wrong: %q", items[0].NextRound)
	}
}

func TestCompare(t *testing.T) {
	result, err := Compare(content(), "a", "b")
	if err != nil {
		t.Fatal(err)
	}
	a := result["a"].(map[string]any)
	if a["salary"] != "₹30L" || a["evidence"] != "Published outcomes" {
		t.Fatalf("unexpected view: %+v", a)
	}
	if a["tuition"] != "Not verified" || a["studyLength"] != "Verify on program page" {
		t.Fatalf("missing labels: %+v", a)
	}
	if _, err := Compare(content(), "a", "missing"); err == nil {
		t.Fatal("expected not-found error")
	}
}
