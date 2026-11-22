// Command uplanner is the local-first career planning CLI.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/shreyaskali/uplanner-ops/internal/config"
	"github.com/shreyaskali/uplanner-ops/internal/data"
	"github.com/shreyaskali/uplanner-ops/internal/doctor"
)

var version = "1.0.0"

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		os.Exit(2)
	}

	// Split global flags (before the command) from the command and its flags.
	var globals []string
	command := ""
	var subArgs []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if !strings.HasPrefix(a, "-") {
			command = a
			subArgs = args[i+1:]
			break
		}
		globals = append(globals, a)
		if isValueFlag(a) && i+1 < len(args) {
			i++
			globals = append(globals, args[i])
		}
	}
	if command == "" {
		usage()
		os.Exit(2)
	}

	jsonOut := takeJSON(&globals)
	if subJSON := takeJSON(&subArgs); subJSON {
		jsonOut = true
	}
	dataDir := flagValue(globals, "data-dir")
	rootFlag := flagValue(globals, "root")

	root := rootFlag
	if root == "" {
		root = findRoot()
	}
	cfg := config.Load(root)
	if dataDir != "" {
		cfg.DataDir = dataDir
	}

	var err error
	switch command {
	case "doctor":
		err = runDoctor(root, cfg, jsonOut)
	case "version":
		err = runVersion(jsonOut)
	case "profile":
		err = runProfile(cfg, subArgs, jsonOut)
	case "fit":
		err = runFit(cfg, subArgs, jsonOut)
	case "resources":
		err = runResources(root, cfg, subArgs, jsonOut)
	case "learn-paths":
		err = runLearnPaths(root, cfg, subArgs, jsonOut)
	case "plan-learning":
		err = runPlanLearning(root, cfg, subArgs, jsonOut)
	case "progress":
		err = runProgress(cfg, subArgs, jsonOut)
	case "content":
		err = runContent(root, jsonOut)
	case "tui":
		err = runTUI(root, cfg)
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", command)
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func isValueFlag(arg string) bool {
	name := strings.TrimLeft(arg, "-")
	return name == "data-dir" || name == "root"
}

// flagValue reads the value of a global --flag value pair.
func flagValue(args []string, name string) string {
	for i, a := range args {
		if strings.TrimLeft(a, "-") == name && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

// takeJSON removes --json from a slice and reports whether it was present.
func takeJSON(args *[]string) bool {
	found := false
	out := make([]string, 0, len(*args))
	for _, a := range *args {
		if strings.TrimLeft(a, "-") == "json" {
			found = true
			continue
		}
		out = append(out, a)
	}
	*args = out
	return found
}

func usage() {
	fmt.Println(`uplanner-ops — local-first career planning

Usage:
  uplanner doctor [--json]       check the install and data root
  uplanner version [--json]      print version
  uplanner content [--json]      show generated content counts
  uplanner profile [flags]       set structured profile fields
  uplanner fit [flags]           score a job description against your profile
  uplanner resources [flags]     find free learning resources by domain
  uplanner learn-paths [flags]   find learning paths by domain
  uplanner plan-learning [flags] build a free learning plan from a gap
  uplanner progress [flags]      show or mark completed resources
  uplanner tui                   open the Bubble Tea interface
  uplanner help

Profile flags:
  --degree --field --institution --grade --years --role --company
  --skills --exam --quant --verbal --language

Fit flags:
  --text TEXT        job description text
  --text-file PATH   read the job description from a file
  --save             write a report (gated; needs --confirm)
  --confirm          confirm a gated write

Learning flags:
  --domain LIST      comma-separated domains (ai, data, finance, cloud, lead…)
  --gap TEXT         describe a gap for plan-learning
  --limit N          maximum resources
  --done ID          mark a resource done in progress (gated)

Global flags:
  --data-dir DIR   override the user data root
  --root DIR       override the install root
  --json           machine-readable output`)
}

// findRoot walks upward from the executable and the working directory looking
// for the AGENTS.md + go.mod sentinels.
func findRoot() string {
	starts := []string{}
	if exe, err := os.Executable(); err == nil {
		starts = append(starts, filepath.Dir(exe))
	}
	if wd, err := os.Getwd(); err == nil {
		starts = append(starts, wd)
	}
	for _, start := range starts {
		dir := start
		for {
			if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); err == nil {
				if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
					return dir
				}
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return "."
}

func runDoctor(root string, cfg config.Config, jsonOut bool) error {
	report := doctor.Run(root, cfg)
	if jsonOut {
		return json.NewEncoder(os.Stdout).Encode(report)
	}
	fmt.Print(report.Text())
	if !report.OK {
		return fmt.Errorf("doctor found failures; see above")
	}
	return nil
}

func runVersion(jsonOut bool) error {
	if jsonOut {
		return json.NewEncoder(os.Stdout).Encode(map[string]string{"version": version})
	}
	fmt.Println("uplanner-ops", version)
	return nil
}

func runContent(root string, jsonOut bool) error {
	content, err := data.Load(root)
	if err != nil {
		return err
	}
	if jsonOut {
		return json.NewEncoder(os.Stdout).Encode(content.Manifest)
	}
	m := content.Manifest
	fmt.Printf("Generated %s\n", m.GeneratedAt)
	fmt.Printf("  programs:  %d\n  reports:   %d\n  resources: %d (free only)\n  paths:     %d\n  roles:     %d\n",
		m.Programs, m.Reports, m.Resources, m.Paths, m.Roles)
	return nil
}

// runProfile updates the structured profile from flags. It writes only the
// user layer and only fields explicitly supplied.
func runProfile(cfg config.Config, args []string, jsonOut bool) error {
	if err := cfg.Ensure(); err != nil {
		return err
	}
	fs := flag.NewFlagSet("profile", flag.ContinueOnError)
	fields := map[string]*string{}
	for _, name := range []string{"degree", "field", "institution", "grade", "years", "role", "company", "skills", "exam", "quant", "verbal", "language"} {
		fields[name] = fs.String(name, "", "set "+name)
	}
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}

	updates := map[string]string{}
	fs.Visit(func(f *flag.Flag) {
		value := ""
		if ptr := fields[f.Name]; ptr != nil {
			value = *ptr
		}
		updates[f.Name] = value
	})
	if len(updates) == 0 {
		return fmt.Errorf("no fields supplied; see `uplanner help`")
	}

	path := cfg.Paths()["profile"]
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	newContent := upsertYAML(string(raw), updates)
	if err := os.WriteFile(path, []byte(newContent), 0o644); err != nil {
		return err
	}
	if jsonOut {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"written": path, "fields": updates})
	}
	fmt.Println("Updated", path)
	keys := make([]string, 0, len(updates))
	for k := range updates {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("  %-12s %s\n", k, updates[k])
	}
	return nil
}

// upsertYAML replaces or appends top-level scalar keys, preserving comments.
func upsertYAML(current string, updates map[string]string) string {
	lines := strings.Split(strings.TrimRight(current, "\n"), "\n")
	seen := map[string]bool{}
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") || trimmed == "" {
			continue
		}
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		if value, ok := updates[key]; ok {
			lines[i] = key + ": " + quoteYAML(value)
			seen[key] = true
		}
	}
	keys := make([]string, 0, len(updates))
	for k := range updates {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if !seen[k] {
			lines = append(lines, k+": "+quoteYAML(updates[k]))
		}
	}
	return strings.Join(lines, "\n") + "\n"
}

func quoteYAML(value string) string {
	if value == "" {
		return `""`
	}
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return value
	}
	if strings.ContainsAny(value, ":#\n\"'") {
		return strconv.Quote(value)
	}
	return value
}
