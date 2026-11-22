package main

import (
	"strings"
	"testing"
)

func TestIsFree(t *testing.T) {
	free := []struct{ url, format string }{
		{"https://arxiv.org/abs/1706.03762", "PDF"},
		{"https://www.youtube.com/watch?v=abc", "YouTube"},
		{"https://ocw.mit.edu/courses/x", "Course"},
		{"https://jalammar.github.io/illustrated-transformer/", "Article"},
		{"https://course.fast.ai/", "Course"},
	}
	for _, tc := range free {
		if !isFree(tc.url, tc.format) {
			t.Errorf("expected free: %s %s", tc.url, tc.format)
		}
	}
	paid := []struct{ url, format string }{
		{"https://amzn.in/d/abc", "Book"},
		{"https://www.amazon.com/dp/x", "Book"},
		{"https://example.com/x", "Book"},
		{"https://www.garp.org/frm", "Course"},
		{"https://cfainstitute.org/programs/cfa", "Course"},
	}
	for _, tc := range paid {
		if isFree(tc.url, tc.format) {
			t.Errorf("expected not free: %s %s", tc.url, tc.format)
		}
	}
}

func TestEndOfArray(t *testing.T) {
	s := `resources: r([
  { title: "A", url: "https://x" },
  { title: "B", url: "https://y" }
]),`
	end := endOfArray(s, "resources: r([")
	if end < 0 {
		t.Fatal("no end found")
	}
	if !strings.HasPrefix(strings.TrimSpace(s[end:]), "),") {
		t.Fatalf("unexpected tail: %q", s[end:])
	}
}

func TestEndOfArrayQuotedBracket(t *testing.T) {
	s := `resources: r([ { title: "a ] b", url: "u" } ]),`
	end := endOfArray(s, "resources: r([")
	if end < 1 || s[end-1] != ']' {
		t.Fatalf("unexpected end: %d %q", end, s)
	}
}

func TestParseFields(t *testing.T) {
	fields := parseFields(`title: "Hello", source: 'World', format: "Article", note: "A note"`)
	if fields["title"] != "Hello" || fields["source"] != "World" || fields["format"] != "Article" {
		t.Fatalf("unexpected fields: %+v", fields)
	}
}

func TestStringValue(t *testing.T) {
	if got := stringValue([]byte(`2026`)); got != "2026" {
		t.Fatalf("number: %q", got)
	}
	if got := stringValue([]byte(`"2026"`)); got != "2026" {
		t.Fatalf("string: %q", got)
	}
}
