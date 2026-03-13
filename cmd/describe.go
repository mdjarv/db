package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/mdjarv/db/internal/schema"
)

var describeCmd = &cobra.Command{
	Use:   "describe [table]",
	Short: "Describe a table's columns, indexes, constraints, and foreign keys",
	Args:  cobra.ExactArgs(1),
	RunE:  runDescribe,
}

func init() {
	describeCmd.Flags().String("schema", "public", "schema name")
	rootCmd.AddCommand(describeCmd)
}

func runDescribe(cmd *cobra.Command, args []string) error {
	conn, err := connectFromFlags(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close(cmd.Context()) }()

	schemaName, _ := cmd.Flags().GetString("schema")
	table := args[0]
	ctx := cmd.Context()
	insp := schema.NewPostgresInspector(conn)

	cols, err := insp.Columns(ctx, schemaName, table)
	if err != nil {
		return err
	}
	indexes, err := insp.Indexes(ctx, schemaName, table)
	if err != nil {
		return err
	}
	constraints, err := insp.Constraints(ctx, schemaName, table)
	if err != nil {
		return err
	}
	fks, err := insp.ForeignKeys(ctx, schemaName, table)
	if err != nil {
		return err
	}

	w := os.Stdout
	if err := writeColumns(w, cols); err != nil {
		return err
	}
	if err := writeIndexes(w, indexes); err != nil {
		return err
	}
	if err := writeConstraints(w, constraints); err != nil {
		return err
	}
	return writeForeignKeys(w, fks)
}

func writeColumns(out io.Writer, cols []schema.ColumnInfo) error {
	if len(cols) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(out, "Columns"); err != nil {
		return err
	}
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "  #\tNAME\tTYPE\tNULLABLE\tDEFAULT\tPK"); err != nil {
		return err
	}
	for _, c := range cols {
		pk := ""
		if c.IsPK {
			pk = "yes"
		}
		nullable := "no"
		if c.Nullable {
			nullable = "yes"
		}
		if _, err := fmt.Fprintf(w, "  %d\t%s\t%s\t%s\t%s\t%s\n",
			c.Position, c.Name, c.TypeName, nullable, c.Default, pk); err != nil {
			return err
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	_, err := fmt.Fprintln(out)
	return err
}

func writeIndexes(out io.Writer, indexes []schema.Index) error {
	if len(indexes) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(out, "Indexes"); err != nil {
		return err
	}
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "  NAME\tCOLUMNS\tUNIQUE\tTYPE\tSIZE"); err != nil {
		return err
	}
	for _, idx := range indexes {
		unique := ""
		if idx.Unique {
			unique = "yes"
		}
		if _, err := fmt.Fprintf(w, "  %s\t%s\t%s\t%s\t%s\n",
			idx.Name, strings.Join(idx.Columns, ", "), unique, idx.Type, idx.Size); err != nil {
			return err
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	_, err := fmt.Fprintln(out)
	return err
}

func writeConstraints(out io.Writer, constraints []schema.Constraint) error {
	if len(constraints) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(out, "Constraints"); err != nil {
		return err
	}
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "  NAME\tTYPE\tCOLUMNS\tDEFINITION"); err != nil {
		return err
	}
	for _, c := range constraints {
		if _, err := fmt.Fprintf(w, "  %s\t%s\t%s\t%s\n",
			c.Name, c.Type, strings.Join(c.Columns, ", "), c.Definition); err != nil {
			return err
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	_, err := fmt.Fprintln(out)
	return err
}

func writeForeignKeys(out io.Writer, fks []schema.ForeignKey) error {
	if len(fks) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(out, "Foreign Keys"); err != nil {
		return err
	}
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "  NAME\tCOLUMNS\tREFERENCES\tON DELETE\tON UPDATE"); err != nil {
		return err
	}
	for _, fk := range fks {
		ref := fmt.Sprintf("%s.%s(%s)",
			fk.ReferencedSchema, fk.ReferencedTable,
			strings.Join(fk.ReferencedColumns, ", "))
		if _, err := fmt.Fprintf(w, "  %s\t%s\t%s\t%s\t%s\n",
			fk.Name, strings.Join(fk.Columns, ", "), ref, fk.OnDelete, fk.OnUpdate); err != nil {
			return err
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	_, err := fmt.Fprintln(out)
	return err
}
