package theme

// DefaultDark returns the default dark theme based on Catppuccin Mocha.
func DefaultDark() *Theme {
	t := &Theme{
		Name: "default-dark",
		Colors: Colors{
			Chrome: ChromeColors{
				Border:         "#585b70", // surface2
				BorderFocused:  "#89b4fa", // blue
				BorderVisual:   "#fab387", // peach
				StatusBarBG:    "#1e1e2e", // base
				StatusBarFG:    "#cdd6f4", // text
				ModeNormalBG:   "#89b4fa", // blue
				ModeNormalFG:   "#1e1e2e", // base
				ModeInsertBG:   "#a6e3a1", // green
				ModeInsertFG:   "#1e1e2e", // base
				ModeCommandBG:  "#fab387", // peach
				ModeCommandFG:  "#1e1e2e", // base
				ConnectedFG:    "#a6e3a1", // green
				DisconnectedFG: "#f38ba8", // red
				TxFG:           "#6c7086", // overlay0
				PromptFG:       "#fab387", // peach
			},
			Syntax: SyntaxColors{
				Keyword:  "#cba6f7", // mauve
				String:   "#a6e3a1", // green
				Number:   "#fab387", // peach
				Comment:  "#6c7086", // overlay0
				Type:     "#89dceb", // sky
				Function: "#f9e2af", // yellow
				Operator: "#94e2d5", // teal
			},
			Data: DataColors{
				Null:      "#6c7086", // overlay0
				Boolean:   "#89dceb", // sky
				BoolTrue:  "#a6e3a1", // green
				BoolFalse: "#f38ba8", // red
				Number:    "#fab387", // peach
				String:    "#a6e3a1", // green
				Date:      "#94e2d5", // teal
				UUID:      "#b4befe", // lavender
			},
			UI: UIColors{
				Cursor:         "#89b4fa", // blue
				CursorFG:       "#1e1e2e", // base
				CursorRow:      "#cdd6f4", // text
				Selection:      "#f5c2e7", // pink
				SelectionFG:    "#1e1e2e", // base
				ColSelection:   "#313244", // surface0
				ColSelectionFG: "#f9e2af", // yellow
				Dim:            "#585b70", // surface2
				Header:         "#89b4fa", // blue
				Separator:      "#45475a", // surface1
				Error:          "#f38ba8", // red
				Warning:        "#fab387", // peach
				Success:        "#a6e3a1", // green
				InsertCursor:   "#94e2d5", // teal
				EditorSelect:   "#45475a", // surface1
				EditorSelectFG: "#cdd6f4", // text
				Gutter:         "#585b70", // surface2
				Modified:       "#f9e2af", // yellow
				ModifiedFG:     "#1e1e2e", // base
				Deleted:        "#585b70", // surface2
			},
		},
	}
	return t.Build()
}
