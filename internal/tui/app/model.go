// Package app implements the main TUI application model.
package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/db"
	"github.com/mdjarv/db/internal/editor"
	"github.com/mdjarv/db/internal/export"
	"github.com/mdjarv/db/internal/schema"
	"github.com/mdjarv/db/internal/tui/components/commandbar"
	"github.com/mdjarv/db/internal/tui/components/dialog"
	"github.com/mdjarv/db/internal/tui/components/queryeditor"
	"github.com/mdjarv/db/internal/tui/components/resultview"
	"github.com/mdjarv/db/internal/tui/components/statusbar"
	"github.com/mdjarv/db/internal/tui/components/table"
	"github.com/mdjarv/db/internal/tui/components/tablelist"
	"github.com/mdjarv/db/internal/tui/core"
	"github.com/mdjarv/db/internal/tui/pane"
	"github.com/mdjarv/db/internal/tui/theme"
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
	dialog      *dialog.Model
	buffers     *BufferManager
	conn        db.Conn
	inspector   schema.Inspector

	// data editing
	changeBuf  *editor.ChangeBuffer
	autocommit bool
	editTable  string              // current table being edited
	editSchema string              // current schema
	editPKCols []string            // PK column names for current result
	editPKIdx  []int               // PK column indices in result set
	editCols   []schema.ColumnInfo // column info for current result

	width     int
	height    int
	leftRatio float64
	showHelp  bool
	ready     bool
	pending   string // for multi-key sequences (gt, gT)
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

	bm := NewBufferManager()

	return Model{
		mode:        core.ModeNormal,
		panes:       pm,
		tableList:   tl,
		queryEditor: qe,
		resultView:  rv,
		statusBar:   statusbar.New(),
		commandBar:  commandbar.New(),
		dialog:      dialog.New(),
		buffers:     bm,
		changeBuf:   editor.NewChangeBuffer(),
		leftRatio:   defaultLeftRatio,
	}
}

