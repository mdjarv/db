package pane

// ID identifies a pane.
type ID int

// Pane identifiers.
const (
	TableList ID = iota
	QueryEditor
	ResultView
)

// Manager tracks panes and focus state.
type Manager struct {
	panes  map[ID]Pane
	order  []ID
	active int
}

// NewManager creates a pane manager.
func NewManager() *Manager {
	return &Manager{
		panes: make(map[ID]Pane),
		order: []ID{TableList, QueryEditor, ResultView},
	}
}

// Register adds a pane.
func (m *Manager) Register(id ID, p Pane) {
	m.panes[id] = p
}

// Get returns a pane by ID.
func (m *Manager) Get(id ID) Pane {
	return m.panes[id]
}

// Active returns the focused pane.
func (m *Manager) Active() Pane {
	if len(m.order) == 0 {
		return nil
	}
	return m.panes[m.order[m.active]]
}

// ActiveID returns the focused pane's ID.
func (m *Manager) ActiveID() ID {
	return m.order[m.active]
}

// SetActive focuses a specific pane.
func (m *Manager) SetActive(id ID) {
	for i, oid := range m.order {
		if oid == id {
			if old := m.Active(); old != nil {
				old.SetFocused(false)
			}
			m.active = i
			if p := m.Active(); p != nil {
				p.SetFocused(true)
			}
			return
		}
	}
}

// CycleForward moves focus to the next pane.
func (m *Manager) CycleForward() {
	if old := m.Active(); old != nil {
		old.SetFocused(false)
	}
	m.active = (m.active + 1) % len(m.order)
	if p := m.Active(); p != nil {
		p.SetFocused(true)
	}
}

// CycleBackward moves focus to the previous pane.
func (m *Manager) CycleBackward() {
	if old := m.Active(); old != nil {
		old.SetFocused(false)
	}
	m.active = (m.active - 1 + len(m.order)) % len(m.order)
	if p := m.Active(); p != nil {
		p.SetFocused(true)
	}
}

// FocusLeft moves focus to the left pane.
func (m *Manager) FocusLeft() {
	id := m.ActiveID()
	if id == QueryEditor || id == ResultView {
		m.SetActive(TableList)
	}
}

// FocusRight moves focus to the right pane.
func (m *Manager) FocusRight() {
	if m.ActiveID() == TableList {
		m.SetActive(ResultView)
	}
}

// FocusUp moves focus up within the right column.
func (m *Manager) FocusUp() {
	if m.ActiveID() == ResultView {
		m.SetActive(QueryEditor)
	}
}

// FocusDown moves focus down within the right column.
func (m *Manager) FocusDown() {
	if m.ActiveID() == QueryEditor {
		m.SetActive(ResultView)
	}
}

// FocusByNumber focuses pane n (1-indexed).
func (m *Manager) FocusByNumber(n int) {
	idx := n - 1
	if idx >= 0 && idx < len(m.order) {
		m.SetActive(m.order[idx])
	}
}

// All returns all registered panes.
func (m *Manager) All() map[ID]Pane {
	return m.panes
}
