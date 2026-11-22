package main

import (
	"strings"
	"testing"

	"github.com/shreyaskali/uplanner-ops/internal/config"
	"github.com/shreyaskali/uplanner-ops/internal/data"
)

func testModel() model {
	m := newTUI("/root", config.Config{DataDir: "/data", Root: "/root"})
	m.ready = true
	m.content = &data.Content{
		Manifest:  data.Manifest{Programs: 179, Reports: 63, Resources: 258, Paths: 31, Roles: 37},
		Programs:  make([]data.Program, 179),
		Reports:   make([]data.Report, 63),
		Resources: make([]data.Resource, 258),
		Paths:     make([]data.LearningPath, 31),
		Roles:     make([]data.Role, 37),
	}
	return m
}

func TestTUIViewRendersHeaderAndTabs(t *testing.T) {
	m := testModel()
	view := m.View()
	for _, want := range []string{"uplanner-ops", version, "Today", "Profile", "Learn", "Degree", "Track", "Settings"} {
		if !strings.Contains(view, want) {
			t.Errorf("view missing %q", want)
		}
	}
}

func TestTUIViewTodayShowsCounts(t *testing.T) {
	m := testModel()
	view := m.View()
	for _, want := range []string{"179", "63", "258", "31", "37"} {
		if !strings.Contains(view, want) {
			t.Errorf("today view missing count %q", want)
		}
	}
}

func TestTUIViewAllTabs(t *testing.T) {
	m := testModel()
	for i := range m.tabs {
		m.tab = i
		view := m.View()
		if !strings.Contains(view, m.tabs[i]) {
			t.Errorf("tab %d view missing its name", i)
		}
		if strings.Contains(view, "Loading") {
			t.Errorf("tab %d still shows loading", i)
		}
	}
}

func TestTUINotReadyShowsGuidance(t *testing.T) {
	m := newTUI("/root", config.Config{DataDir: "/data", Root: "/root"})
	view := m.View()
	if !strings.Contains(view, "go run ./cmd/gen-data") {
		t.Fatal("loading view missing generator guidance")
	}
}
