package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/mdjarv/db/internal/schema"
)

var tablesCmd = &cobra.Command{
	Use:   "tables",
	Short: "List tables in a schema",
	RunE:  runTables,
}

func init() {
	tablesCmd.Flags().String("schema", "public", "schema name")
	rootCmd.AddCommand(tablesCmd)
}

func runTables(cmd *cobra.Command, _ []string) error {
	conn, err := connectFromFlags(cmd)
	if err != nil {
		return err // already wrapped
	}
	defer func() { _ = conn.Close(cmd.Context()) }()

	schemaName, _ := cmd.Flags().GetString("schema")
	insp := schema.NewPostgresInspector(conn)

	tables, err := insp.Tables(cmd.Context(), schemaName)
	if err != nil {
		return wrapQuery("list tables", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "NAME\tTYPE\tROWS\tSIZE"); err != nil {
		return err
	}
	for _, t := range tables {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", t.Name, t.Type, t.RowEstimate, t.Size); err != nil {
			return err
		}
	}
	return w.Flush()
}
