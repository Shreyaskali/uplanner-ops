// Package tracker reads and writes the canonical application tracker.
// The Markdown table is the source of truth; writes are atomic and gated.
package tracker

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Record is one row in data/applications.md.
type Record struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Deadline  string `json:"deadline"`
	Target    string `json:"target"`
	Documents string `json:"documents"`
}

var validStatus = map[string]bool{
	"not started": true, "preparing": true, "ready": true, "submitted": true,
	"planning": true, "interview": true, "offer": true, "rejected": true,
}

// Read parses the Markdown table. Missing files yield an empty tracker.
func Read(path string) ([]Record, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Record{}, nil
		}
		return nil, err
	}
	defer f.Close()

	records := []Record{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "|") || strings.Contains(line, "---") {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 6 {
			continue
		}
		id := strings.TrimSpace(parts[1])
		if id == "" || id == "id" {
			continue
		}
		records = append(records, Record{
			ID:        id,
			Status:    strings.TrimSpace(parts[2]),
			Deadline:  strings.TrimSpace(parts[3]),
			Target:    strings.TrimSpace(parts[4]),
			Documents: strings.TrimSpace(parts[5]),
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].Deadline != records[j].Deadline {
			if records[i].Deadline == "" {
				return false
			}
			if records[j].Deadline == "" {
				return true
			}
			return records[i].Deadline < records[j].Deadline
		}
		return records[i].ID < records[j].ID
	})
	return records, nil
}

// Write performs an atomic temp-write + rename so a crash cannot corrupt
// the canonical file.
func Write(path string, records []Record) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("# Applications\n\n| id | status | deadline | target | documents |\n|---|---|---|---|---|\n")
	for _, r := range records {
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n", r.ID, r.Status, r.Deadline, r.Target, r.Documents)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(b.String()), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Upsert replaces a record by id or appends it.
func Upsert(records []Record, record Record) []Record {
	out := make([]Record, 0, len(records)+1)
	found := false
	for _, r := range records {
		if r.ID == record.ID {
			out = append(out, record)
			found = true
		} else {
			out = append(out, r)
		}
	}
	if !found {
		out = append(out, record)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// IsValidStatus reports whether a status is in the allowed vocabulary.
func IsValidStatus(status string) bool {
	return validStatus[strings.ToLower(strings.TrimSpace(status))]
}
