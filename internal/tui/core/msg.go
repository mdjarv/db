package core

// ModeChangedMsg signals a vim mode transition.
type ModeChangedMsg struct {
	Mode Mode
}

// StatusMsg carries a status bar message.
type StatusMsg struct {
	Text string
}

// QuerySubmittedMsg carries a submitted SQL query.
type QuerySubmittedMsg struct {
	SQL string
}

// YankMsg carries content to copy to clipboard.
type YankMsg struct {
	Content string
}
