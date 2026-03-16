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
				Null:    "#6c7086", // overlay0
				Boolean: "#89dceb", // sky
				Number:  "#fab387", // peach
				String:  "#a6e3a1", // green
				Date:    "#f9e2af", // yellow
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

// DefaultLight returns the default light theme based on Catppuccin Latte.
func DefaultLight() *Theme {
	t := &Theme{
		Name: "default-light",
		Colors: Colors{
			Chrome: ChromeColors{
				Border:         "#9ca0b0", // overlay0
				BorderFocused:  "#1e66f5", // blue
				BorderVisual:   "#fe640b", // peach
				StatusBarBG:    "#eff1f5", // base
				StatusBarFG:    "#4c4f69", // text
				ModeNormalBG:   "#1e66f5", // blue
				ModeNormalFG:   "#eff1f5", // base
				ModeInsertBG:   "#40a02b", // green
				ModeInsertFG:   "#eff1f5", // base
				ModeCommandBG:  "#fe640b", // peach
				ModeCommandFG:  "#eff1f5", // base
				ConnectedFG:    "#40a02b", // green
				DisconnectedFG: "#d20f39", // red
				TxFG:           "#8c8fa1", // overlay1
				PromptFG:       "#fe640b", // peach
			},
			Syntax: SyntaxColors{
				Keyword:  "#8839ef", // mauve
				String:   "#40a02b", // green
				Number:   "#fe640b", // peach
				Comment:  "#9ca0b0", // overlay0
				Type:     "#04a5e5", // sky
				Function: "#df8e1d", // yellow
				Operator: "#179299", // teal
			},
			Data: DataColors{
				Null:    "#9ca0b0", // overlay0
				Boolean: "#04a5e5", // sky
				Number:  "#fe640b", // peach
				String:  "#40a02b", // green
				Date:    "#df8e1d", // yellow
			},
			UI: UIColors{
				Cursor:         "#1e66f5", // blue
				CursorFG:       "#eff1f5", // base
				CursorRow:      "#4c4f69", // text
				Selection:      "#ea76cb", // pink
				SelectionFG:    "#eff1f5", // base
				ColSelection:   "#e6e9ef", // mantle
				ColSelectionFG: "#df8e1d", // yellow
				Dim:            "#9ca0b0", // overlay0
				Header:         "#1e66f5", // blue
				Separator:      "#bcc0cc", // surface1
				Error:          "#d20f39", // red
				Warning:        "#fe640b", // peach
				Success:        "#40a02b", // green
				InsertCursor:   "#179299", // teal
				EditorSelect:   "#ccd0da", // surface0
				EditorSelectFG: "#4c4f69", // text
				Gutter:         "#9ca0b0", // overlay0
				Modified:       "#df8e1d", // yellow
				ModifiedFG:     "#eff1f5", // base
				Deleted:        "#9ca0b0", // overlay0
			},
		},
	}
	return t.Build()
}

