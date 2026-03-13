package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/mdjarv/db/internal/db"
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
	queryCmd.Flags().StringVarP(&queryFormat, "format", "F", "table", "output format (table)")
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
	if format != "table" {
		return fmt.Errorf("format %q not yet supported", format)
	}

	var rows [][]any
	for result.Rows.Next() {
		vals, err := result.Rows.Values()
		if err != nil {
			return fmt.Errorf("read row: %w", err)
		}
		row := make([]any, len(vals))
		copy(row, vals)
		rows = append(rows, row)
	}
	result.Rows.Close()
	if err := result.Rows.Err(); err != nil {
		return fmt.Errorf("rows: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if !noHeader {
		for i, col := range result.Columns {
			if i > 0 {
				if _, err := fmt.Fprint(w, "\t"); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprint(w, col.Name); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}

	for _, row := range rows {
		for i, val := range row {
			if i > 0 {
				if _, err := fmt.Fprint(w, "\t"); err != nil {
					return err
				}
			}
			s := "NULL"
			if val != nil {
				s = fmt.Sprintf("%v", val)
			}
			if _, err := fmt.Fprint(w, s); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}

	return w.Flush()
}
