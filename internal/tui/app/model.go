// Package app implements the main TUI application model.
package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mdjarv/db/internal/conn"
	"github.com/mdjarv/db/internal/db"
	"github.com/mdjarv/db/internal/dump"
	"github.com/mdjarv/db/internal/editor"
	"github.com/mdjarv/db/internal/export"
	"github.com/mdjarv/db/internal/schema"
	"github.com/mdjarv/db/internal/tui/components/bufferlist"
	"github.com/mdjarv/db/internal/tui/components/commandbar"
	"github.com/mdjarv/db/internal/tui/components/connform"
	"github.com/mdjarv/db/internal/tui/components/connselector"
	"github.com/mdjarv/db/internal/tui/components/dialog"
	"github.com/mdjarv/db/internal/tui/components/dumpform"
	"github.com/mdjarv/db/internal/tui/components/editdialog"
	"github.com/mdjarv/db/internal/tui/components/helpoverlay"
	"github.com/mdjarv/db/internal/tui/components/queryeditor"
	"github.com/mdjarv/db/internal/tui/components/resultview"
	"github.com/mdjarv/db/internal/tui/components/statusbar"
	"github.com/mdjarv/db/internal/tui/components/table"
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
	mode         core.Mode
	panes        *pane.Manager
	tableList    *tablelist.Model
	queryEditor  *queryeditor.Model
	resultView   *resultview.Model
	statusBar    *statusbar.Model
	commandBar   *commandbar.Model
	dialog       *dialog.Model
	editDialog   *editdialog.Model
	connSelector *connselector.Model
	connForm     *connform.Model
	dumpForm     *dumpform.Model
	dumpProgress *dialog.ProgressModel
	bufferList   *bufferlist.Model






	buffers      *BufferManager
	conn         db.Conn
	inspector    schema.Inspector

	// connection management
	stores                 []*conn.Store
	creds                  *conn.CredentialStore
	gitRoot                string
	pendingDeleteCandidate *conn.Candidate
	lastCandidate          *conn.Candidate // last successful connection, for reconnect
	disconnected           bool            // true when connection lost
	dumpCancel             context.CancelFunc

	// data editing
	changeBuf  *editor.ChangeBuffer
	autocommit bool
	editTable  string              // current table being edited
	editSchema string              // current schema
	editPKCols []string            // PK column names for current result
	editPKIdx  []int               // PK column indices in result set
	editCols   []schema.ColumnInfo // column info for current result

	queryTimeout time.Duration // 0 = no timeout

	width       int
	height      int
	leftRatio   float64
	showHelp    bool // kept for test compat, mirrors helpOverlay.IsActive()
	helpScroll  int  // kept for test compat
	helpOverlay *helpoverlay.Model
	ready       bool
	keySeq      core.KeySeq
}

// Options configures the app model.
type Options struct {
	Conn      db.Conn
	Inspector schema.Inspector
	ConnInfo  string
	Stores    []*conn.Store
	Creds     *conn.CredentialStore
	GitRoot   string
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
		mode:         core.ModeNormal,
		panes:        pm,
		tableList:    tl,
		queryEditor:  qe,
		resultView:   rv,
		statusBar:    statusbar.New(),
		commandBar:   commandbar.New(),
		dialog:       dialog.New(),
		editDialog:   editdialog.New(),
		connSelector: connselector.New(),
		connForm:     connform.New(),
		dumpForm:     dumpform.New(),
		dumpProgress: dialog.NewProgress(),
		bufferList:   bufferlist.New(),
		helpOverlay:  helpoverlay.New(),
		buffers:      bm,
		changeBuf:    editor.NewChangeBuffer(),
		leftRatio:    defaultLeftRatio,
	}
}

// NewWithConn creates the app model with a database connection.
func NewWithConn(c db.Conn, insp schema.Inspector, connInfo string) Model {
	m := New()
	m.conn = c
	m.inspector = insp
	m.statusBar.SetConn(connInfo)
	return m
}

// NewWithOpts creates the app model with full options.
func NewWithOpts(opts Options) Model {
	m := New()
	m.conn = opts.Conn
	m.inspector = opts.Inspector
	m.stores = opts.Stores
	m.creds = opts.Creds
	m.gitRoot = opts.GitRoot
	if opts.ConnInfo != "" {
		m.statusBar.SetConn(opts.ConnInfo)
	}
	return m
}