// SolarizedDark returns the Solarized Dark theme.
func SolarizedDark() *Theme {
	t := &Theme{
		Name: "solarized-dark",
		Colors: Colors{
			Chrome: ChromeColors{
				Border:         "#586e75",
				BorderFocused:  "#268bd2",
				BorderVisual:   "#cb4b16",
				StatusBarBG:    "#073642",
				StatusBarFG:    "#839496",
				ModeNormalBG:   "#268bd2",
				ModeNormalFG:   "#fdf6e3",
				ModeInsertBG:   "#859900",
				ModeInsertFG:   "#fdf6e3",
				ModeCommandBG:  "#b58900",
				ModeCommandFG:  "#fdf6e3",
				ConnectedFG:    "#2aa198",
				DisconnectedFG: "#dc322f",
				TxFG:           "#657b83",
				PromptFG:       "#b58900",
			},
			Syntax: SyntaxColors{
				Keyword:  "#268bd2",
				String:   "#2aa198",
				Number:   "#d33682",
				Comment:  "#586e75",
				Type:     "#b58900",
				Function: "#859900",
				Operator: "#657b83",
			},
			Data: DataColors{
				Null:    "#586e75",
				Boolean: "#2aa198",
				Number:  "#d33682",
				String:  "#2aa198",
				Date:    "#b58900",
			},
			UI: UIColors{
				Cursor:         "#268bd2",
				CursorFG:       "#fdf6e3",
				CursorRow:      "#93a1a1",
				Selection:      "#cb4b16",
				SelectionFG:    "#fdf6e3",
				ColSelection:   "#073642",
				ColSelectionFG: "#b58900",
				Dim:            "#586e75",
				Header:         "#268bd2",
				Separator:      "#586e75",
				Error:          "#dc322f",
				Warning:        "#cb4b16",
				Success:        "#859900",
				InsertCursor:   "#2aa198",
				EditorSelect:   "#073642",
				EditorSelectFG: "#93a1a1",
				Gutter:         "#586e75",
				Modified:       "#b58900",
				ModifiedFG:     "#002b36",
				Deleted:        "#586e75",
			},
		},
	}
	return t.Build()
}

// SolarizedLight returns the Solarized Light theme.
func SolarizedLight() *Theme {
	t := &Theme{
		Name: "solarized-light",
		Colors: Colors{
			Chrome: ChromeColors{
				Border:         "#93a1a1",
				BorderFocused:  "#268bd2",
				BorderVisual:   "#cb4b16",
				StatusBarBG:    "#eee8d5",
				StatusBarFG:    "#657b83",
				ModeNormalBG:   "#268bd2",
				ModeNormalFG:   "#fdf6e3",
				ModeInsertBG:   "#859900",
				ModeInsertFG:   "#fdf6e3",
				ModeCommandBG:  "#b58900",
				ModeCommandFG:  "#fdf6e3",
				ConnectedFG:    "#2aa198",
				DisconnectedFG: "#dc322f",
				TxFG:           "#93a1a1",
				PromptFG:       "#b58900",
			},
			Syntax: SyntaxColors{
				Keyword:  "#268bd2",
				String:   "#2aa198",
				Number:   "#d33682",
				Comment:  "#93a1a1",
				Type:     "#b58900",
				Function: "#859900",
				Operator: "#586e75",
			},
			Data: DataColors{
				Null:    "#93a1a1",
				Boolean: "#2aa198",
				Number:  "#d33682",
				String:  "#2aa198",
				Date:    "#b58900",
			},
			UI: UIColors{
				Cursor:         "#268bd2",
				CursorFG:       "#fdf6e3",
				CursorRow:      "#073642",
				Selection:      "#cb4b16",
				SelectionFG:    "#fdf6e3",
				ColSelection:   "#eee8d5",
				ColSelectionFG: "#b58900",
				Dim:            "#93a1a1",
				Header:         "#268bd2",
				Separator:      "#93a1a1",
				Error:          "#dc322f",
				Warning:        "#cb4b16",
				Success:        "#859900",
				InsertCursor:   "#2aa198",
				EditorSelect:   "#eee8d5",
				EditorSelectFG: "#586e75",
				Gutter:         "#93a1a1",
				Modified:       "#b58900",
				ModifiedFG:     "#fdf6e3",
				Deleted:        "#93a1a1",
			},
		},
	}
	return t.Build()
}

