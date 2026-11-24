package fit

import (
	"regexp"

	"github.com/shreyaskali/uplanner-ops/internal/profile"
)

func check(label, status, detail string, weight int) Check {
	return Check{Label: label, Status: status, Detail: detail, Weight: weight}
}

// requirementChecks reads explicit requirements from a job description and
// compares them with the user's recorded profile. Word boundaries matter:
// "painting" must not satisfy an "ai" check.
func requirementChecks(lower string, p profile.Profile) []Check {
	checks := []Check{}

	if m := regexp.MustCompile(`\b(?:at least\s*)?(\d{1,2})(?:\s*(?:-|to)\s*(\d{1,2}))?\s*(?:\+|plus)?\s*(?:years?|yrs?)\b`).FindStringSubmatch(lower); len(m) > 0 {
		min := atoi(m[1])
		years := p.Years
		if p.Years == 0 {
			years = 2 // neutral default when the user has not recorded years
		}
		// JS renders the range as "1–3" or the floor as "15+".
		requested := m[1]
		if len(m) > 2 && m[2] != "" {
			requested = m[1] + "–" + m[2]
		} else {
			requested = m[1] + "+"
		}
		status, weight, detail := "met", 1, ""
		switch {
		case min <= years:
			detail = requested + " years requested; your experience covers the early-career band (about " + itoa(years) + " years)."
		case min <= years+2:
			status, weight = "partial", -2
			detail = requested + " years requested; frame your experience as equivalent scope, but expect a screen on tenure."
		default:
			status, weight = "gap", -7
			detail = requested + " years requested; your profile has about " + itoa(years) + " years. This is a seniority gap, not a wording problem."
		}
		checks = append(checks, check("Experience", status, detail, weight))
	}

	if re(`\bph\.?d\b|\bdoctorate\b`).MatchString(lower) {
		checks = append(checks, check("Doctorate", "gap", "A doctorate is requested and is not recorded in your profile.", -9))
	} else if re(`\bmba\b|\bmaster'?s?\b|\bmasters\b|\bmsc\b|\bpgdm?\b`).MatchString(lower) {
		checks = append(checks, check("Postgraduate degree", "partial", "A master’s or MBA is requested. Your 2027 application is the route to meet it, not a credential you hold today.", -3))
	} else if re(`\bbachelor'?s?\b|\bundergraduate degree\b|\bdegree\b`).MatchString(lower) {
		checks = append(checks, check("Degree", "met", "Your degree meets the stated undergraduate requirement.", 1))
	}

	if re(`\bsql\b`).MatchString(lower) {
		strong := hasField(p, "computer", "data", "statistics", "analytics", "information technology", "software", "engineering")
		if strong {
			checks = append(checks, check("SQL", "met", "Your field of study covers database work; be ready to show a query or dashboard example.", 1))
		} else {
			checks = append(checks, check("SQL", "partial", "SQL is not shown by your profile yet. Bring one working example to the conversation.", -2))
		}
	}
	if re(`\bpython\b`).MatchString(lower) {
		strong := hasField(p, "computer", "data", "statistics", "analytics", "artificial intelligence", "machine learning")
		if strong {
			checks = append(checks, check("Python", "partial", "Your field is relevant, but the profile does not show a production Python project. Bring one.", -2))
		} else {
			checks = append(checks, check("Python", "gap", "Python is not evidenced by your profile. Treat this as a real gap.", -7))
		}
	}
	if re(`\bexcel\b`).MatchString(lower) {
		checks = append(checks, check("Excel", "met", "Commercial and analytical work is credible Excel evidence.", 1))
	}
	if re(`\b(aws|azure|gcp|google cloud|cloud platform|cloud platforms)\b`).MatchString(lower) {
		strong := hasField(p, "cloud", "platform", "infrastructure", "devops", "software", "engineering") ||
			reWord("analyst").MatchString(norm(p.Role)) || reWord("engineer").MatchString(norm(p.Role))
		if strong {
			checks = append(checks, check("Cloud platforms", "met", "Your field or role covers cloud work directly.", 1))
		} else {
			checks = append(checks, check("Cloud platforms", "partial", "Cloud work is not shown by your profile yet. Name the closest platform you have touched.", -2))
		}
	}
	if re(`\b(tableau|power bi|looker)\b`).MatchString(lower) {
		checks = append(checks, check("BI tooling", "partial", "The specific BI tool is not recorded. Name the closest tool you have used.", -2))
	}
	if re(`\b(machine learning|artificial intelligence|generative ai|llm|mlops)\b`).MatchString(lower) {
		strong := hasField(p, "artificial intelligence", "machine learning", "data", "statistics", "computer")
		if strong {
			checks = append(checks, check("AI / ML", "partial", "Your background is adjacent to AI work, but model building is not yet evidenced. Distinguish coursework from deployment.", -3))
		} else {
			checks = append(checks, check("AI / ML", "gap", "AI is not evidenced by your profile. This is a genuine gap for this posting.", -7))
		}
	}

	if re(`\bcfa\b`).MatchString(lower) {
		preferred := re(`\bpreferred\b|\badvantage\b|\ba plus\b`).MatchString(lower)
		if preferred {
			checks = append(checks, check("CFA", "partial", "CFA is preferred rather than mandatory; note progress if started.", -2))
		} else {
			checks = append(checks, check("CFA", "gap", "CFA is requested and the profile does not show it. This is a hard screen for many finance roles.", -8))
		}
	}
	if re(`\b(cpa|acca|chartered accountant)\b`).MatchString(lower) {
		checks = append(checks, check("Accounting qualification", "gap", "A CPA/ACCA/CA-style qualification is requested and is not recorded.", -9))
	}
	if re(`\bfrm\b`).MatchString(lower) {
		checks = append(checks, check("FRM", "partial", "FRM is not recorded. If the role is risk-focused, treat it as a study commitment.", -2))
	}
	if re(`\bpmp\b`).MatchString(lower) {
		checks = append(checks, check("PMP", "partial", "PMP is not recorded; delivery evidence may substitute.", -2))
	}
	if re(`\b(six sigma|lean)\b`).MatchString(lower) {
		checks = append(checks, check("Lean / Six Sigma", "partial", "Engineering training gives the vocabulary; a named project would make it concrete.", -1))
	}

	if re(`\b(security clearance|u\.?s\.? citizen|permanent resident|green card|authori[sz]ed to work|right to work|work permit)\b`).MatchString(lower) {
		domestic := re(`\b(india|bengaluru|bangalore|pune|mumbai|hyderabad|gurugram|noida|chennai|delhi|remote)\b`).MatchString(lower)
		if domestic {
			checks = append(checks, check("Work authorisation", "note", "The posting names a domestic or remote setup. Confirm the exact right-to-work wording.", 0))
		} else {
			checks = append(checks, check("Work authorisation", "gap", "The posting asks for existing work authorisation. From India you would likely need sponsorship or a local permit.", -8))
		}
	}

	if re(`\b(director|vice president|vp|head of|principal|staff engineer|senior manager|general manager)\b`).MatchString(lower) {
		checks = append(checks, check("Seniority", "gap", "The title uses “director” language, which is above the early-career band this profile currently supports.", -6))
	} else if re(`\bsenior\b`).MatchString(lower) && re(`\b(\d{1,2})\s*(?:\+|plus)?\s*(?:years?|yrs?)\b`).MatchString(lower) {
		// JS adds a middle branch: "senior" plus 4+ years is partial, but a
		// lower number (e.g. "senior stakeholders" in a 1-3 year role) adds
		// nothing at all — it must not fall through to the early-career "met".
		if m := re(`\b(\d{1,2})\s*(?:\+|plus)?\s*(?:years?|yrs?)\b`).FindStringSubmatch(lower); len(m) > 1 && atoi(m[1]) >= 4 {
			checks = append(checks, check("Seniority", "partial", "The role asks for a senior level above a first post-degree move.", -3))
		}
	} else if re(`\b(entry level|entry-level|graduate|junior|associate|analyst)\b`).MatchString(lower) {
		checks = append(checks, check("Seniority", "met", "The language is early-career or analyst-level, which fits the profile's stage.", 0))
	}

	for _, item := range []struct{ term, label string }{
		{"registered nurse", "Clinical licence"}, {"clinical licence", "Clinical licence"},
		{"attorney", "Legal qualification"}, {"commercial pilot", "Aviation licence"},
		{"physician", "Medical licence"}, {"pharmacist", "Pharmacy licence"},
		{"licensed therapist", "Clinical licence"},
	} {
		if reWord(item.term).MatchString(lower) {
			checks = append(checks, check(item.label, "gap", "The description asks for a licence or qualification the profile does not hold. This is a different career track.", -12))
			break
		}
	}
	if re(`\b(quota|cold calling|commission-only|commission only)\b`).MatchString(lower) {
		checks = append(checks, check("Quota-carrying sales", "gap", "The role carries a sales quota or cold-calling expectation the profile is not built around.", -8))
	}
	return checks
}

var reCache = map[string]*regexp.Regexp{}

func re(pattern string) *regexp.Regexp {
	if cached, ok := reCache[pattern]; ok {
		return cached
	}
	compiled := regexp.MustCompile(pattern)
	reCache[pattern] = compiled
	return compiled
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func atoiOr(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
