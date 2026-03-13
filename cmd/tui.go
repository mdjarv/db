package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/mdjarv/db/internal/db"
	"github.com/mdjarv/db/internal/schema"
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

func runTUI(cmd *cobra.Command, _ []string) error {
	cfg, err := resolveConnection(cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connection error: %v\n", err)
		return err
	}
	conn, err := db.Open(cmd.Context(), "postgres", cfg.DSN())
	if err != nil {
		fmt.Fprintf(os.Stderr, "connection error: %v\n", err)
		return err
	}
	defer func() { _ = conn.Close(cmd.Context()) }()

	connInfo := fmt.Sprintf("%s@%s/%s", cfg.User, cfg.Host, cfg.DBName)
	insp := schema.NewPostgresInspector(conn)
	m := app.NewWithConn(conn, insp, connInfo)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return err
	}
	return nil
}
