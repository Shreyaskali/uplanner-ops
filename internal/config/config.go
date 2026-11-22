// Package config resolves the data root and user-layer paths.
package config

import (
	"os"
	"path/filepath"
)

const appDir = ".uplanner-ops"

// Config holds non-secret runtime settings.
type Config struct {
	DataDir string
	Root    string
}

// Load resolves the data root with the documented precedence.
func Load(root string) Config {
	dir := os.Getenv("UPLANNER_DATA_DIR")
	if dir == "" {
		marker := filepath.Join(root, ".uplanner-data")
		if raw, err := os.ReadFile(marker); err == nil {
			if value := string(raw); len(value) > 0 {
				dir = value
			}
		}
	}
	if dir == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			dir = filepath.Join(home, appDir)
		} else {
			dir = filepath.Join(root, ".data")
		}
	}
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(root, dir)
	}
	return Config{DataDir: dir, Root: root}
}

// Ensure creates the user-layer directories and scaffold files. It never
// overwrites an existing file.
func (c Config) Ensure() error {
	dirs := []string{
		c.DataDir,
		filepath.Join(c.DataDir, "data", "fit"),
		filepath.Join(c.DataDir, "data", "reports"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	scaffolds := map[string]string{
		filepath.Join(c.DataDir, "profile.yml"):             "# Structured profile — edit by hand or with `uplanner profile`\nversion: 1\n",
		filepath.Join(c.DataDir, "evidence.md"):             "# Evidence\n\nOnly source-backed facts about the person belong here.\n",
		filepath.Join(c.DataDir, "resume.md"):               "# Resume\n\nPaste the master CV here.\n",
		filepath.Join(c.DataDir, "data", "applications.md"): "# Applications\n\n| id | status | deadline | target | documents |\n|---|---|---|---|---|\n",
		filepath.Join(c.DataDir, "data", "learning.md"):     "# Learning progress\n\n| resource | state | note |\n|---|---|---|\n",
	}
	for path, content := range scaffolds {
		if _, err := os.Stat(path); err == nil {
			continue
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// Paths returns the canonical user-layer file paths.
func (c Config) Paths() map[string]string {
	return map[string]string{
		"profile":      filepath.Join(c.DataDir, "profile.yml"),
		"resume":       filepath.Join(c.DataDir, "resume.md"),
		"evidence":     filepath.Join(c.DataDir, "evidence.md"),
		"applications": filepath.Join(c.DataDir, "data", "applications.md"),
		"learning":     filepath.Join(c.DataDir, "data", "learning.md"),
	}
}
