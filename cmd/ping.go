package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/mdjarv/db/internal/db"
)

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Test database connection",
	RunE:  runPing,
}

func init() {
	rootCmd.AddCommand(pingCmd)
}

func runPing(cmd *cobra.Command, _ []string) error {
	cfg, err := resolveConnection(cmd)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := time.Now()
	c, err := db.Open(ctx, "postgres", cfg.DSN())
	if err != nil {
		return fmt.Errorf("connect failed: %w", err)
	}
	defer func() { _ = c.Close(ctx) }()

	if _, err := c.Query(ctx, "SELECT 1"); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	fmt.Printf("OK (%s)\n", time.Since(start).Round(time.Millisecond))
	return nil
}
