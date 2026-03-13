package cmd

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/mdjarv/db/internal/db"
)

var introspectCmd = &cobra.Command{
	Use:   "introspect [table]",
	Short: "Show type introspection details for a table's columns",
	Args:  cobra.ExactArgs(1),
	RunE:  runIntrospect,
}

func init() {
	introspectCmd.Flags().String("schema", "public", "schema name")
	rootCmd.AddCommand(introspectCmd)
}

func runIntrospect(cmd *cobra.Command, args []string) error {
	conn, err := connectFromFlags(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close(cmd.Context()) }()

	ctx := cmd.Context()
	schemaName, _ := cmd.Flags().GetString("schema")
	table := args[0]

	// Query with LIMIT 0 to get column metadata without fetching rows.
	q := fmt.Sprintf("SELECT * FROM %s.%s LIMIT 0", quoteIdent(schemaName), quoteIdent(table))
	result, err := conn.Query(ctx, q)
	if err != nil {
		return err
	}
	result.Rows.Close()

	introspector, ok := conn.(db.TypeIntrospector)
	if !ok {
		return fmt.Errorf("driver does not support type introspection")
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "COLUMN\tTYPE\tOID\tARRAY\tELEM TYPE\tELEM OID\tENUM VALUES\tCOMPOSITE FIELDS"); err != nil {
		return err
	}

	for _, col := range result.Columns {
		d := introspector.TypeDetail(col.TypeOID)

		arrayStr := ""
		elemType := ""
		elemOID := ""
		if d.IsArray {
			arrayStr = "yes"
			elemType = d.ElemTypeName
			elemOID = fmt.Sprintf("%d", d.ElemOID)
		}

		enumStr := ""
		if len(d.EnumValues) > 0 {
			enumStr = strings.Join(d.EnumValues, ", ")
		}

		compStr := ""
		if len(d.CompositeFields) > 0 {
			parts := make([]string, len(d.CompositeFields))
			for i, f := range d.CompositeFields {
				parts[i] = f.Name + " " + f.TypeName
			}
			compStr = strings.Join(parts, ", ")
		}

		if _, err := fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\t%s\t%s\n",
			col.Name, d.Name, d.OID, arrayStr, elemType, elemOID, enumStr, compStr); err != nil {
			return err
		}
	}

	return w.Flush()
}

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}
