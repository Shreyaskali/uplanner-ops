package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shreyaskali/uplanner-ops/internal/config"
	"github.com/shreyaskali/uplanner-ops/internal/data"
)

var (
	accent = lipgloss.Color("#e30613")
	ink    = lipgloss.Color("#111111")
	muted  = lipgloss.Color("#77776f")
	soft   = lipgloss.Color("#f4f4f2")

	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(ink).Background(soft).Padding(0, 1)
	tabStyle   = lipgloss.NewStyle().Foreground(muted).Padding(0, 2)
	activeTab  = lipgloss.NewStyle().Bold(true).Foreground(accent).Underline(true).Padding(0, 2)
	labelStyle = lipgloss.NewStyle().Foreground(muted).Width(22)
	valueStyle = lipgloss.NewStyle().Bold(true).Foreground(ink)
	boxStyle   = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true).BorderForeground(muted).Padding(1, 2)
)

type model struct {
	tabs    []string
	tab     int
	content *data.Content
	cfg     config.Config
	root    string
	status  string
	ready   bool
}

type loadedMsg struct {
	content *data.Content
	err     error
}

func newTUI(root string, cfg config.Config) model {
	return model{
		tabs:   []string{"Today", "Profile", "Learn", "Degree", "Track", "Settings"},
		cfg:    cfg,
		root:   root,
		status: "loading generated content…",
	}
}

func (m model) Init() tea.Cmd {
	return func() tea.Msg {
		content, err := data.Load(m.root)
		return loadedMsg{content: content, err: err}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case loadedMsg:
		if msg.err != nil {
			m.status = "content missing — run `go run ./cmd/gen-data`"
		} else {
			m.content = msg.content
			m.status = "ready · " + m.content.Manifest.GeneratedAt
			m.ready = true
		}
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c", "esc"))):
			return m, tea.Quit
		case key.Matches(msg, key.NewBinding(key.WithKeys("left", "h"))):
			if m.tab > 0 {
				m.tab--
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("right", "l"))):
			if m.tab < len(m.tabs)-1 {
				m.tab++
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("1", "2", "3", "4", "5", "6"))):
			idx := int(msg.String()[0] - '1')
			if idx >= 0 && idx < len(m.tabs) {
				m.tab = idx
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(" uplanner-ops "+version+" ") + "  " + mutedStyle().Render(m.status) + "\n\n")

	for i, tab := range m.tabs {
		style := tabStyle
		if i == m.tab {
			style = activeTab
		}
		b.WriteString(style.Render(fmt.Sprintf("%d %s", i+1, tab)))
	}
	b.WriteString("\n\n")

	body := m.tabBody()
	b.WriteString(boxStyle.Width(84).Render(body))
	b.WriteString("\n\n" + mutedStyle().Render("←/→ or 1-6: switch · q: quit"))
	return b.String()
}

func mutedStyle() lipgloss.Style { return lipgloss.NewStyle().Foreground(muted) }

func (m model) tabBody() string {
	if !m.ready {
		return "Loading…\n\nIf this persists, run:\n  go run ./cmd/gen-data --uplanner DIR --intersections DIR"
	}
	switch m.tabs[m.tab] {
	case "Today":
		lines := []string{
			row("Programs researched", fmt.Sprintf("%d", len(m.content.Programs))),
			row("Hiring reports", fmt.Sprintf("%d", len(m.content.Reports))),
			row("Free resources", fmt.Sprintf("%d", len(m.content.Resources))),
			row("Learning paths", fmt.Sprintf("%d", len(m.content.Paths))),
			row("Role profiles", fmt.Sprintf("%d", len(m.content.Roles))),
			"",
			mutedStyle().Render("Next: set your profile with `uplanner profile`, then paste a JD for a fit score."),
		}
		return strings.Join(lines, "\n")
	case "Profile":
		return strings.Join([]string{
			row("Data root", m.cfg.DataDir),
			row("Profile file", m.cfg.Paths()["profile"]),
			"",
			mutedStyle().Render("Structured fields: degree, field, institution, grade, years, role,"),
			mutedStyle().Render("company, skills, exam, quant, verbal, language."),
			mutedStyle().Render("Use `uplanner profile --field '...' --years 3` to set them."),
		}, "\n")
	case "Learn":
		free := 0
		for _, path := range m.content.Paths {
			if path.Tier == "free" {
				free++
			}
		}
		return strings.Join([]string{
			row("Free paths", fmt.Sprintf("%d", free)),
			row("Free resources", fmt.Sprintf("%d", len(m.content.Resources))),
			"",
			mutedStyle().Render("Resources are filtered to free-only at generation time."),
			mutedStyle().Render("Learning modes arrive in Phase 3."),
		}, "\n")
	case "Degree":
		return strings.Join([]string{
			row("Directory", fmt.Sprintf("%d programs", len(m.content.Programs))),
			row("Evidence", fmt.Sprintf("%d reports", len(m.content.Reports))),
			"",
			mutedStyle().Render("Degree planning, comparison and cost modes arrive in Phase 4."),
		}, "\n")
	case "Track":
		return strings.Join([]string{
			row("Tracker", m.cfg.Paths()["applications"]),
			"",
			mutedStyle().Render("Status, deadlines and document checklists arrive in Phase 4."),
		}, "\n")
	case "Settings":
		return strings.Join([]string{
			row("Version", version),
			row("Data root", m.cfg.DataDir),
			row("LLM key", func() string {
				if os.Getenv("UPLANNER_LLM_KEY") != "" {
					return "configured (hidden)"
				}
				return "not set — offline modes work"
			}()),
			"",
			mutedStyle().Render("Run `uplanner doctor` for the full setup report."),
		}, "\n")
	}
	return ""
}

func row(label, value string) string {
	return labelStyle.Render(label) + valueStyle.Render(value)
}

func runTUI(root string, cfg config.Config) error {
	if _, err := tea.NewProgram(newTUI(root, cfg)).Run(); err != nil {
		return err
	}
	return nil
}
