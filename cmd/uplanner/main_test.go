package main

import (
	"strings"
	"testing"
)

func TestUpsertYAML(t *testing.T) {
	current := "# comment\nversion: 1\nfield: Old\n"
	updated := upsertYAML(current, map[string]string{
		"field": "Computer Science",
		"years": "3",
	})
	if !strings.Contains(updated, "field: Computer Science") {
		t.Fatalf("field not updated: %q", updated)
	}
	if strings.Contains(updated, "Old") {
		t.Fatalf("old value remains: %q", updated)
	}
	if !strings.Contains(updated, "years: 3") {
		t.Fatalf("new key missing: %q", updated)
	}
	if !strings.Contains(updated, "# comment") {
		t.Fatal("comment lost")
	}
	if strings.Count(updated, "field:") != 1 {
		t.Fatalf("duplicate key: %q", updated)
	}
}

func TestQuoteYAML(t *testing.T) {
	cases := map[string]string{
		"":      `""`,
		"3":     "3",
		"3.5":   "3.5",
		"a: b":  `"a: b"`,
		"a # b": `"a # b"`,
		"plain": "plain",
	}
	for in, want := range cases {
		if got := quoteYAML(in); got != want {
			t.Errorf("quoteYAML(%q) = %q, want %q", in, got, want)
		}
	}
}
