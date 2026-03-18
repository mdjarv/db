package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/mdjarv/db/internal/config"
	"github.com/mdjarv/db/internal/conn"
	"github.com/mdjarv/db/internal/db"
	"github.com/mdjarv/db/internal/schema"
	"github.com/mdjarv/db/internal/tui/app"
	"github.com/mdjarv/db/internal/tui/theme"
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
	applyTheme(cmd)

	stores := connectionStores()
	creds := conn.NewCredentialStore(conn.OSKeyring{})
	gitRoot := config.GitRoot()

	cfg, err := resolveConnection(cmd)

	var m app.Model
	if err == nil {
		c, err := db.Open(cmd.Context(), "postgres", cfg.DSN())
		if err != nil {
			return classifyConnError(err)
		}
		connInfo := fmt.Sprintf("%s@%s/%s", cfg.User, cfg.Host, cfg.DBName)
		insp := schema.NewPostgresInspector(c)
		m = app.NewWithOpts(app.Options{
			Conn: c, Inspector: insp, ConnInfo: connInfo,
			Stores: stores, Creds: creds, GitRoot: gitRoot,
		})
	} else {
		m = app.NewWithOpts(app.Options{
			Stores: stores, Creds: creds, GitRoot: gitRoot,
		})
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if final, ok := result.(app.Model); ok {
		final.Cleanup()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return err
	}
	return nil
}

func applyTheme(cmd *cobra.Command) {
	// CLI flag takes priority
	name, _ := cmd.Flags().GetString("theme")
	if name == "" {
		// fallback to config file
		appCfg, cfgErr := config.Load()
		if cfgErr != nil {
			fmt.Fprintf(os.Stderr, "warning: %v\n", cfgErr)
		}
		name = appCfg.Theme
	}
	if name == "" {
		return // keep default
	}
	t, err := theme.Resolve(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: unknown theme %q, using default\n", name)
		return
	}
	theme.Set(t)
}
