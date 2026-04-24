package cmd

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var introspectCmd = &cobra.Command{
	Use:   "introspect [table]",
	Short: "Show driver-level type details for a table's columns",
	Args:  cobra.ExactArgs(1),
	RunE:  runIntrospect,
}

func init() {
	introspectCmd.Flags().String("schema", "", "schema name (driver default if empty)")
	rootCmd.AddCommand(introspectCmd)
}

func runIntrospect(cmd *cobra.Command, args []string) error {
	conn, err := connectFromFlags(cmd)
	if err != nil {
		return err // already wrapped
	}
	defer func() { _ = conn.Close(cmd.Context()) }()

	ctx := cmd.Context()
	schemaName, _ := cmd.Flags().GetString("schema")
	table := args[0]

	// LIMIT 0 probes column metadata without transferring rows.
	dialect := conn.Dialect()
	qualified := dialect.QualifyTable(schemaName, table)
	q := fmt.Sprintf("SELECT * FROM %s LIMIT 0", qualified)
	result, err := conn.Query(ctx, q)
	if err != nil {
		return wrapQuery("introspect", err)
	}
	result.Rows.Close()

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "COLUMN\tTYPE\tARRAY\tELEM TYPE\tENUM VALUES\tCOMPOSITE FIELDS"); err != nil {
		return err
	}

	for _, col := range result.Columns {
		arrayStr := ""
		elemType := ""
		if col.IsArray() {
			arrayStr = "yes"
			elemType = col.ElemTypeName()
		}

		enumStr := ""
		if len(col.EnumValues) > 0 {
			enumStr = strings.Join(col.EnumValues, ", ")
		}

		compStr := ""
		if len(col.CompositeFields) > 0 {
			parts := make([]string, len(col.CompositeFields))
			for i, f := range col.CompositeFields {
				parts[i] = f.Name + " " + f.TypeName
			}
			compStr = strings.Join(parts, ", ")
		}

		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			col.Name, col.TypeName, arrayStr, elemType, enumStr, compStr); err != nil {
			return err
		}
	}

	return w.Flush()
}
