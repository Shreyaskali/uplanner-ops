// Package outbound is the only network boundary in the tool.
// It fails closed: no request leaves the machine unless it passes the gate.
package outbound

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/shreyaskali/uplanner-ops/internal/profile"
)

var sensitivePatterns = []struct {
	name string
	re   *regexp.Regexp
}{
	{"email", regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`)},
	{"phone", regexp.MustCompile(`(?:\+?\d[\d\s-]{8,}\d)`)},
	{"card", regexp.MustCompile(`\b(?:\d[ -]*?){13,19}\b`)},
}

// Gate describes what was checked before a request.
type Gate struct {
	Allowed       bool     `json:"allowed"`
	Reason        string   `json:"reason"`
	Categories    []string `json:"categories"`
	KeyConfigured bool     `json:"keyConfigured"`
}

// Check inspects a payload and the environment without sending anything.
func Check(payload string, p profile.Profile) Gate {
	gate := Gate{Allowed: true, KeyConfigured: os.Getenv("UPLANNER_LLM_KEY") != ""}
	if !gate.KeyConfigured {
		gate.Allowed = false
		gate.Reason = "no LLM key configured"
		return gate
	}
	values := []string{p.Institution, p.Role, p.Company, p.Education, p.Experience, p.TestScores}
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			gate.Categories = append(gate.Categories, "profile-text")
			break
		}
	}
	for _, rule := range sensitivePatterns {
		if rule.re.MatchString(payload) {
			gate.Categories = append(gate.Categories, rule.name)
		}
	}
	if len(gate.Categories) > 0 {
		gate.Allowed = false
		gate.Reason = "payload contains protected information"
	}
	return gate
}

// Request sends a POST only when the gate allows it. It never logs the key
// or the payload, and the error names only a category, never a value.
func Request(url string, payload []byte, p profile.Profile) ([]byte, error) {
	gate := Check(string(payload), p)
	if !gate.Allowed {
		return nil, fmt.Errorf("outbound request blocked: %s (%s)", gate.Reason, strings.Join(gate.Categories, ", "))
	}
	key := os.Getenv("UPLANNER_LLM_KEY")
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("outbound request failed")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, errors.New("outbound response could not be read")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("outbound request failed with status %d", resp.StatusCode)
	}
	return body, nil
}
