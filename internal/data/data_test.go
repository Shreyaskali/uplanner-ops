package data

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadGeneratedContent(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "content")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	fixture := `{"manifest":{"programs":1,"freeOnly":true},"programs":[{"id":"p1","institution":"Example","program":"MSc","country":"UK","group":"AI / analytics","milestones":[],"requirements":[]}],"reports":[],"resources":[],"paths":[],"roles":[]}`
	if err := os.WriteFile(filepath.Join(dir, "uplanner.json"), []byte(fixture), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Programs) != 1 || c.Programs[0].ID != "p1" {
		t.Fatalf("unexpected programs: %+v", c.Programs)
	}
	if !c.Manifest.FreeOnly {
		t.Fatal("free-only flag not preserved")
	}
}

func TestLoadRejectsEmptyContent(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "content")
	_ = os.MkdirAll(dir, 0o755)
	_ = os.WriteFile(filepath.Join(dir, "uplanner.json"), []byte(`{"programs":[]}`), 0o644)
	if _, err := Load(root); err == nil {
		t.Fatal("expected an error for empty content")
	}
}

func TestLookups(t *testing.T) {
	c := Content{
		Programs: []Program{{ID: "p1"}},
		Reports:  []Report{{ID: "r1"}},
	}
	if c.ProgramByID("p1") == nil || c.ProgramByID("missing") != nil {
		t.Fatal("ProgramByID failed")
	}
	if c.ReportByID("r1") == nil || c.ReportByID("missing") != nil {
		t.Fatal("ReportByID failed")
	}
}