// NewWithConn creates the app model with a database connection.
func NewWithConn(conn db.Conn, insp schema.Inspector, connInfo string) Model {
	m := New()
	m.conn = conn
	m.inspector = insp
	m.statusBar.SetConn(connInfo)
	return m
}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{tea.WindowSize()}
	if m.inspector != nil {
		cmds = append(cmds, m.loadSchema())
	}
	return tea.Batch(cmds...)
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

	// M8: mode changes from components (a/A/I/o/O in queryeditor)
	case core.ModeChangedMsg:
		m.mode = msg.Mode
		m.statusBar.SetMode(m.mode)
		if msg.Mode == core.ModeInsert {
			m.statusBar.SetMessage("")
		}
		cmd := m.queryEditor.Update(msg)
		return m, cmd

	case core.QuerySubmittedMsg:
		if m.conn == nil {
			m.statusBar.SetMessage("not connected")
			return m, nil
		}
		m.statusBar.SetMessage("executing...")
		return m, m.executeQuery(msg.SQL)

	// M7: schema messages
	case core.SchemaLoadedMsg:
		cmd := m.tableList.Update(msg)
		if msg.Err != nil {
			m.statusBar.SetMessage("schema load failed: " + msg.Err.Error())
		} else {
			m.statusBar.SetMessage(fmt.Sprintf("loaded %d tables", len(msg.Tables)))
		}
		return m, cmd

	case core.TableSelectedMsg:
		m.statusBar.SetMessage(msg.Table.Name)
		if m.inspector != nil {
			return m, m.loadTableDetail(msg.Table)
		}
		return m, nil

	case core.TableDetailMsg:
		m.tableList.Update(msg)
		return m, nil

	case core.QueryRequestMsg:
		m.queryEditor.SetContent(msg.SQL)
		m.panes.SetActive(pane.QueryEditor)
		m.recalcLayout()
		m.statusBar.SetMessage("Query: " + msg.SQL)
		return m, nil

	case core.RefreshSchemaMsg:
		if m.inspector != nil {
			m.statusBar.SetMessage("refreshing schema...")
			return m, m.loadSchema()
		}
		return m, nil

	// M9: query result messages
	case core.QueryResultMsg:
		m.resultView.SetResult(msg.Columns, msg.Rows, msg.Duration)
		m.resultView.ClearModified()
		m.changeBuf.Clear()
		m.panes.SetActive(pane.ResultView)
		m.statusBar.SetMessage(fmt.Sprintf("Query OK: %d rows in %s", len(msg.Rows), msg.Duration))
		// sync result into active buffer
		buf := m.buffers.Active()
		buf.Columns = msg.Columns
		buf.Rows = msg.Rows
		buf.Duration = msg.Duration
		buf.HasData = true
		buf.ErrMsg = ""
		// set up editing context from query
		sql := m.queryEditor.Content()
		tableName, schemaName := parseTableFromSQL(sql)
		if tableName != "" {
			m.setEditContext(tableName, schemaName, msg.Columns)
		} else {
			m.editPKCols = nil
			m.editPKIdx = nil
		}
		return m, nil

	case core.QueryErrorMsg:
		m.resultView.SetError(msg.Err)
		m.statusBar.SetMessage("Query error: " + msg.Err.Error())
		buf := m.buffers.Active()
		buf.HasData = false
		buf.ErrMsg = msg.Err.Error()
		return m, nil

	case core.ExportRequestMsg:
		return m, m.exportResult(msg.Format, msg.Path)

	case exportResultMsg:
		if msg.err != nil {
			m.statusBar.SetMessage("export failed: " + msg.err.Error())
		} else {
			m.statusBar.SetMessage("exported to " + msg.path)
		}
		return m, nil

	case core.EditCellMsg:
		result, cmd := m.handleEditCell(msg)
		return result, cmd

	case core.EditCancelMsg:
		result, cmd := m.handleEditCancel()
		return result, cmd

	case core.DeleteRowMsg:
		result, cmd := m.handleDeleteRow(msg)
		return result, cmd

	case core.InsertRowMsg:
		result, cmd := m.handleInsertRow()
		return result, cmd

	case core.UndoMsg:
		result, cmd := m.handleUndo()
		return result, cmd

	case dialog.ResultMsg:
		result, cmd := m.handleDialogResult(msg)
		return result, cmd

	case commitResultMsg:
		result, cmd := m.handleCommitResult(msg)
		return result, cmd

	case core.YankMsg:
		m.mode = core.ModeNormal
		m.statusBar.SetMode(m.mode)
		return m, copyToClipboard(msg.Content)

	case yankResultMsg:
		if msg.err != nil {
			m.statusBar.SetMessage("yank failed: " + msg.err.Error())
		} else {
			m.statusBar.SetMessage("yanked to clipboard")
		}
		return m, nil

	case tea.KeyMsg:
		if m.dialog.IsActive() {
			cmd := m.dialog.Update(msg)
			return m, cmd
		}

		if m.commandBar.Active() {
			cmd := m.commandBar.Update(msg)
			return m, cmd
		}

		if m.showHelp {
			m.showHelp = false
			return m, nil
		}

		// forward keys to resultview when in edit mode
		if m.mode.IsEdit() {
			cmd := m.resultView.Update(msg)
			return m, cmd
		}

		if m.mode.IsVisual() {
			if msg.String() == "esc" {
				m.resultView.ExitVisual()
				m.mode = core.ModeNormal
				m.statusBar.SetMode(m.mode)
				return m, nil
			}
			cmd := m.resultView.Update(msg)
			return m, cmd
		}

		// M12: multi-key buffer switching (gt/gT)
		if m.mode == core.ModeNormal && m.pending == "g" {
			m.pending = ""
			switch msg.String() {
			case "t":
				return m.handleAction(ActionBufferNext)
			case "T":
				return m.handleAction(ActionBufferPrev)
			default:
				// not a buffer key — forward "g" then current key to pane
				active := m.panes.Active()
				if active != nil {
					gMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")}
					active.Update(gMsg)
					_, cmd := active.Update(msg)
					cmds = append(cmds, cmd)
				}
				m.recalcLayout()
				return m, tea.Batch(cmds...)
			}
		}
		if m.mode == core.ModeNormal && msg.String() == "g" {
			m.pending = "g"
			return m, nil
		}

		// v/V on query editor → editor-local visual mode, not result table visual
		if m.mode == core.ModeNormal && m.panes.ActiveID() == pane.QueryEditor {
			if k := msg.String(); k == "v" || k == "V" {
				cmd := m.queryEditor.Update(msg)
				m.recalcLayout()
				return m, cmd
			}
		}

		action := MatchGlobal(msg, m.mode)
		if action != ActionNone {
			return m.handleAction(action)
		}

		// M8: forward keys to active pane in normal mode,
		// or to queryeditor in insert mode
		active := m.panes.Active()
		if active != nil {
			if m.mode == core.ModeNormal {
				_, cmd := active.Update(msg)
				cmds = append(cmds, cmd)
			} else if m.mode == core.ModeInsert && m.panes.ActiveID() == pane.QueryEditor {
				cmd := m.queryEditor.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
		m.recalcLayout()
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
		m.queryEditor.Update(core.ModeChangedMsg{Mode: core.ModeNormal})
	case ActionModeInsert:
		m.mode = core.ModeInsert
		m.statusBar.SetMode(m.mode)
		m.statusBar.SetMessage("")
		m.queryEditor.Update(core.ModeChangedMsg{Mode: core.ModeInsert})
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
	case ActionModeVisualLine:
		if m.panes.ActiveID() == pane.ResultView {
			m.mode = core.ModeVisualLine
			m.statusBar.SetMode(m.mode)
			m.resultView.EnterVisualLine()
		}
	case ActionModeVisualBlock:
		if m.panes.ActiveID() == pane.ResultView {
			m.mode = core.ModeVisualBlock
			m.statusBar.SetMode(m.mode)
			m.resultView.EnterVisualBlock()
		}
	case ActionBufferNext:
		m.saveBufferState()
		m.buffers.Next()
		m.restoreBufferState()
		m.statusBar.SetMessage(fmt.Sprintf("buffer %d", m.buffers.ActiveIndex()))
	case ActionBufferPrev:
		m.saveBufferState()
		m.buffers.Prev()
		m.restoreBufferState()
		m.statusBar.SetMessage(fmt.Sprintf("buffer %d", m.buffers.ActiveIndex()))
	}
	return m, nil
}

