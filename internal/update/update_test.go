package update

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPlanBlocksUserLayer(t *testing.T) {
	root := t.TempDir()
	source := t.TempDir()
	for _, p := range SystemPaths {
		write(source, p, "system")
		write(root, p, "system")
	}
	// A source checkout with user data must be reported as blocked.
	write(source, "profile.yml", "email: someone")
	write(source, "data/applications.md", "| id | status |")

	changes, err := Plan(root, source)
	if err != nil {
		t.Fatal(err)
	}
	blocked := 0
	for _, c := range changes {
		if c.Action == "blocked" {
			blocked++
			if c.Path == "profile.yml" || c.Path == "data/applications.md" {
				continue
			}
			t.Fatalf("unexpected blocked path: %s", c.Path)
		}
		if c.Action == "add" || c.Action == "update" {
			t.Fatalf("expected only unchanged/blocked, got %+v", c)
		}
	}
	if blocked < 2 {
		t.Fatalf("expected user-layer paths to be blocked, got %d", blocked)
	}
}

func TestApplyOnlySystemLayer(t *testing.T) {
	root := t.TempDir()
	source := t.TempDir()
	for _, p := range SystemPaths {
		write(source, p, "new")
		write(root, p, "new")
	}
	// One system file is missing locally: it should be added.
	if err := os.Remove(filepath.Join(root, "modes", "doctor.md")); err != nil {
		t.Fatal(err)
	}
	write(source, "profile.yml", "email: someone")
	write(root, "profile.yml", "email: me")

	changes, err := Apply(root, source)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range changes {
		if c.Path == "profile.yml" {
			t.Fatal("user layer was touched by update")
		}
	}
	got, err := os.ReadFile(filepath.Join(root, "profile.yml"))
	if err != nil || string(got) != "email: me" {
		t.Fatalf("user profile changed: %q %v", got, err)
	}
	added, err := os.ReadFile(filepath.Join(root, "modes", "doctor.md"))
	if err != nil || string(added) != "new" {
		t.Fatalf("missing system file not added: %q %v", added, err)
	}
}

func TestLocalEditPreserved(t *testing.T) {
	root := t.TempDir()
	source := t.TempDir()
	write(source, "SKILL.md", "upstream")
	write(root, "SKILL.md", "my local edit")

	_, err := Apply(root, source)
	if err != nil {
		t.Fatal(err)
	}
	current, _ := os.ReadFile(filepath.Join(root, "SKILL.md"))
	if string(current) != "my local edit" {
		t.Fatalf("local edit overwritten: %q", current)
	}
	backup, err := os.ReadFile(filepath.Join(root, "SKILL.md.bak"))
	if err != nil || string(backup) != "my local edit" {
		t.Fatalf("backup missing: %q %v", backup, err)
	}
}

func write(dir, rel, content string) {
	path := filepath.Join(dir, rel)
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	_ = os.WriteFile(path, []byte(content), 0o644)
}
