package main

import (
	"testing"
)

func TestContainsString(t *testing.T) {
	list := []string{"ai", "data"}
	if !containsString(list, "ai") || containsString(list, "cloud") {
		t.Fatal("containsString failed")
	}
}

func TestExamsHavePhases(t *testing.T) {
	if len(exams) == 0 {
		t.Fatal("no exams defined")
	}
	for _, e := range exams {
		if e.Key == "" || e.Name == "" || len(e.Phases) == 0 {
			t.Fatalf("incomplete exam: %+v", e)
		}
		if e.Note == "" {
			t.Fatalf("exam %s has no verification note", e.Key)
		}
	}
}

func TestSplitListAndMerge(t *testing.T) {
	got := splitList("ai, data, ,cloud")
	if len(got) != 3 {
		t.Fatalf("unexpected split: %v", got)
	}
	merged := mergeUnique([]string{"data", "ai"}, []string{"ai", "cloud"})
	if len(merged) != 3 || merged[0] != "ai" {
		t.Fatalf("unexpected merge: %v", merged)
	}
}

func TestDash(t *testing.T) {
	if dash("  ") != "—" || dash("x") != "x" {
		t.Fatal("dash helper failed")
	}
}