// Cleanup closes the database connection. Call after tea.Program exits.
func (m Model) Cleanup() {
	if m.conn != nil {
		_ = m.conn.Close(context.Background())
	}
}

// Init returns the initial command.
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{tea.WindowSize()}
	if m.conn == nil && (len(m.stores) > 0 || m.gitRoot != "") {
		cmds = append(cmds, m.discoverConnections())
	} else if m.inspector != nil {
		cmds = append(cmds, m.loadSchema())
	}
	return tea.Batch(cmds...)
}

// Update handles all messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.recalcLayout()
		return m, nil

	case core.ClearErrorMsg:
		m.statusBar.HandleClearError(msg)
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
		if msg.Err != nil && db.IsConnectionError(msg.Err) {
			return m.handleConnectionLost()
		}
		cmd := m.tableList.Update(msg)
		if msg.Err != nil {
			errCmd := m.statusBar.SetError("schema load failed: " + msg.Err.Error())
			return m, tea.Batch(cmd, errCmd)
		}
		m.statusBar.SetSuccess(fmt.Sprintf("loaded %d tables", len(msg.Tables)))
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
		m.recalcLayout()
		if m.conn != nil {
			m.statusBar.SetMessage("executing...")
			return m, m.executeQuery(msg.SQL)
		}
		m.panes.SetActive(pane.QueryEditor)
		m.statusBar.SetMessage("Query: " + msg.SQL)
		return m, nil

	case core.RefreshSchemaMsg:
		if m.inspector != nil {
			if ci, ok := m.inspector.(*schema.CachedInspector); ok {
				ci.Invalidate()
			}
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
		m.statusBar.SetSuccess(fmt.Sprintf("Query OK: %d rows in %s", len(msg.Rows), msg.Duration))
		// sync result into active buffer
		buf := m.buffers.Active()
		buf.Columns = msg.Columns
		buf.Rows = msg.Rows
		buf.Duration = msg.Duration
		buf.HasData = true
		buf.ErrMsg = ""
		buf.LastExecutedQuery = m.queryEditor.Content()
		buf.Modified = false
		m.statusBar.SetBufferModified(false)
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
		if db.IsConnectionError(msg.Err) {
			return m.handleConnectionLost()
		}
		m.resultView.SetError(msg.Err)
		errCmd := m.statusBar.SetError("Query error: " + msg.Err.Error())
		buf := m.buffers.Active()
		buf.HasData = false
		buf.ErrMsg = msg.Err.Error()
		return m, errCmd

	case core.ExportRequestMsg:
		return m, m.exportResult(msg.Format, msg.Path)

	case exportResultMsg:
		if msg.err != nil {
			errCmd := m.statusBar.SetError("export failed: " + msg.err.Error())
			return m, errCmd
		}
		m.statusBar.SetSuccess("exported to " + msg.path)
		return m, nil

	case core.EditRequestMsg:
		return m.handleEditRequest(msg)

	case editdialog.SubmitMsg:
		return m.handleEditSubmit(msg)

	case editdialog.CancelMsg:
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

	case core.ConnSelectorMsg:
		m.connSelector.Open(msg.Candidates)
		return m, nil

	case connselector.SelectMsg:
		m.connSelector.Close()
		m.statusBar.SetMessage("connecting...")
		return m, m.connectTo(msg.Candidate)

	case connselector.CancelMsg:
		if m.conn == nil {
			return m, tea.Quit
		}
		return m, nil

	case connselector.QuitMsg:
		return m, tea.Quit

	case connselector.AddMsg:
		if len(m.stores) == 0 {
			errCmd := m.statusBar.SetError("no connection store available")
			return m, errCmd
		}
		m.connForm.OpenAdd(msg.Source)
		return m, nil

	case connselector.EditMsg:
		password := ""
		if m.creds != nil {
			password, _ = m.creds.GetPassword(msg.Candidate.Config.Name)
		}
		m.connForm.OpenEdit(msg.Candidate.Config, password, msg.Candidate.Source)
		return m, nil

	case connselector.DeleteMsg:
		c := msg.Candidate
		m.pendingDeleteCandidate = &c
		m.dialog.Open("delete-conn", "Delete connection?", c.Config.Name)
		return m, nil

	case connform.SubmitMsg:
		return m.handleConnFormSubmit(msg)

	case connform.CancelMsg:
		return m, nil

	case dumpform.SubmitMsg:
		return m.handleDumpFormSubmit(msg)

	case dumpform.CancelMsg:
		return m, nil

	case core.DumpTableMsg:
		return m.openDumpForm(msg.Table, false)

	case core.DumpSchemaMsg:
		return m.openDumpForm(msg.Table, true)

	case core.DumpProgressMsg:
		return m.handleDumpProgress(msg)

	case core.DumpCompleteMsg:
		return m.handleDumpComplete(msg)

	case dialog.ProgressCancelMsg:
		return m.handleDumpCancel()

	case connSelectorRefreshMsg:
		m.connSelector.Refresh(msg.candidates, msg.restoreName)
		return m, nil

	case core.ConnectedMsg:
		if m.conn != nil {
			_ = m.conn.Close(context.Background())
		}
		m.conn = msg.Conn
		m.inspector = msg.Inspector
		m.disconnected = false
		c := msg.Candidate
		m.lastCandidate = &c
		m.statusBar.SetConn(msg.ConnInfo)
		m.changeBuf.Clear()
		m.resultView.ClearModified()
		m.editPKCols = nil
		m.editPKIdx = nil
		m.statusBar.SetSuccess("connected: " + msg.ConnInfo)
		m.saveDefault(msg.Candidate)
		return m, m.loadSchema()

	case core.ConnectErrorMsg:
		errCmd := m.statusBar.SetError("connection failed: " + msg.Err.Error())
		if m.disconnected {
			// reconnect attempt failed — offer retry
			m.dialog.Open("reconnect", "Reconnect failed", "Try again?")
			return m, errCmd
		}
		if m.conn == nil {
			return m, tea.Batch(errCmd, m.discoverConnections())
		}
		return m, errCmd

	case core.YankMsg:
		m.mode = core.ModeNormal
		m.statusBar.SetMode(m.mode)
		return m, copyToClipboard(msg.Content)

	case yankResultMsg:
		if msg.err != nil {
			errCmd := m.statusBar.SetError("yank failed: " + msg.err.Error())
			return m, errCmd
		}
		m.statusBar.SetSuccess("yanked to clipboard")
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyInput(msg)
	}

	return m, nil
}

