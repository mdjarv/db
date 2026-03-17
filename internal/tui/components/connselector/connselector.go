// Package connselector provides a modal connection picker overlay.
package connselector

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/conn"
	"github.com/mdjarv/db/internal/tui/theme"
)

// SelectMsg signals a connection was chosen.
type SelectMsg struct {
	Candidate conn.Candidate
}

// CancelMsg signals the selector was dismissed (esc).
type CancelMsg struct{}

// QuitMsg signals the user wants to quit the app (q).
type QuitMsg struct{}

// AddMsg signals the user wants to add a new connection.
type AddMsg struct {
	Source conn.Source
}

// EditMsg signals the user wants to edit a connection.
type EditMsg struct {
	Candidate conn.Candidate
}

// DeleteMsg signals the user wants to delete a connection.
type DeleteMsg struct {
	Candidate conn.Candidate
}

// Model is the connection selector state.
type Model struct {
	active     bool
	candidates []conn.Candidate
	cursor     int
	width      int
	height     int
}

// New creates an inactive selector.
func New() *Model {
	return &Model{}
}

// IsActive returns whether the selector is visible.
func (m *Model) IsActive() bool { return m.active }

// Open shows the selector with the given candidates.
func (m *Model) Open(candidates []conn.Candidate) {
	m.active = true
	m.candidates = candidates
	m.cursor = 0
}

// Close dismisses the selector.
func (m *Model) Close() {
	m.active = false
}

// SetSize updates the available render area.
func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// Update handles key input.
func (m *Model) Update(msg tea.KeyMsg) tea.Cmd {
	if !m.active || len(m.candidates) == 0 {
		return nil
	}
	switch msg.String() {
	case "j", "down":
		if m.cursor < len(m.candidates)-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "g":
		m.cursor = 0
	case "G":
		m.cursor = len(m.candidates) - 1
	case "enter":
		c := m.candidates[m.cursor]
		return func() tea.Msg { return SelectMsg{Candidate: c} }
	case "a":
		source := conn.SourceProjectStore
		if m.cursor < len(m.candidates) {
			s := m.candidates[m.cursor].Source
			if s == conn.SourceProjectStore || s == conn.SourceGlobalStore {
				source = s
			}
		}
		return func() tea.Msg { return AddMsg{Source: source} }
	case "e":
		c := m.candidates[m.cursor]
		if !isEditable(c) {
			return nil
		}
		return func() tea.Msg { return EditMsg{Candidate: c} }
	case "d":
		c := m.candidates[m.cursor]
		if !isEditable(c) {
			return nil
		}
		return func() tea.Msg { return DeleteMsg{Candidate: c} }
	case "q":
		m.Close()
		return func() tea.Msg { return QuitMsg{} }
	case "esc":
		m.Close()
		return func() tea.Msg { return CancelMsg{} }
	}
	return nil
}

var sourceOrder = []conn.Source{
	conn.SourceProjectStore,
	conn.SourceGlobalStore,
	conn.SourceEnvVar,
	conn.SourceDotEnv,
}

var sourceHeaders = map[conn.Source]string{
	conn.SourceProjectStore: "Saved (project)",
	conn.SourceGlobalStore:  "Saved (global)",
	conn.SourceEnvVar:       "Environment",
	conn.SourceDotEnv:       ".env",
}

// View renders the selector overlay.
func (m *Model) View(containerW, containerH int) string {
	if !m.active {
		return ""
	}

	t := theme.Current()
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(t.Styles.BorderFocused)
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Bold(true)
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle := lipgloss.NewStyle().Bold(true).Reverse(true)
	normalStyle := lipgloss.NewStyle()
	defaultMark := lipgloss.NewStyle().Foreground(lipgloss.Color("220"))

	w := min(containerW-4, 60)
	if w < 30 {
		w = 30
	}

	// Group candidates by source, sorted alphabetically within each group.
	type group struct {
		source     conn.Source
		candidates []int // indices into m.candidates
	}
	groupMap := make(map[conn.Source][]int)
	for i, c := range m.candidates {
		groupMap[c.Source] = append(groupMap[c.Source], i)
	}
	var groups []group
	for _, src := range sourceOrder {
		if idxs, ok := groupMap[src]; ok {
			sort.Slice(idxs, func(a, b int) bool {
				return m.candidates[idxs[a]].Config.Name < m.candidates[idxs[b]].Config.Name
			})
			groups = append(groups, group{source: src, candidates: idxs})
		}
	}

	var lines []string
	lines = append(lines, titleStyle.Render("Select Connection"))
	lines = append(lines, "")

	for gi, g := range groups {
		if header, ok := sourceHeaders[g.source]; ok {
			lines = append(lines, headerStyle.Render(header))
		}
		for _, idx := range g.candidates {
			c := m.candidates[idx]
			label := formatCandidate(c)
			if c.IsDefault {
				label = defaultMark.Render("* ") + label
			} else {
				label = "  " + label
			}
			if idx == m.cursor {
				label = cursorStyle.Render(label)
			} else {
				label = normalStyle.Render(label)
			}
			lines = append(lines, label)
		}
		if gi < len(groups)-1 {
			lines = append(lines, "")
		}
	}

	if len(m.candidates) == 0 {
		lines = append(lines, hintStyle.Render("  No connections found"))
	}

	lines = append(lines, "")
	lines = append(lines, hintStyle.Render("j/k navigate  Enter select  a add  e edit  d delete"))
	lines = append(lines, hintStyle.Render("q quit  Esc cancel"))

	// Trim to fit.
	maxLines := containerH - 6
	if maxLines < 5 {
		maxLines = 5
	}
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}

	content := strings.Join(lines, "\n")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Styles.BorderFocused).
		Padding(1, 2).
		Width(w).
		Render(content)

	return lipgloss.Place(containerW, containerH, lipgloss.Center, lipgloss.Center, box)
}

// Refresh updates the candidate list and restores cursor to the named connection.
func (m *Model) Refresh(candidates []conn.Candidate, restoreName string) {
	m.candidates = candidates
	m.cursor = 0
	for i, c := range candidates {
		if c.Config.Name == restoreName {
			m.cursor = i
			return
		}
	}
	if m.cursor >= len(candidates) && len(candidates) > 0 {
		m.cursor = len(candidates) - 1
	}
}

func isEditable(c conn.Candidate) bool {
	return c.Source == conn.SourceProjectStore || c.Source == conn.SourceGlobalStore
}

func formatCandidate(c conn.Candidate) string {
	cfg := c.Config
	name := cfg.Name
	if name == "" {
		name = c.Label
	}
	host := cfg.Host
	if cfg.Port != 0 && cfg.Port != 5432 {
		host = fmt.Sprintf("%s:%d", host, cfg.Port)
	}
	detail := fmt.Sprintf("%s@%s/%s", cfg.User, host, cfg.DBName)
	return fmt.Sprintf("%s — %s", name, detail)
}
