package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/mdjarv/db/internal/db"
	"github.com/mdjarv/db/internal/export"
	"github.com/mdjarv/db/internal/query"
)

var queryCmd = &cobra.Command{
	Use:   "query [SQL]",
	Short: "Execute a SQL query",
	Long:  "Execute SQL from argument, file, or stdin and print results.",
	RunE:  runQuery,
}

var (
	queryFile     string
	queryFormat   string
	queryNoHeader bool
)

func init() {
	queryCmd.Flags().StringVarP(&queryFile, "file", "f", "", "read SQL from file")
	queryCmd.Flags().StringVarP(&queryFormat, "format", "F", "table", "output format (table, csv, json, sql)")
	queryCmd.Flags().BoolVar(&queryNoHeader, "no-header", false, "suppress column headers")
	rootCmd.AddCommand(queryCmd)
}

func runQuery(cmd *cobra.Command, args []string) error {
	sql, err := resolveSQL(args)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := connectFromFlags(cmd)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer func() { _ = conn.Close(ctx) }()

	e := query.NewExecutor(conn, query.AutoCommit)
	res, err := e.Execute(ctx, sql)
	if err != nil {
		return fmt.Errorf("execute: %w", err)
	}

	if res.IsQuery {
		if err := printResult(res.Result, queryFormat, queryNoHeader); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "(%s)\n", res.Duration.Truncate(time.Millisecond))
	} else {
		fmt.Fprintf(os.Stderr, "%d row(s) affected (%s)\n",
			res.ExecResult.RowsAffected,
			res.Duration.Truncate(time.Millisecond))
	}
	return nil
}

func resolveSQL(args []string) (string, error) {
	if queryFile != "" {
		data, err := os.ReadFile(queryFile)
		if err != nil {
			return "", fmt.Errorf("read file: %w", err)
		}
		return string(data), nil
	}
	if len(args) > 0 {
		return args[0], nil
	}
	// Check stdin
	info, err := os.Stdin.Stat()
	if err != nil {
		return "", fmt.Errorf("stat stdin: %w", err)
	}
	if (info.Mode() & os.ModeCharDevice) == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("read stdin: %w", err)
		}
		return string(data), nil
	}
	return "", fmt.Errorf("no SQL provided (use argument, -f file, or pipe to stdin)")
}

func printResult(result *db.Result, format string, noHeader bool) error {
	defer result.Rows.Close()

	var efmt export.Format
	switch format {
	case "table":
		efmt = export.FormatTable
	case "csv":
		efmt = export.FormatCSV
	case "json":
		efmt = export.FormatJSON
	case "sql":
		efmt = export.FormatSQL
	default:
		return fmt.Errorf("unknown format: %s (table, csv, json, sql)", format)
	}

	exp := export.NewExporter(efmt, export.Options{
		NoHeader:    noHeader,
		NullString:  "NULL",
		MaxColWidth: 50,
	})
	return exp.Export(os.Stdout, result)
}