func (m Model) handleAction(action Action) (tea.Model, tea.Cmd) {
	switch action {
	case ActionQuit:
		if m.changeBuf.Len() > 0 {
			m.dialog.Open("quit", "Uncommitted changes",
				fmt.Sprintf("%d pending changes will be lost. Quit?", m.changeBuf.Len()))
			return m, nil
		}
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
		if m.helpOverlay.IsActive() {
			m.helpOverlay.Close()
		} else {
			m.helpOverlay.Open()
		}
		m.showHelp = m.helpOverlay.IsActive()
		m.helpScroll = 0
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
	case ActionCommit:
		return m.handleCommit()
	case ActionConnSelector:
		if m.changeBuf.Len() > 0 {
			m.dialog.Open("switch-conn", "Uncommitted changes",
				fmt.Sprintf("%d pending changes will be lost. Switch?", m.changeBuf.Len()))
			return m, nil
		}
		return m, m.discoverConnections()
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
	timeout := m.queryTimeout
	return func() tea.Msg {
		ctx := context.Background()
		if timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
		result, err := conn.Query(ctx, sql)
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return core.QueryErrorMsg{Err: fmt.Errorf("query timed out after %s", timeout)}
			}
			return core.QueryErrorMsg{Err: err}
		}
		defer result.Rows.Close()

		cols := make([]core.ResultColumn, len(result.Columns))
		for i, c := range result.Columns {
			cols[i] = core.ResultColumn{
				Name:            c.Name,
				TypeName:        c.TypeName,
				EnumValues:      c.EnumValues,
				CompositeFields: convertCompositeFields(c.CompositeFields),
			}
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

func convertCompositeFields(fields []db.CompositeField) []core.CompositeField {
	if fields == nil {
		return nil
	}
	out := make([]core.CompositeField, len(fields))
	for i, f := range fields {
		out[i] = core.CompositeField{Name: f.Name, TypeName: f.TypeName}
	}
	return out
}

func formatValue(v any) string {
	switch val := v.(type) {
	case [16]byte:
		return fmt.Sprintf("%x-%x-%x-%x-%x", val[0:4], val[4:6], val[6:8], val[8:10], val[10:16])
	case []any:
		return formatSlice(val)
	case []string:
		parts := make([]any, len(val))
		for i, s := range val {
			parts[i] = s
		}
		return formatSlice(parts)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func formatSlice(vals []any) string {
	var sb strings.Builder
	sb.WriteByte('{')
	for i, v := range vals {
		if i > 0 {
			sb.WriteByte(',')
		}
		if v == nil {
			sb.WriteString("NULL")
		} else {
			fmt.Fprintf(&sb, "%v", v)
		}
	}
	sb.WriteByte('}')
	return sb.String()
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

// Connection management

func (m Model) discoverConnections() tea.Cmd {
	stores := m.stores
	creds := m.creds
	gitRoot := m.gitRoot
	return func() tea.Msg {
		candidates := conn.Discover(conn.DiscoverOptions{
			Stores:  stores,
			Creds:   creds,
			GitRoot: gitRoot,
		})
		return core.ConnSelectorMsg{Candidates: candidates}
	}
}

// handleConnectionLost marks the connection as lost and shows the reconnect dialog.
func (m *Model) handleConnectionLost() (Model, tea.Cmd) {
	m.disconnected = true
	m.conn = nil
	m.inspector = nil
	m.statusBar.SetConn("")
	m.statusBar.SetError("connection lost")
	if m.lastCandidate != nil {
		cfg := m.lastCandidate.Config
		label := cfg.Name
		if label == "" {
			label = fmt.Sprintf("%s@%s/%s", cfg.User, cfg.Host, cfg.DBName)
		}
		m.dialog.Open("reconnect", "Connection lost", "Reconnect to "+label+"?")
	} else {
		m.dialog.Open("reconnect", "Connection lost", "Reconnect?")
	}
	return *m, nil
}

func (m Model) connectTo(candidate conn.Candidate) tea.Cmd {
	cfg := candidate.Config
	return func() tea.Msg {
		c, err := db.Open(context.Background(), "postgres", cfg.DSN())
		if err != nil {
			return core.ConnectErrorMsg{Err: err}
		}
		connInfo := fmt.Sprintf("%s@%s/%s", cfg.User, cfg.Host, cfg.DBName)
		insp := schema.NewCachedInspector(schema.NewPostgresInspector(c))
		return core.ConnectedMsg{Conn: c, Inspector: insp, ConnInfo: connInfo, Candidate: candidate}
	}
}

// saveDefault persists the selected candidate as the default connection.
func (m *Model) saveDefault(candidate conn.Candidate) {
	if len(m.stores) == 0 {
		return
	}

	cfg := candidate.Config

	switch candidate.Source {
	case conn.SourceProjectStore, conn.SourceGlobalStore:
		// already in a store — just set default on the matching store
		if cfg.Name == "" {
			return
		}
		for _, s := range m.stores {
			if _, err := s.Get(cfg.Name); err == nil {
				if err := s.SetDefault(cfg.Name); err != nil {
					m.statusBar.SetMessage("default save failed: " + err.Error())
				}
				return
			}
		}

	case conn.SourceEnvVar, conn.SourceDotEnv:
		// not in any store — save to preferred store, then set default
		name := fmt.Sprintf("%s@%s/%s", cfg.User, cfg.Host, cfg.DBName)
		password := cfg.Password
		cfg.Name = name
		store := m.stores[0] // project store if available, else global
		if err := store.Add(cfg); err != nil {
			m.statusBar.SetMessage("save connection failed: " + err.Error())
			return
		}
		if password != "" && m.creds != nil {
			_ = m.creds.SetPassword(name, password)
		}
		if err := store.SetDefault(name); err != nil {
			m.statusBar.SetMessage("default save failed: " + err.Error())
		}
	}
}

// Dump support

// openDumpForm opens the dump form with connection info pre-filled.
func (m *Model) openDumpForm(tableName string, schemaOnly bool) (Model, tea.Cmd) {
	if m.lastCandidate == nil {
		errCmd := m.statusBar.SetError("not connected")
		return *m, errCmd
	}
	cfg := m.lastCandidate.Config
	port := "5432"
	if cfg.Port != 0 {
		port = fmt.Sprintf("%d", cfg.Port)
	}
	password := cfg.Password
	if password == "" && m.creds != nil {
		password, _ = m.creds.GetPassword(cfg.Name)
	}
	if schemaOnly {
		m.dumpForm.OpenSchemaOnly(tableName, cfg.DBName, cfg.Host, port, cfg.User, password, cfg.SSLMode)
	} else {
		m.dumpForm.Open(tableName, cfg.DBName, cfg.Host, port, cfg.User, password, cfg.SSLMode)
	}
	return *m, nil
}

// handleDumpFormSubmit starts the dump process.
func (m *Model) handleDumpFormSubmit(msg dumpform.SubmitMsg) (Model, tea.Cmd) {
	binPath, err := dump.FindPgDump("")
	if err != nil {
		errCmd := m.statusBar.SetError(err.Error())
		return *m, errCmd
	}

	m.dumpProgress.Open("Dumping " + msg.Config.DBName)

	ctx, cancel := context.WithCancel(context.Background())
	m.dumpCancel = cancel

	cfg := msg.Config
	return *m, m.runDump(ctx, binPath, cfg)
}

// runDump starts pg_dump and returns a tea.Cmd that sends progress/complete messages.
func (m *Model) runDump(ctx context.Context, binPath string, cfg dump.Config) tea.Cmd {
	return func() tea.Msg {
		runner := dump.NewRunner(binPath)
		start := time.Now()

		ch, err := runner.Run(ctx, cfg, 0)
		if err != nil {
			return core.DumpCompleteMsg{
				Path: cfg.OutputPath,
				Err:  err,
			}
		}

		// Drain progress channel, forwarding events via a program-level
		// subscription is not possible from a tea.Cmd. Instead we collect
		// all events and emit the final result. Progress updates are sent
		// as intermediate messages via the returned Batch.
		var lastEvent dump.ProgressEvent
		for ev := range ch {
			lastEvent = ev
		}

		if lastEvent.Err != nil {
			return core.DumpCompleteMsg{
				Path:     cfg.OutputPath,
				Duration: time.Since(start),
				Err:      lastEvent.Err,
			}
		}

		var size int64
		if fi, err := os.Stat(cfg.OutputPath); err == nil {
			size = fi.Size()
		}

		return core.DumpCompleteMsg{
			Path:     cfg.OutputPath,
			Size:     size,
			Duration: time.Since(start),
		}
	}
}

// handleDumpProgress updates the progress modal.
func (m *Model) handleDumpProgress(msg core.DumpProgressMsg) (Model, tea.Cmd) {
	ev := msg.Event
	if ev.Err != nil {
		m.dumpProgress.Close()
		errCmd := m.statusBar.SetError("dump failed: " + ev.Err.Error())
		return *m, errCmd
	}
	m.dumpProgress.SetProgress(ev.Object, ev.Index, ev.Total)
	return *m, nil
}

// handleDumpComplete dismisses the progress modal and shows result.
func (m *Model) handleDumpComplete(msg core.DumpCompleteMsg) (Model, tea.Cmd) {
	m.dumpProgress.Close()
	m.dumpCancel = nil
	if msg.Err != nil {
		errCmd := m.statusBar.SetError("dump failed: " + msg.Err.Error())
		return *m, errCmd
	}
	sizeStr := formatSize(msg.Size)
	m.statusBar.SetSuccess(fmt.Sprintf("dump complete: %s (%s, %s)", msg.Path, sizeStr, msg.Duration.Truncate(time.Millisecond)))
	return *m, nil
}

// handleDumpCancel cancels a running dump.
func (m *Model) handleDumpCancel() (Model, tea.Cmd) {
	if m.dumpCancel != nil {
		m.dumpCancel()
		m.dumpCancel = nil
	}
	m.dumpProgress.Close()
	m.statusBar.SetMessage("dump cancelled")
	return *m, nil
}

func formatSize(bytes int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case bytes >= gb:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
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

// handleKeyInput processes keyboard input with modal dispatch.
func (m Model) handleKeyInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// overlay UIs get first priority
	if m.dumpForm.IsActive() {
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		return m, m.dumpForm.Update(msg)
	}
	if m.dumpProgress.IsActive() {
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		return m, m.dumpProgress.Update(msg)
	}
	if m.connForm.IsActive() {
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		return m, m.connForm.Update(msg)
	}
	if m.dialog.IsActive() {
		return m, m.dialog.Update(msg)
	}
	if m.connSelector.IsActive() {
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		return m, m.connSelector.Update(msg)
	}
	if m.commandBar.Active() {
		return m, m.commandBar.Update(msg)
	}
	if m.helpOverlay.IsActive() {
		m.helpOverlay.Update(msg)
		m.showHelp = m.helpOverlay.IsActive()
		return m, nil
	}
	if m.bufferList.IsActive() {
		m.bufferList.Update(msg)
		return m, nil
	}
	if m.editDialog.IsActive() {
		return m, m.editDialog.Update(msg)
	}

	// visual mode
	if m.mode.IsVisual() {
		if msg.String() == "esc" {
			m.resultView.ExitVisual()
			m.mode = core.ModeNormal
			m.statusBar.SetMode(m.mode)
			return m, nil
		}
		return m, m.resultView.Update(msg)
	}

	// multi-key buffer switching (gt/gT)
	if m.mode == core.ModeNormal && m.keySeq.Active() {
		first := m.keySeq.Consume()
		if first == "g" {
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
					m.recalcLayout()
					return m, cmd
				}
				return m, nil
			}
		}
	}
	if m.mode == core.ModeNormal && msg.String() == "g" {
		m.keySeq.Start("g")
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

	// global keybindings
	if action := MatchGlobal(msg, m.mode); action != ActionNone {
		return m.handleAction(action)
	}

	// forward to active pane
	active := m.panes.Active()
	if active == nil {
		return m, nil
	}
	var cmd tea.Cmd
	if m.mode == core.ModeNormal {
		_, cmd = active.Update(msg)
	} else if m.mode == core.ModeInsert && m.panes.ActiveID() == pane.QueryEditor {
		cmd = m.queryEditor.Update(msg)
	}
	m.recalcLayout()
	return m, cmd
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

	// update modified indicator based on current editor content vs last executed
	buf := m.buffers.Active()
	buf.Modified = m.queryEditor.Content() != buf.LastExecutedQuery
	m.statusBar.SetBufferModified(buf.Modified)
}

// View renders the full TUI.
func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	if m.dumpForm.IsActive() {
		return m.dumpForm.View(m.width, m.height)
	}
	if m.dumpProgress.IsActive() {
		return m.dumpProgress.View(m.width, m.height)
	}
	if m.connForm.IsActive() {
		return m.connForm.View(m.width, m.height)
	}
	if m.dialog.IsActive() {
		return m.dialog.View(m.width, m.height)
	}
	if m.connSelector.IsActive() {
		return m.connSelector.View(m.width, m.height)
	}

	if m.helpOverlay.IsActive() {
		return m.helpOverlay.View(m.panes.ActiveID(), m.mode, m.width, m.height)
	}

	if m.bufferList.IsActive() {
		return m.bufferList.View(m.width, m.height)
	}

	if m.editDialog.IsActive() {
		return m.editDialog.View(m.width, m.height)
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

// OpenHelp opens the help overlay. Exported for command handler use.
func (m *Model) OpenHelp(topic string) {
	if topic == "" {
		m.helpOverlay.Open()
	} else {
		m.helpOverlay.OpenTopic(topic)
	}
	m.showHelp = true
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
	buf.Modified = buf.Query != buf.LastExecutedQuery

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
	m.statusBar.SetBufferModified(buf.Modified)
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
		if table.IsNull(v) {
			vals[i] = nil
		} else {
			vals[i] = v
		}
	}
	return vals, nil
}

func (it *sliceRowIterator) Err() error { return nil }
func (it *sliceRowIterator) Close()     {}