type yankResultMsg struct {
	err error
}

type exportResultMsg struct {
	path string
	err  error
}

func copyToClipboard(content string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("wl-copy", "--")
		cmd.Stdin = strings.NewReader(content)
		return yankResultMsg{err: cmd.Run()}
	}
}

func (m Model) executeQuery(sql string) tea.Cmd {
	conn := m.conn
	return func() tea.Msg {
		ctx := context.Background()
		result, err := conn.Query(ctx, sql)
		if err != nil {
			return core.QueryErrorMsg{Err: err}
		}
		defer result.Rows.Close()

		cols := make([]core.ResultColumn, len(result.Columns))
		for i, c := range result.Columns {
			cols[i] = core.ResultColumn{Name: c.Name, TypeName: c.TypeName}
		}

		var rows [][]string
		for result.Rows.Next() {
			vals, err := result.Rows.Values()
			if err != nil {
				return core.QueryErrorMsg{Err: err}
			}
			row := make([]string, len(vals))
			for i, v := range vals {
				if v == nil {
					row[i] = table.NullPlaceholder
				} else {
					row[i] = formatValue(v)
				}
			}
			rows = append(rows, row)
		}
		if err := result.Rows.Err(); err != nil {
			return core.QueryErrorMsg{Err: err}
		}

		return core.QueryResultMsg{Columns: cols, Rows: rows}
	}
}

func formatValue(v any) string {
	switch val := v.(type) {
	case [16]byte:
		return fmt.Sprintf("%x-%x-%x-%x-%x", val[0:4], val[4:6], val[6:8], val[8:10], val[10:16])
	default:
		return fmt.Sprintf("%v", v)
	}
}

// M7: schema loading

func (m Model) loadSchema() tea.Cmd {
	insp := m.inspector
	return func() tea.Msg {
		tables, err := insp.Tables(context.Background(), "public")
		return core.SchemaLoadedMsg{Tables: tables, Err: err}
	}
}

