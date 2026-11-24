// parity-go emits a normalised fit result for a parity fixture.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/shreyaskali/uplanner-ops/internal/core/fit"
	"github.com/shreyaskali/uplanner-ops/internal/profile"
)

type fixture struct {
	Name    string         `json:"name"`
	Profile map[string]any `json:"profile"`
	JD      string         `json:"jd"`
}

func str(values map[string]any, key string) string {
	if value, ok := values[key].(string); ok {
		return value
	}
	return ""
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: parity-go <fixture.json>")
		os.Exit(2)
	}
	raw, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var fx fixture
	if err := json.Unmarshal(raw, &fx); err != nil {
		panic(err)
	}
	p := profile.Profile{
		DegreeLevel: str(fx.Profile, "degreeLevel"),
		Field:       str(fx.Profile, "field"),
		Institution: str(fx.Profile, "institution"),
		Grade:       str(fx.Profile, "grade"),
		Role:        str(fx.Profile, "role"),
		Company:     str(fx.Profile, "company"),
		Skills:      str(fx.Profile, "skills"),
		Exam:        str(fx.Profile, "exam"),
		QuantScore:  str(fx.Profile, "quantScore"),
		VerbalScore: str(fx.Profile, "verbalScore"),
		Education:   str(fx.Profile, "education"),
		Experience:  str(fx.Profile, "experience"),
		TestScores:  str(fx.Profile, "testScores"),
	}
	if value, ok := fx.Profile["years"].(float64); ok {
		p.Years = int(value)
	}
	analysis := fit.Analyse(fx.JD, p)
	result := map[string]any{
		"name": fx.Name, "ok": analysis.OK, "score": analysis.Score,
		"band": analysis.Band, "confidence": analysis.Confidence,
		"dimensions": analysis.Dimensions, "requirements": analysis.Requirements,
		"matchedKeywords": analysis.MatchedKeywords,
	}
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		panic(err)
	}
}
