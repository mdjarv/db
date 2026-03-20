// Package helpoverlay provides a modal help overlay showing context-sensitive keybindings.
package helpoverlay

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/tui/core"
	"github.com/mdjarv/db/internal/tui/pane"
	"github.com/mdjarv/db/internal/tui/theme"
)

// Section groups related keybindings under a heading.
type Section struct {
	Title    string
	Bindings []Binding
}

// Binding is a key-description pair.
type Binding struct {
	Key  string
	Desc string
}

// Topic is a named help topic shown via :help <topic>.
type Topic struct {
	Name     string
	Title    string
	Sections []Section
}

// Model holds the help overlay state.
type Model struct {
	active bool
	scroll int
	topic  string // "" = context-sensitive, else topic name
}

// New creates an inactive help overlay.
func New() *Model {
	return &Model{}
}

// IsActive returns whether the overlay is visible.
func (m *Model) IsActive() bool { return m.active }

// Open shows the context-sensitive help overlay.
func (m *Model) Open() {
	m.active = true
	m.scroll = 0
	m.topic = ""
}

// OpenTopic shows help for a specific topic.
func (m *Model) OpenTopic(topic string) {
	m.active = true
	m.scroll = 0
	m.topic = topic
}

// Close dismisses the overlay.
func (m *Model) Close() {
	m.active = false
	m.scroll = 0
	m.topic = ""
}

// Update handles key input when the overlay is active.
func (m *Model) Update(msg tea.KeyMsg) tea.Cmd {
	if !m.active {
		return nil
	}
	switch msg.String() {
	case "j", "down":
		m.scroll++
	case "k", "up":
		if m.scroll > 0 {
			m.scroll--
		}
	case "g":
		m.scroll = 0
	case "G":
		m.scroll = 9999
	case "ctrl+d":
		m.scroll += 10
	case "ctrl+u":
		m.scroll -= 10
		if m.scroll < 0 {
			m.scroll = 0
		}
	default:
		m.Close()
	}
	return nil
}

// View renders the help overlay centered in the given dimensions.
func (m *Model) View(activePane pane.ID, mode core.Mode, containerW, containerH int) string {
	if !m.active {
		return ""
	}

	t := theme.Current()

	var sections []Section
	if m.topic != "" {
		sections = topicSections(m.topic)
	} else {
		sections = contextSections(activePane, mode)
	}

	// render sections into styled lines
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(t.Styles.BorderFocused)
	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252"))
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("215"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	var lines []string

	// header
	header := "Keybindings"
	if m.topic != "" {
		if tp := findTopic(m.topic); tp != nil {
			header = tp.Title
		} else {
			header = "Help: " + m.topic
		}
	}
	lines = append(lines, titleStyle.Render(header))
	lines = append(lines, "")

	maxKeyW := 0
	for _, sec := range sections {
		for _, b := range sec.Bindings {
			if len(b.Key) > maxKeyW {
				maxKeyW = len(b.Key)
			}
		}
	}

	for i, sec := range sections {
		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, sectionStyle.Render(sec.Title))
		for _, b := range sec.Bindings {
			pad := strings.Repeat(" ", maxKeyW-len(b.Key)+2)
			lines = append(lines, "  "+keyStyle.Render(b.Key)+pad+descStyle.Render(b.Desc))
		}
	}

	// viewport
	maxLines := containerH - 6
	if maxLines < 5 {
		maxLines = 5
	}

	totalLines := len(lines)
	maxScroll := totalLines - maxLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.scroll > maxScroll {
		m.scroll = maxScroll
	}

	start := m.scroll
	end := start + maxLines
	if end > totalLines {
		end = totalLines
	}
	visible := lines[start:end]

	// scroll hint
	if totalLines > maxLines {
		visible = append(visible, "")
		visible = append(visible, hintStyle.Render(
			fmt.Sprintf("j/k scroll  g/G top/bottom  (%d/%d)", start+1, totalLines)))
	} else {
		visible = append(visible, "")
		visible = append(visible, hintStyle.Render("? / Esc / q to dismiss"))
	}

	boxW := min(containerW-4, 58)
	if boxW < 30 {
		boxW = 30
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Styles.BorderFocused).
		Padding(1, 2).
		Width(boxW).
		Render(strings.Join(visible, "\n"))

	return lipgloss.Place(containerW, containerH, lipgloss.Center, lipgloss.Center, box)
}

