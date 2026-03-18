// Package cmd defines CLI commands.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:           "db",
	Short:         "TUI database client for PostgreSQL",
	Long:          "A vim-modal TUI for PostgreSQL, built for backend developers who want fast schema browsing, querying, and data editing without leaving the terminal.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(version)
		},
	})

	rootCmd.Version = version
}

// Execute runs the root command.
func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
	return err
}
