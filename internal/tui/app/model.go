// Package app implements the main TUI application model.
package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/tui/components/commandbar"
	"github.com/mdjarv/db/internal/tui/components/queryeditor"
	"github.com/mdjarv/db/internal/tui/components/resultview"
	"github.com/mdjarv/db/internal/tui/components/statusbar"
	"github.com/mdjarv/db/internal/tui/components/tablelist"
	"github.com/mdjarv/db/internal/tui/core"
	"github.com/mdjarv/db/internal/tui/pane"
)

const (
	defaultLeftRatio = 0.20
	minPaneWidth     = 10
	minPaneHeight    = 3
)

// Model is the top-level bubbletea model composing all TUI components.
type Model struct {
	mode        core.Mode
	panes       *pane.Manager
	tableList   *tablelist.Model
	queryEditor *queryeditor.Model
	resultView  *resultview.Model
	statusBar   *statusbar.Model
	commandBar  *commandbar.Model

	width     int
	height    int
	leftRatio float64
	showHelp  bool
	ready     bool
}

// New creates the app model with all sub-components.
func New() Model {
	tl := tablelist.New()
	qe := queryeditor.New()
	rv := resultview.New()

	pm := pane.NewManager()
	pm.Register(pane.TableList, &paneAdapter{tablelist: tl})
	pm.Register(pane.QueryEditor, &paneAdapter{queryeditor: qe})
	pm.Register(pane.ResultView, &paneAdapter{resultview: rv})
	pm.SetActive(pane.TableList)

	return Model{
		mode:        core.ModeNormal,
		panes:       pm,
		tableList:   tl,
		queryEditor: qe,
		resultView:  rv,
		statusBar:   statusbar.New(),
		commandBar:  commandbar.New(),
		leftRatio:   defaultLeftRatio,
	}
}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	return tea.WindowSize()
}