// TopicNames returns all valid help topic names.
func TopicNames() []string {
	names := make([]string, len(topics))
	for i, t := range topics {
		names[i] = t.Name
	}
	return names
}

// contextSections returns keybinding sections for the given pane and mode.
func contextSections(activePane pane.ID, mode core.Mode) []Section {
	var sections []Section

	// global navigation
	sections = append(sections, Section{
		Title: "Navigation",
		Bindings: []Binding{
			{"ctrl+h/j/k/l", "focus left/down/up/right"},
			{"tab / shift+tab", "cycle panes"},
			{"1 / 2 / 3", "jump to pane"},
			{"+  /  -", "grow/shrink pane"},
		},
	})

	sections = append(sections, Section{
		Title: "Modes",
		Bindings: []Binding{
			{"i", "insert mode"},
			{"Esc", "normal mode"},
			{":", "command mode"},
			{"?", "toggle help"},
		},
	})

	sections = append(sections, Section{
		Title: "Connection",
		Bindings: []Binding{
			{"ctrl+o", "switch connection"},
			{":connect", "switch connection"},
		},
	})

	sections = append(sections, Section{
		Title: "Buffers",
		Bindings: []Binding{
			{"gt / gT", "next/prev buffer"},
			{":new", "new buffer"},
			{":bd", "close buffer"},
			{":bn / :bp", "next/prev buffer"},
			{":b N", "switch to buffer N"},
			{":ls", "list buffers"},
		},
	})

	sections = append(sections, Section{
		Title: "Commands",
		Bindings: []Binding{
			{":q", "quit"},
			{":help [topic]", "show help (:help commands for full list)"},
			{":run", "execute query"},
			{":clear", "clear query buffer"},
			{":set <opt>", "toggle settings"},
			{":commit / :w", "apply pending changes"},
			{":rollback", "discard pending changes"},
			{":changes", "list pending changes"},
			{":export <fmt> <file>", "export results (csv/json/sql)"},
			{":connect", "open connection selector"},
			{":refresh", "reload schema"},
			{":dump [table]", "dump table/database"},
			{":theme [name]", "list/switch themes"},
		},
	})

	// pane-specific
	switch activePane {
	case pane.TableList:
		sections = append(sections, Section{
			Title: "Table List",
			Bindings: []Binding{
				{"j / k", "navigate tables"},
				{"gg / G", "top/bottom"},
				{"/", "filter tables"},
				{"Enter", "query selected table"},
				{"d", "describe table"},
				{"y", "yank table name"},
				{"R", "refresh schema"},
			},
		})

	case pane.QueryEditor:
		sections = append(sections, Section{
			Title: "Query Editor (Normal)",
			Bindings: []Binding{
				{"j/k/h/l", "navigate"},
				{"0 / $", "line start/end"},
				{"w / b", "word forward/back"},
				{"gg / G", "top/bottom"},
				{"dd", "delete line"},
				{"D", "delete to end"},
				{"u", "undo"},
				{"x", "delete char"},
				{"y", "yank, p/P paste"},
				{"v / V", "visual mode"},
				{"Enter", "run query"},
			},
		})
		if mode == core.ModeInsert {
			sections = append(sections, Section{
				Title: "Query Editor (Insert)",
				Bindings: []Binding{
					{"Esc", "back to normal"},
					{"any key", "type text"},
				},
			})
		} else {
			sections = append(sections, Section{
				Title: "Insert Variants",
				Bindings: []Binding{
					{"i", "insert at cursor"},
					{"a / A", "append / append at EOL"},
					{"I", "insert at line start"},
					{"o / O", "open line below/above"},
				},
			})
		}

	case pane.ResultView:
		sections = append(sections, Section{
			Title: "Results",
			Bindings: []Binding{
				{"j/k/h/l", "navigate"},
				{"g / G", "top/bottom"},
				{"0 / $", "first/last column"},
				{"ctrl+d / ctrl+u", "half page down/up"},
				{"ctrl+f / ctrl+b", "full page down/up"},
				{"e", "edit cell"},
				{"y / Y", "yank cell/row"},
			},
		})
		if mode.IsVisual() {
			sections = append(sections, Section{
				Title: "Visual Mode",
				Bindings: []Binding{
					{"V / v", "line/block select"},
					{"j/k/h/l", "extend selection"},
					{"tab", "toggle axis (V mode)"},
					{"y", "yank selection"},
					{"Esc", "cancel"},
				},
			})
		} else {
			sections = append(sections, Section{
				Title: "Visual & Editing",
				Bindings: []Binding{
					{"V / v", "line/block select"},
					{"dR", "delete row"},
					{"oR", "insert row"},
					{"u", "undo last change"},
					{"ctrl+s", "commit changes"},
					{":commit", "apply changes"},
					{":rollback", "discard changes"},
				},
			})
		}
	}

	return sections
}