func (m Model) loadTableDetail(t schema.Table) tea.Cmd {
	insp := m.inspector
	return func() tea.Msg {
		ctx := context.Background()
		cols, _ := insp.Columns(ctx, t.Schema, t.Name)
		idxs, _ := insp.Indexes(ctx, t.Schema, t.Name)
		cons, _ := insp.Constraints(ctx, t.Schema, t.Name)
		fks, _ := insp.ForeignKeys(ctx, t.Schema, t.Name)
		return core.TableDetailMsg{
			Table:       t,
			Columns:     cols,
			Indexes:     idxs,
			Constraints: cons,
			ForeignKeys: fks,
		}
	}
}

// M9: export support

func (m *Model) parseExport(args string) tea.Cmd {
	parts := strings.SplitN(strings.TrimSpace(args), " ", 2)
	if len(parts) < 2 || parts[1] == "" {
		return func() tea.Msg {
			return exportResultMsg{err: fmt.Errorf("usage: export <csv|json|sql> <file>")}
		}
	}
	format := parts[0]
	path := parts[1]
	switch format {
	case "csv", "json", "sql":
	default:
		return func() tea.Msg {
			return exportResultMsg{err: fmt.Errorf("unknown format: %s (csv, json, sql)", format)}
		}
	}
	return func() tea.Msg {
		return core.ExportRequestMsg{Format: format, Path: path}
	}
}

func (m *Model) exportResult(format, path string) tea.Cmd {
	cols, rows := m.resultView.ResultData()
	if cols == nil {
		return func() tea.Msg {
			return exportResultMsg{err: fmt.Errorf("no results to export")}
		}
	}

	dbCols := make([]db.Column, len(cols))
	for i, c := range cols {
		dbCols[i] = db.Column{Name: c.Name, TypeName: c.TypeName}
	}

	rowsCopy := make([][]string, len(rows))
	copy(rowsCopy, rows)

	return func() tea.Msg {
		f, err := os.Create(path)
		if err != nil {
			return exportResultMsg{path: path, err: err}
		}

		var efmt export.Format
		switch format {
		case "csv":
			efmt = export.FormatCSV
		case "json":
			efmt = export.FormatJSON
		case "sql":
			efmt = export.FormatSQL
		}

		iter := &sliceRowIterator{data: rowsCopy}
		result := &db.Result{
			Columns: dbCols,
			Rows:    iter,
		}

		exp := export.NewExporter(efmt, export.Options{
			NullString: table.NullPlaceholder,
		})
		err = exp.Export(f, result)
		if cerr := f.Close(); err == nil {
			err = cerr
		}
		return exportResultMsg{path: path, err: err}
	}
}

