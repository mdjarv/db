package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/mdjarv/db/internal/tui/app"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch TUI interface",
	RunE:  runTUI,
}

func init() {
	rootCmd.AddCommand(tuiCmd)
	rootCmd.RunE = runTUI
}

func runTUI(_ *cobra.Command, _ []string) error {
	m := app.New()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return err
	}
	return nil
}
