package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shreyaskali/uplanner-ops/internal/config"
	"github.com/shreyaskali/uplanner-ops/internal/core/fit"
	"github.com/shreyaskali/uplanner-ops/internal/profile"
)

func runFit(cfg config.Config, args []string, jsonOut bool) error {
	fs := flag.NewFlagSet("fit", flag.ContinueOnError)
	text := fs.String("text", "", "job description text")
	textFile := fs.String("text-file", "", "read the job description from a file")
	save := fs.Bool("save", false, "write a report to the user layer (gated)")
	confirm := fs.Bool("confirm", false, "confirm a gated write")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}

	source := *text
	if *textFile != "" {
		raw, err := os.ReadFile(*textFile)
		if err != nil {
			return err
		}
		source = string(raw)
	}
	if source == "" {
		raw, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		source = string(raw)
	}
	if strings.TrimSpace(source) == "" {
		return fmt.Errorf("no job description supplied; use --text, --text-file or stdin")
	}

	p, err := profile.Load(cfg.Paths()["profile"])
	if err != nil {
		return err
	}
	analysis := fit.Analyse(source, p)

	if *save && analysis.OK {
		if !*confirm {
			token := fmt.Sprintf("needs_confirmation: fit_write %d characters - rerun with --confirm", len(source))
			if jsonOut {
				return json.NewEncoder(os.Stdout).Encode(map[string]any{
					"needs_confirmation": token, "analysis": analysis,
				})
			}
			fmt.Println(token)
			return nil
		}
		name := time.Now().UTC().Format("20060102-150405") + ".md"
		path := filepath.Join(cfg.DataDir, "data", "fit", name)
		body := renderFitMarkdown(analysis)
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			return err
		}
		if jsonOut {
			return json.NewEncoder(os.Stdout).Encode(map[string]any{"saved": path, "analysis": analysis})
		}
		fmt.Println("Saved", path)
	}

	if jsonOut {
		return json.NewEncoder(os.Stdout).Encode(analysis)
	}
	fmt.Print(renderFitText(analysis))
	return nil
}

func renderFitText(a fit.Analysis) string {
	if !a.OK {
		return a.Summary + "\n"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s — %d/100 · %s (%s confidence)\n\n", a.Headline, a.Score, a.Band, a.Confidence)
	fmt.Fprintln(&b, a.Summary)
	if len(a.Dimensions) > 0 {
		fmt.Fprintln(&b, "\nStrongest areas")
		for _, d := range a.Dimensions {
			fmt.Fprintf(&b, "  %-28s %3d/100  demand %d%%\n", d.Label, d.Fit, d.Demand)
		}
	}
	if len(a.Gaps) > 0 {
		fmt.Fprintln(&b, "\nRequirements to address")
		for _, g := range a.Gaps {
			fmt.Fprintf(&b, "  [%s] %s — %s\n", g.Status, g.Label, g.Detail)
		}
	}
	if len(a.NextMoves) > 0 {
		fmt.Fprintln(&b, "\nNext moves")
		for _, m := range a.NextMoves {
			fmt.Fprintln(&b, "  - "+m)
		}
	}
	fmt.Fprintln(&b, "\n"+a.Caveat)
	return b.String()
}

func renderFitMarkdown(a fit.Analysis) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Fit: %s\n\n", a.Title)
	fmt.Fprintf(&b, "Score: **%d/100 — %s** (%s confidence)\n\n", a.Score, a.Band, a.Confidence)
	fmt.Fprintln(&b, a.Summary)
	fmt.Fprintln(&b, "\n## Dimensions")
	for _, d := range a.Dimensions {
		fmt.Fprintf(&b, "- %s: %d/100 (demand %d%%)\n", d.Label, d.Fit, d.Demand)
	}
	fmt.Fprintln(&b, "\n## Requirements")
	for _, c := range a.Requirements {
		fmt.Fprintf(&b, "- [%s] %s — %s\n", c.Status, c.Label, c.Detail)
	}
	fmt.Fprintln(&b, "\n## Next moves")
	for _, m := range a.NextMoves {
		fmt.Fprintf(&b, "- %s\n", m)
	}
	fmt.Fprintln(&b, "\n"+a.Caveat)
	return b.String()
}