var topics = []Topic{
	{
		Name:  "navigation",
		Title: "Navigation",
		Sections: []Section{
			{
				Title: "Pane Focus",
				Bindings: []Binding{
					{"ctrl+h", "focus left pane"},
					{"ctrl+j", "focus down pane"},
					{"ctrl+k", "focus up pane"},
					{"ctrl+l", "focus right pane"},
					{"tab", "next pane"},
					{"shift+tab", "prev pane"},
					{"1", "table list pane"},
					{"2", "query editor pane"},
					{"3", "result view pane"},
				},
			},
			{
				Title: "Pane Resize",
				Bindings: []Binding{
					{"+", "grow left pane"},
					{"-", "shrink left pane"},
				},
			},
		},
	},
	{
		Name:  "modes",
		Title: "Vim Modes",
		Sections: []Section{
			{
				Title: "Mode Switching",
				Bindings: []Binding{
					{"i", "enter insert mode"},
					{"Esc", "return to normal mode"},
					{":", "enter command mode"},
					{"v", "visual block mode (results)"},
					{"V", "visual line mode (results)"},
				},
			},
			{
				Title: "Mode Descriptions",
				Bindings: []Binding{
					{"NORMAL", "navigation and actions"},
					{"INSERT", "text input in query editor"},
					{"COMMAND", "ex-style commands (:q, :run, etc.)"},
					{"V-LINE", "line-wise selection in results"},
					{"V-BLOCK", "block-wise selection in results"},
					{"EDIT", "inline cell editing"},
				},
			},
		},
	},
	{
		Name:  "tables",
		Title: "Table List",
		Sections: []Section{
			{
				Title: "Navigation",
				Bindings: []Binding{
					{"j / k", "move down/up"},
					{"gg", "jump to top"},
					{"G", "jump to bottom"},
					{"/", "filter/search tables"},
				},
			},
			{
				Title: "Actions",
				Bindings: []Binding{
					{"Enter", "query selected table (SELECT * ... LIMIT 100)"},
					{"d", "describe table schema"},
					{"y", "yank table name to clipboard"},
					{"R", "refresh schema from database"},
				},
			},
		},
	},
	{
		Name:  "editor",
		Title: "Query Editor",
		Sections: []Section{
			{
				Title: "Normal Mode",
				Bindings: []Binding{
					{"h/j/k/l", "move left/down/up/right"},
					{"0", "beginning of line"},
					{"$", "end of line"},
					{"w", "next word"},
					{"b", "previous word"},
					{"gg", "top of buffer"},
					{"G", "bottom of buffer"},
					{"dd", "delete line"},
					{"D", "delete to end of line"},
					{"x", "delete character"},
					{"u", "undo"},
					{"y", "yank line"},
					{"p / P", "paste after/before"},
				},
			},
			{
				Title: "Entering Insert Mode",
				Bindings: []Binding{
					{"i", "insert before cursor"},
					{"a", "append after cursor"},
					{"A", "append at end of line"},
					{"I", "insert at beginning of line"},
					{"o", "open line below"},
					{"O", "open line above"},
				},
			},
			{
				Title: "Execution",
				Bindings: []Binding{
					{"Enter", "run query (normal mode)"},
					{":run", "run query (command mode)"},
				},
			},
		},
	},
	{
		Name:  "results",
		Title: "Result View",
		Sections: []Section{
			{
				Title: "Navigation",
				Bindings: []Binding{
					{"h/j/k/l", "move left/down/up/right"},
					{"g / G", "first/last row"},
					{"0 / $", "first/last column"},
					{"ctrl+d", "half page down"},
					{"ctrl+u", "half page up"},
					{"ctrl+f", "full page down"},
					{"ctrl+b", "full page up"},
				},
			},
			{
				Title: "Selection",
				Bindings: []Binding{
					{"V", "visual line mode"},
					{"v", "visual block mode"},
					{"j/k/h/l", "extend selection"},
					{"tab", "toggle axis (line mode)"},
					{"y", "yank selection"},
					{"Esc", "cancel selection"},
				},
			},
			{
				Title: "Editing",
				Bindings: []Binding{
					{"e", "edit cell value"},
					{"dR", "delete current row"},
					{"oR", "insert new row"},
					{"u", "undo last change"},
					{"ctrl+s", "commit all changes"},
					{"y / Y", "yank cell/row"},
				},
			},
			{
				Title: "Commit/Rollback",
				Bindings: []Binding{
					{":commit", "apply pending changes to database"},
					{":rollback", "discard pending changes"},
					{":changes", "list pending changes"},
				},
			},
		},
	},
	{
		Name:  "commands",
		Title: "Commands",
		Sections: []Section{
			{
				Title: "General",
				Bindings: []Binding{
					{":q", "quit (warns on unsaved changes)"},
					{":q!", "force quit"},
					{":run", "execute query"},
					{":clear", "clear query buffer"},
					{":set <opt>", "toggle setting (e.g. autocommit)"},
					{":help [topic]", "show help"},
				},
			},
			{
				Title: "Buffers",
				Bindings: []Binding{
					{":new", "create new buffer"},
					{":bd", "close current buffer"},
					{":bn / :bp", "next/prev buffer"},
					{":b N", "switch to buffer N"},
					{":ls", "list buffers"},
				},
			},
			{
				Title: "Data",
				Bindings: []Binding{
					{":commit", "apply pending changes"},
					{":w", "apply pending changes"},
					{":rollback", "discard pending changes"},
					{":rollback!", "force discard"},
					{":changes", "list pending changes"},
					{":export <fmt> <file>", "export results (csv/json/sql)"},
				},
			},
			{
				Title: "Connection & Schema",
				Bindings: []Binding{
					{":connect", "open connection selector"},
					{":refresh", "reload schema from database"},
					{":dump [table]", "dump table (or database if no arg)"},
					{":theme", "list available themes"},
					{":theme <name>", "switch theme"},
				},
			},
		},
	},
	{
		Name:  "export",
		Title: "Export",
		Sections: []Section{
			{
				Title: "Usage",
				Bindings: []Binding{
					{":export csv <file>", "export as CSV"},
					{":export json <file>", "export as JSON"},
					{":export sql <file>", "export as SQL INSERT statements"},
				},
			},
		},
	},
	{
		Name:  "editing",
		Title: "Data Editing",
		Sections: []Section{
			{
				Title: "Cell Editing",
				Bindings: []Binding{
					{"e", "edit cell value (opens dialog)"},
					{"Enter", "confirm edit"},
					{"Esc", "cancel edit"},
				},
			},
			{
				Title: "Row Operations",
				Bindings: []Binding{
					{"dR", "mark row for deletion"},
					{"oR", "insert new blank row"},
					{"u", "undo last change"},
				},
			},
			{
				Title: "Committing",
				Bindings: []Binding{
					{"ctrl+s", "commit all pending changes"},
					{":commit", "commit all pending changes"},
					{":rollback", "discard pending changes"},
					{":set autocommit", "toggle autocommit mode"},
				},
			},
		},
	},
	{
		Name:  "connections",
		Title: "Connections",
		Sections: []Section{
			{
				Title: "Opening Selector",
				Bindings: []Binding{
					{"ctrl+o", "open connection selector"},
					{":connect", "open connection selector"},
				},
			},
			{
				Title: "In Selector",
				Bindings: []Binding{
					{"j / k", "navigate connections"},
					{"Enter", "connect to selected"},
					{"a", "add new connection"},
					{"e", "edit connection"},
					{"d", "delete connection"},
					{"Esc", "cancel"},
				},
			},
		},
	},
}

func findTopic(name string) *Topic {
	for i := range topics {
		if topics[i].Name == name {
			return &topics[i]
		}
	}
	return nil
}

func topicSections(name string) []Section {
	t := findTopic(name)
	if t == nil {
		return []Section{{
			Title: "Unknown topic: " + name,
			Bindings: []Binding{
				{"", "Available topics: " + strings.Join(TopicNames(), ", ")},
			},
		}}
	}
	return t.Sections
}
