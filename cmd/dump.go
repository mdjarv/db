package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/mdjarv/db/internal/dump"
)

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump database using pg_dump",
	Long:  "Export a PostgreSQL database using pg_dump with progress reporting.",
	RunE:  runDump,
}

var (
	dumpSchemaOnly bool
	dumpTables     []string
	dumpFormat     string
	dumpOutput     string
	dumpVerbose    bool
	dumpTimeout    int
)

func init() {
	dumpCmd.Flags().BoolVarP(&dumpSchemaOnly, "schema-only", "s", false, "dump DDL only, no data")
	dumpCmd.Flags().StringArrayVarP(&dumpTables, "table", "t", nil, "table to dump (repeatable)")
	dumpCmd.Flags().StringVarP(&dumpFormat, "format", "F", "custom", "output format (plain, custom, directory, tar)")
	dumpCmd.Flags().StringVarP(&dumpOutput, "output", "o", "", "output file path")
	dumpCmd.Flags().BoolVarP(&dumpVerbose, "verbose", "v", false, "print raw pg_dump output instead of progress bar")
	dumpCmd.Flags().IntVar(&dumpTimeout, "timeout", 0, "timeout in seconds (0 = no timeout)")
	rootCmd.AddCommand(dumpCmd)
}

func runDump(cmd *cobra.Command, _ []string) error {
	cfg, err := resolveConnection(cmd)
	if err != nil {
		return classifyConnError(err)
	}

	format, err := dump.ParseFormat(dumpFormat)
	if err != nil {
		return wrapIO("parse format", err)
	}

	binary, err := dump.FindPgDump("")
	if err != nil {
		return wrapIO("find pg_dump", err)
	}

	outputPath := dumpOutput
	if outputPath == "" {
		outputPath = dump.DefaultOutputPath(cfg.DBName, format)
	}

	dcfg := dump.Config{
		Host:       cfg.Host,
		Port:       fmt.Sprintf("%d", cfg.Port),
		User:       cfg.User,
		Password:   cfg.Password,
		DBName:     cfg.DBName,
		SSLMode:    cfg.SSLMode,
		Format:     format,
		SchemaOnly: dumpSchemaOnly,
		Tables:     dumpTables,
		OutputPath: outputPath,
	}

	ctx := context.Background()
	var cancel context.CancelFunc
	if dumpTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(dumpTimeout)*time.Second)
		defer cancel()
	}

	runner := dump.NewRunner(binary)
	ch, err := runner.Run(ctx, dcfg, 0)
	if err != nil {
		return wrapIO("start pg_dump", err)
	}

	tableCount := 0
	for ev := range ch {
		if ev.Err != nil {
			return wrapQuery("pg_dump", ev.Err)
		}
		if ev.Done {
			break
		}
		if dumpVerbose {
			fmt.Fprintln(os.Stderr, ev.Object)
		} else if ev.Index > 0 {
			tableCount = ev.Index
			fmt.Fprintf(os.Stderr, "\rDumping: %s [%d tables]", ev.Object, tableCount)
		}
	}

	if !dumpVerbose && tableCount > 0 {
		fmt.Fprintln(os.Stderr)
	}

	info, err := os.Stat(outputPath)
	if err == nil {
		fmt.Fprintf(os.Stderr, "Dumped to %s (%s)\n", outputPath, formatSize(info.Size()))
	} else {
		fmt.Fprintf(os.Stderr, "Dumped to %s\n", outputPath)
	}

	return nil
}

func formatSize(bytes int64) string {
	switch {
	case bytes >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(1<<30))
	case bytes >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(1<<20))
	case bytes >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
