package tracker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadAndWriteRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "applications.md")
	records := []Record{
		{ID: "p1", Status: "preparing", Deadline: "2027-01-06", Target: "2026-12-20", Documents: "transcripts,resume"},
		{ID: "p2", Status: "not started", Deadline: "", Target: "", Documents: ""},
	}
	if err := Write(path, records); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatal("temporary file left behind")
	}
	got, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 records, got %d", len(got))
	}
	// Dated deadlines sort first; undated rows go last.
	if got[0].ID != "p1" || got[1].ID != "p2" {
		t.Fatalf("unexpected order: %+v", got)
	}
	if got[0].Documents != "transcripts,resume" {
		t.Fatalf("documents not preserved: %q", got[0].Documents)
	}
}

func TestReadMissingFile(t *testing.T) {
	records, err := Read(filepath.Join(t.TempDir(), "none.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 0 {
		t.Fatalf("expected empty, got %+v", records)
	}
}

func TestUpsert(t *testing.T) {
	records := []Record{{ID: "a", Status: "planning"}, {ID: "b", Status: "ready"}}
	updated := Upsert(records, Record{ID: "a", Status: "submitted"})
	if len(updated) != 2 {
		t.Fatalf("expected 2, got %d", len(updated))
	}
	for _, r := range updated {
		if r.ID == "a" && r.Status != "submitted" {
			t.Fatal("not updated")
		}
	}
	added := Upsert(records, Record{ID: "c", Status: "planning"})
	if len(added) != 3 {
		t.Fatalf("expected 3, got %d", len(added))
	}
}

func TestStatusVocabulary(t *testing.T) {
	for _, s := range []string{"not started", "Preparing", "READY", "submitted"} {
		if !IsValidStatus(s) {
			t.Errorf("%q should be valid", s)
		}
	}
	for _, s := range []string{"", "done", "maybe"} {
		if IsValidStatus(s) {
			t.Errorf("%q should be invalid", s)
		}
	}
}