// Nord returns the Nord theme.
func Nord() *Theme {
	t := &Theme{
		Name: "nord",
		Colors: Colors{
			Chrome: ChromeColors{
				Border:         "#4c566a",
				BorderFocused:  "#88c0d0",
				BorderVisual:   "#d08770",
				StatusBarBG:    "#3b4252",
				StatusBarFG:    "#d8dee9",
				ModeNormalBG:   "#5e81ac",
				ModeNormalFG:   "#eceff4",
				ModeInsertBG:   "#a3be8c",
				ModeInsertFG:   "#2e3440",
				ModeCommandBG:  "#ebcb8b",
				ModeCommandFG:  "#2e3440",
				ConnectedFG:    "#a3be8c",
				DisconnectedFG: "#bf616a",
				TxFG:           "#4c566a",
				PromptFG:       "#ebcb8b",
			},
			Syntax: SyntaxColors{
				Keyword:  "#81a1c1",
				String:   "#a3be8c",
				Number:   "#b48ead",
				Comment:  "#4c566a",
				Type:     "#8fbcbb",
				Function: "#88c0d0",
				Operator: "#81a1c1",
			},
			Data: DataColors{
				Null:    "#4c566a",
				Boolean: "#8fbcbb",
				Number:  "#b48ead",
				String:  "#a3be8c",
				Date:    "#ebcb8b",
			},
			UI: UIColors{
				Cursor:         "#88c0d0",
				CursorFG:       "#2e3440",
				CursorRow:      "#eceff4",
				Selection:      "#d08770",
				SelectionFG:    "#2e3440",
				ColSelection:   "#434c5e",
				ColSelectionFG: "#ebcb8b",
				Dim:            "#4c566a",
				Header:         "#88c0d0",
				Separator:      "#4c566a",
				Error:          "#bf616a",
				Warning:        "#d08770",
				Success:        "#a3be8c",
				InsertCursor:   "#88c0d0",
				EditorSelect:   "#434c5e",
				EditorSelectFG: "#d8dee9",
				Gutter:         "#4c566a",
				Modified:       "#ebcb8b",
				ModifiedFG:     "#2e3440",
				Deleted:        "#4c566a",
			},
		},
	}
	return t.Build()
}

// Dracula returns the Dracula theme.
func Dracula() *Theme {
	t := &Theme{
		Name: "dracula",
		Colors: Colors{
			Chrome: ChromeColors{
				Border:         "#6272a4",
				BorderFocused:  "#bd93f9",
				BorderVisual:   "#ffb86c",
				StatusBarBG:    "#44475a",
				StatusBarFG:    "#f8f8f2",
				ModeNormalBG:   "#bd93f9",
				ModeNormalFG:   "#282a36",
				ModeInsertBG:   "#50fa7b",
				ModeInsertFG:   "#282a36",
				ModeCommandBG:  "#f1fa8c",
				ModeCommandFG:  "#282a36",
				ConnectedFG:    "#50fa7b",
				DisconnectedFG: "#ff5555",
				TxFG:           "#6272a4",
				PromptFG:       "#f1fa8c",
			},
			Syntax: SyntaxColors{
				Keyword:  "#ff79c6",
				String:   "#f1fa8c",
				Number:   "#bd93f9",
				Comment:  "#6272a4",
				Type:     "#8be9fd",
				Function: "#50fa7b",
				Operator: "#ff79c6",
			},
			Data: DataColors{
				Null:    "#6272a4",
				Boolean: "#8be9fd",
				Number:  "#bd93f9",
				String:  "#f1fa8c",
				Date:    "#ffb86c",
			},
			UI: UIColors{
				Cursor:         "#bd93f9",
				CursorFG:       "#282a36",
				CursorRow:      "#f8f8f2",
				Selection:      "#ffb86c",
				SelectionFG:    "#282a36",
				ColSelection:   "#44475a",
				ColSelectionFG: "#f1fa8c",
				Dim:            "#6272a4",
				Header:         "#bd93f9",
				Separator:      "#6272a4",
				Error:          "#ff5555",
				Warning:        "#ffb86c",
				Success:        "#50fa7b",
				InsertCursor:   "#8be9fd",
				EditorSelect:   "#44475a",
				EditorSelectFG: "#f8f8f2",
				Gutter:         "#6272a4",
				Modified:       "#f1fa8c",
				ModifiedFG:     "#282a36",
				Deleted:        "#6272a4",
			},
		},
	}
	return t.Build()
}