// Update handles all messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.recalcLayout()
		return m, nil

	case commandbar.ExecuteMsg:
		return m.handleCommand(msg)

	case commandbar.CancelMsg:
		m.mode = core.ModeNormal
		m.statusBar.SetMode(m.mode)
		return m, nil

	case tea.KeyMsg:
		if m.commandBar.Active() {
			cmd := m.commandBar.Update(msg)
			return m, cmd
		}

		if m.showHelp {
			m.showHelp = false
			return m, nil
		}

		action := MatchGlobal(msg, m.mode)
		if action != ActionNone {
			return m.handleAction(action)
		}

		if m.mode == core.ModeNormal {
			active := m.panes.Active()
			if active != nil {
				_, cmd := active.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleAction(action Action) (tea.Model, tea.Cmd) {
	switch action {
	case ActionQuit:
		return m, tea.Quit
	case ActionModeNormal:
		m.mode = core.ModeNormal
		m.statusBar.SetMode(m.mode)
	case ActionModeInsert:
		m.mode = core.ModeInsert
		m.statusBar.SetMode(m.mode)
		m.statusBar.SetMessage("")
	case ActionModeCommand:
		m.mode = core.ModeCommand
		m.statusBar.SetMode(m.mode)
		m.commandBar.Activate()
	case ActionHelp:
		m.showHelp = !m.showHelp
	case ActionFocusNext:
		m.panes.CycleForward()
	case ActionFocusPrev:
		m.panes.CycleBackward()
	case ActionFocusLeft:
		m.panes.FocusLeft()
	case ActionFocusRight:
		m.panes.FocusRight()
	case ActionFocusUp:
		m.panes.FocusUp()
	case ActionFocusDown:
		m.panes.FocusDown()
	case ActionFocusPane1:
		m.panes.FocusByNumber(1)
	case ActionFocusPane2:
		m.panes.FocusByNumber(2)
	case ActionFocusPane3:
		m.panes.FocusByNumber(3)
	case ActionResizeGrow:
		m.leftRatio = min(m.leftRatio+0.05, 0.6)
		m.recalcLayout()
	case ActionResizeShrink:
		m.leftRatio = max(m.leftRatio-0.05, 0.1)
		m.recalcLayout()
	}
	return m, nil
}

func (m Model) handleCommand(msg commandbar.ExecuteMsg) (tea.Model, tea.Cmd) {
	m.mode = core.ModeNormal
	m.statusBar.SetMode(m.mode)

	switch msg.Command {
	case "q", "quit":
		return m, tea.Quit
	case "w":
		sql := m.queryEditor.Content()
		m.statusBar.SetMessage("Query: " + sql)
	case "set":
		m.statusBar.SetMessage("set: " + msg.Args)
	default:
		m.statusBar.SetMessage("unknown command: " + msg.Command)
	}
	return m, nil
}

func (m *Model) recalcLayout() {
	contentHeight := m.height - 1
	if m.commandBar.Active() {
		contentHeight--
	}

	leftW := max(int(float64(m.width)*m.leftRatio), minPaneWidth)
	rightW := max(m.width-leftW, minPaneWidth)

	topH := max(contentHeight/4, minPaneHeight)
	bottomH := max(contentHeight-topH, minPaneHeight)

	m.tableList.SetSize(leftW, contentHeight)
	m.queryEditor.SetSize(rightW, topH)
	m.resultView.SetSize(rightW, bottomH)
	m.statusBar.SetWidth(m.width)
	m.commandBar.SetWidth(m.width)
}

// View renders the full TUI.
func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	if m.showHelp {
		return m.helpView()
	}

	rightCol := lipgloss.JoinVertical(lipgloss.Left,
		m.queryEditor.View(),
		m.resultView.View(),
	)

	content := lipgloss.JoinHorizontal(lipgloss.Top,
		m.tableList.View(),
		rightCol,
	)

	var bottom string
	if m.commandBar.Active() {
		bottom = m.commandBar.View() + "\n" + m.statusBar.View()
	} else {
		bottom = m.statusBar.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left, content, bottom)
}

func (m Model) helpView() string {
	help := HelpText()
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(50)

	overlay := style.Render(help)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, overlay)
}

type paneAdapter struct {
	tablelist   *tablelist.Model
	queryeditor *queryeditor.Model
	resultview  *resultview.Model
}

func (a *paneAdapter) Update(msg tea.Msg) (pane.Pane, tea.Cmd) {
	switch {
	case a.tablelist != nil:
		cmd := a.tablelist.Update(msg)
		return a, cmd
	case a.queryeditor != nil:
		cmd := a.queryeditor.Update(msg)
		return a, cmd
	case a.resultview != nil:
		cmd := a.resultview.Update(msg)
		return a, cmd
	}
	return a, nil
}

func (a *paneAdapter) View() string {
	switch {
	case a.tablelist != nil:
		return a.tablelist.View()
	case a.queryeditor != nil:
		return a.queryeditor.View()
	case a.resultview != nil:
		return a.resultview.View()
	}
	return ""
}

func (a *paneAdapter) Focused() bool {
	switch {
	case a.tablelist != nil:
		return a.tablelist.Focused()
	case a.queryeditor != nil:
		return a.queryeditor.Focused()
	case a.resultview != nil:
		return a.resultview.Focused()
	}
	return false
}

func (a *paneAdapter) SetFocused(f bool) {
	switch {
	case a.tablelist != nil:
		a.tablelist.SetFocused(f)
	case a.queryeditor != nil:
		a.queryeditor.SetFocused(f)
	case a.resultview != nil:
		a.resultview.SetFocused(f)
	}
}

func (a *paneAdapter) SetSize(w, h int) {
	switch {
	case a.tablelist != nil:
		a.tablelist.SetSize(w, h)
	case a.queryeditor != nil:
		a.queryeditor.SetSize(w, h)
	case a.resultview != nil:
		a.resultview.SetSize(w, h)
	}
}
