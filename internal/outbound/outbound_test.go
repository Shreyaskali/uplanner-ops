package outbound

import (
	"testing"

	"github.com/shreyaskali/uplanner-ops/internal/profile"
)

func TestCheckWithoutKey(t *testing.T) {
	t.Setenv("UPLANNER_LLM_KEY", "")
	gate := Check("clean text", profile.Profile{})
	if gate.Allowed || gate.Reason != "no LLM key configured" {
		t.Fatalf("expected blocked without key: %+v", gate)
	}
}

func TestCheckBlocksPII(t *testing.T) {
	t.Setenv("UPLANNER_LLM_KEY", "test-key")
	gate := Check("contact me at shreyas@example.com", profile.Profile{})
	if gate.Allowed {
		t.Fatal("email should block the request")
	}
	found := false
	for _, c := range gate.Categories {
		if c == "email" {
			found = true
		}
	}
	if !found {
		t.Fatalf("email category missing: %+v", gate)
	}
}

func TestCheckBlocksProfileText(t *testing.T) {
	t.Setenv("UPLANNER_LLM_KEY", "test-key")
	gate := Check("clean text", profile.Profile{Role: "Cloud strategy analyst"})
	if gate.Allowed {
		t.Fatal("profile text should block the request")
	}
	clean := Check("clean text", profile.Profile{Years: 3})
	if !clean.Allowed {
		t.Fatalf("a number-only profile should pass: %+v", clean)
	}
}

func TestRequestFailsClosed(t *testing.T) {
	t.Setenv("UPLANNER_LLM_KEY", "")
	_, err := Request("https://example.invalid", []byte("{}"), profile.Profile{})
	if err == nil {
		t.Fatal("expected an error without a key")
	}
}
