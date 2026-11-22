// Package profile reads the structured user-layer profile.
// The format is intentionally flat scalars so it stays human-editable.
package profile

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// Profile is the source-backed structured profile used by every deterministic
// mode. Fields use the same names as the web configurator.
type Profile struct {
	Version     string `json:"version"`
	DegreeLevel string `json:"degreeLevel"`
	Field       string `json:"field"`
	Institution string `json:"institution"`
	Grade       string `json:"grade"`
	Years       int    `json:"years"`
	Role        string `json:"role"`
	Company     string `json:"company"`
	Skills      string `json:"skills"`
	Exam        string `json:"exam"`
	QuantScore  string `json:"quantScore"`
	VerbalScore string `json:"verbalScore"`
	Language    string `json:"language"`
	Education   string `json:"education"`
	Experience  string `json:"experience"`
	TestScores  string `json:"testScores"`
	Complete    bool   `json:"complete"`
}

// Load reads a flat key: value YAML file. Missing files yield an empty profile.
func Load(path string) (Profile, error) {
	var p Profile
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return p, nil
		}
		return p, err
	}
	defer f.Close()

	values := map[string]string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := unquote(strings.TrimSpace(parts[1]))
		values[key] = value
	}
	if err := scanner.Err(); err != nil {
		return p, err
	}

	p.Version = values["version"]
	p.DegreeLevel = values["degreeLevel"]
	p.Field = values["field"]
	p.Institution = values["institution"]
	p.Grade = values["grade"]
	p.Role = values["role"]
	p.Company = values["company"]
	p.Skills = values["skills"]
	p.Exam = values["exam"]
	p.QuantScore = values["quantScore"]
	p.VerbalScore = values["verbalScore"]
	p.Language = values["language"]
	p.Education = values["education"]
	p.Experience = values["experience"]
	p.TestScores = values["testScores"]
	if n, err := strconv.Atoi(values["years"]); err == nil && n >= 0 && n <= 50 {
		p.Years = n
	} else {
		p.Years = 0
	}
	// "complete" means the user actually recorded evidence, not that a field
	// count happens to look full.
	p.Complete = p.Field != "" || p.Role != "" || p.Skills != "" || p.Education != "" || p.Experience != ""
	return p, nil
}

func unquote(value string) string {
	if len(value) >= 2 {
		if value[0] == '"' && value[len(value)-1] == '"' {
			if unquoted, err := strconv.Unquote(value); err == nil {
				return unquoted
			}
		}
		if value[0] == '\'' && value[len(value)-1] == '\'' {
			return strings.ReplaceAll(value[1:len(value)-1], `\'`, `'`)
		}
	}
	return value
}