func (m Model) handleCommand(msg commandbar.ExecuteMsg) (tea.Model, tea.Cmd) {
	m.mode = core.ModeNormal
	m.statusBar.SetMode(m.mode)

	switch msg.Command {
	case "q", "quit":
		return m, tea.Quit
	case "w":
		sql := m.queryEditor.Content()
		return m, func() tea.Msg {
			return core.QuerySubmittedMsg{SQL: sql}
		}
	case "clear":
		m.queryEditor.SetContent("")
		m.recalcLayout()
		m.statusBar.SetMessage("buffer cleared")
	case "set":
		m.handleSetCommand(msg.Args)
	case "commit":
		result, cmd := m.handleCommit()
		return result, cmd
	case "rollback":
		result, cmd := m.handleRollback()
		return result, cmd
	case "changes":
		result, cmd := m.handleChanges()
		return result, cmd
	case "export":
		return m, m.parseExport(msg.Args)
	case "new", "enew":
		m.saveBufferState()
		if !m.buffers.New() {
			m.statusBar.SetMessage("max buffers reached")
			return m, nil
		}
		m.restoreBufferState()
		m.statusBar.SetMessage(fmt.Sprintf("buffer %d", m.buffers.ActiveIndex()))
	case "bd":
		if !m.buffers.Close() {
			m.statusBar.SetMessage("cannot close last buffer")
			return m, nil
		}
		m.restoreBufferState()
		m.statusBar.SetMessage(fmt.Sprintf("buffer %d", m.buffers.ActiveIndex()))
	case "bn":
		m.saveBufferState()
		m.buffers.Next()
		m.restoreBufferState()
		m.statusBar.SetMessage(fmt.Sprintf("buffer %d", m.buffers.ActiveIndex()))
	case "bp":
		m.saveBufferState()
		m.buffers.Prev()
		m.restoreBufferState()
		m.statusBar.SetMessage(fmt.Sprintf("buffer %d", m.buffers.ActiveIndex()))
	case "b":
		n := 0
		if _, err := fmt.Sscanf(msg.Args, "%d", &n); err != nil {
			m.statusBar.SetMessage("invalid buffer number")
			return m, nil
		}
		m.saveBufferState()
		if !m.buffers.SwitchTo(n) {
			m.statusBar.SetMessage("invalid buffer number")
			return m, nil
		}
		m.restoreBufferState()
		m.statusBar.SetMessage(fmt.Sprintf("buffer %d", m.buffers.ActiveIndex()))
	case "ls", "buffers":
		m.saveBufferState()
		m.statusBar.SetMessage(m.buffers.List())
	case "theme":
		if msg.Args == "" {
			names := theme.Available()
			m.statusBar.SetMessage("themes: " + strings.Join(names, ", "))
		} else {
			t, err := theme.Resolve(msg.Args)
			if err != nil {
				m.statusBar.SetMessage("unknown theme: " + msg.Args)
			} else {
				theme.Set(t)
				m.statusBar.SetMessage("theme: " + t.Name)
			}
		}
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

	maxTopH := max(contentHeight/4, minPaneHeight)
	topH := max(min(m.queryEditor.LineCount()+2, maxTopH), minPaneHeight)
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
	lines := strings.Split(help, "\n")

	// border(2) + padding(2) + some margin
	maxLines := m.height - 6
	if maxLines < 5 {
		maxLines = 5
	}
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Current().Styles.BorderFocused).
		Padding(1, 2).
		Width(50)

	overlay := style.Render(strings.Join(lines, "\n"))

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

// saveBufferState saves the current editor/result state into the active buffer.
func (m *Model) saveBufferState() {
	buf := m.buffers.Active()
	buf.Query = m.queryEditor.Content()
	buf.Modified = buf.Query != ""

	cols, rows := m.resultView.ResultData()
	buf.Columns = cols
	buf.Rows = rows
	buf.HasData = cols != nil

	buf.CursorRow = m.resultView.TableCursorRow()
	buf.CursorCol = m.resultView.TableCursorCol()
	buf.RowOffset = m.resultView.TableRowOffset()
	buf.ColOffset = m.resultView.TableColOffset()
}

// restoreBufferState loads the active buffer state into editor/result components.
func (m *Model) restoreBufferState() {
	buf := m.buffers.Active()
	m.queryEditor.SetContent(buf.Query)

	if buf.HasData {
		m.resultView.SetResult(buf.Columns, buf.Rows, buf.Duration)
		m.resultView.SetTableCursor(buf.CursorRow, buf.CursorCol, buf.RowOffset, buf.ColOffset)
	} else if buf.ErrMsg != "" {
		m.resultView.SetError(fmt.Errorf("%s", buf.ErrMsg))
	} else {
		m.resultView.Clear()
	}

	m.statusBar.SetBuffer(m.buffers.ActiveIndex(), m.buffers.Count())
	m.recalcLayout()
}

type sliceRowIterator struct {
	data [][]string
	pos  int
}

func (it *sliceRowIterator) Next() bool {
	if it.pos < len(it.data) {
		it.pos++
		return true
	}
	return false
}

func (it *sliceRowIterator) Values() ([]any, error) {
	row := it.data[it.pos-1]
	vals := make([]any, len(row))
	for i, v := range row {
		if v == table.NullPlaceholder {
			vals[i] = nil
		} else {
			vals[i] = v
		}
	}
	return vals, nil
}

func (it *sliceRowIterator) Err() error { return nil }
func (it *sliceRowIterator) Close()     {}
