package fit

import (
	"testing"

	"github.com/shreyaskali/uplanner-ops/internal/profile"
)

const jd = `Cloud Strategy Analyst — Financial Services

You will join a small strategy team supporting cloud and data platforms.

Responsibilities
• Analyse cloud spend and usage data with SQL and Excel; build dashboards and KPIs for engineering and finance leaders.
• Translate metrics into recommendations for senior stakeholders and partner with product, platform and risk teams.
• Support AI governance, model observability and cloud cost controls.

Requirements
• 1–3 years in consulting, banking technology or product.
• Bachelor's degree; master's preferred, not required.
• Strong SQL and Excel; exposure to AWS, Azure or GCP.
• Clear written communication and comfort with ambiguity.`

func findCheck(a Analysis, label string) *Check {
	for i := range a.Requirements {
		if a.Requirements[i].Label == label {
			return &a.Requirements[i]
		}
	}
	return nil
}

func TestAnalyseDeterministic(t *testing.T) {
	p := profile.Profile{Field: "Computer Science", Years: 3, Role: "Cloud strategy analyst"}
	a := Analyse(jd, p)
	b := Analyse(jd, p)
	if a.Score != b.Score || a.Band != b.Band || a.Summary != b.Summary {
		t.Fatal("analysis is not deterministic")
	}
}

func TestAnalyseRejectsShortText(t *testing.T) {
	a := Analyse("Product manager", profile.Profile{})
	if a.OK || a.Summary == "" {
		t.Fatal("short text should be rejected with guidance")
	}
}

func TestProfileDrivesResult(t *testing.T) {
	tech := profile.Profile{Field: "Computer Science and Engineering", Years: 3, Role: "Cloud strategy analyst", Skills: "SQL, AWS"}
	arts := profile.Profile{Field: "Fine Arts", Years: 1, Skills: "painting"}

	a := Analyse(jd, tech)
	b := Analyse(jd, arts)

	if c := findCheck(a, "SQL"); c == nil || c.Status != "met" {
		t.Fatalf("relevant field should meet SQL: %+v", c)
	}
	if c := findCheck(a, "Cloud platforms"); c == nil || c.Status != "met" {
		t.Fatalf("cloud role should meet cloud platforms: %+v", c)
	}
	if a.Score < b.Score {
		t.Fatalf("relevant profile (%d) should not score below unrelated (%d)", a.Score, b.Score)
	}
}

func TestWordBoundarySubstring(t *testing.T) {
	if got := countTerm("painting", "ai"); got != 0 {
		t.Fatalf("'painting' must not match 'ai', got %d", got)
	}
	if got := countTerm("AI governance and generative ai", "ai"); got != 2 {
		t.Fatalf("expected two 'ai' matches, got %d", got)
	}
	if hasField(profile.Profile{Field: "Fine Arts", Skills: "painting"}, "ai") {
		t.Fatal("Fine Arts / painting must not match the ai signal")
	}
}

func TestYearsComeFromProfile(t *testing.T) {
	seniorJD := `Senior Director, Private Equity
We are looking for a senior leader with 15+ years of experience to own the P&L and lead a large team.
Requirements: MBA required, CFA required, US citizen or permanent resident required.`
	short := Analyse(seniorJD, profile.Profile{Field: "Finance", Years: 10})
	long := Analyse(seniorJD, profile.Profile{Field: "Finance", Years: 16})
	if c := findCheck(short, "Experience"); c == nil || c.Status != "gap" {
		t.Fatalf("10 years should be short of 15+: %+v", c)
	}
	if c := findCheck(long, "Experience"); c == nil || c.Status != "met" {
		t.Fatalf("16 years should meet 15+: %+v", c)
	}
}

func TestAIIsGenuineGapForUnrelatedField(t *testing.T) {
	aiJD := `Machine Learning Engineer
You will build and deploy machine learning models in Python, using SQL for feature pipelines, and work with deep learning and mlops tooling in production.`
	arts := Analyse(aiJD, profile.Profile{Field: "Fine Arts", Skills: "painting"})
	if c := findCheck(arts, "AI / ML"); c == nil || c.Status != "gap" {
		t.Fatalf("unrelated field should make AI a gap: %+v", c)
	}
}

func TestBandsAndDimensions(t *testing.T) {
	a := Analyse(jd, profile.Profile{Field: "Computer Science", Years: 3})
	if a.Band == "" || a.Tone == "" || a.Headline == "" {
		t.Fatal("missing band, tone or headline")
	}
	if len(a.Dimensions) == 0 || len(a.Dimensions) > 3 {
		t.Fatalf("expected up to 3 dimensions, got %d", len(a.Dimensions))
	}
	if a.Dimensions[0].Demand != 100 {
		t.Fatalf("top demand should be 100, got %d", a.Dimensions[0].Demand)
	}
	if a.Caveat == "" || a.Confidence == "" {
		t.Fatal("missing caveat or confidence")
	}
}
