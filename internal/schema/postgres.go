package schema

import (
	"context"
	"fmt"
	"strings"

	"github.com/mdjarv/db/internal/db"
)

// pgInspector implements Inspector for PostgreSQL.
type pgInspector struct {
	conn db.Conn
}

// NewPostgresInspector returns an Inspector backed by a PostgreSQL connection.
func NewPostgresInspector(conn db.Conn) Inspector {
	return &pgInspector{conn: conn}
}

// resolveSchema returns the target namespace, defaulting to "public" when empty.
func (p *pgInspector) resolveSchema(s string) string {
	if s == "" {
		return "public"
	}
	return s
}

func (p *pgInspector) Tables(ctx context.Context, schemaName string) ([]Table, error) {
	schema := p.resolveSchema(schemaName)
	const q = `
SELECT
	t.table_name,
	t.table_schema,
	CASE t.table_type
		WHEN 'BASE TABLE' THEN 'table'
		WHEN 'VIEW' THEN 'view'
		ELSE lower(t.table_type)
	END AS table_type,
	COALESCE(s.n_live_tup, 0) AS row_estimate,
	COALESCE(pg_size_pretty(pg_total_relation_size(quote_ident(t.table_schema) || '.' || quote_ident(t.table_name))), '') AS size
FROM information_schema.tables t
LEFT JOIN pg_stat_user_tables s
	ON s.schemaname = t.table_schema AND s.relname = t.table_name
WHERE t.table_schema = $1
UNION ALL
SELECT
	c.relname AS table_name,
	n.nspname AS table_schema,
	'materialized view' AS table_type,
	c.reltuples::bigint AS row_estimate,
	pg_size_pretty(pg_total_relation_size(c.oid)) AS size
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE c.relkind = 'm' AND n.nspname = $1
ORDER BY table_type, table_name`

	result, err := p.conn.Query(ctx, q, schema)
	if err != nil {
		return nil, fmt.Errorf("schema: tables: %w", err)
	}
	defer result.Rows.Close()

	var tables []Table
	for result.Rows.Next() {
		vals, err := result.Rows.Values()
		if err != nil {
			return nil, fmt.Errorf("schema: tables row: %w", err)
		}
		tables = append(tables, Table{
			Name:        asString(vals[0]),
			Schema:      asString(vals[1]),
			Type:        asString(vals[2]),
			RowEstimate: asInt64(vals[3]),
			Size:        asString(vals[4]),
		})
	}
	if err := result.Rows.Err(); err != nil {
		return nil, fmt.Errorf("schema: tables iter: %w", err)
	}
	return tables, nil
}

func (p *pgInspector) Columns(ctx context.Context, schemaName, table string) ([]ColumnInfo, error) {
	schema := p.resolveSchema(schemaName)
	const q = `
SELECT
	c.column_name,
	CASE
		WHEN c.data_type = 'ARRAY' THEN substring(c.udt_name from 2) || '[]'
		WHEN c.data_type = 'USER-DEFINED' THEN c.udt_name
		ELSE c.data_type
	END AS data_type,
	CASE WHEN c.is_nullable = 'YES' THEN true ELSE false END AS nullable,
	COALESCE(c.column_default, '') AS col_default,
	COALESCE(pk.is_pk, false) AS is_pk,
	c.ordinal_position
FROM information_schema.columns c
LEFT JOIN (
	SELECT kcu.column_name, true AS is_pk
	FROM information_schema.table_constraints tc
	JOIN information_schema.key_column_usage kcu
		ON tc.constraint_name = kcu.constraint_name
		AND tc.table_schema = kcu.table_schema
	WHERE tc.constraint_type = 'PRIMARY KEY'
		AND tc.table_schema = $1
		AND tc.table_name = $2
) pk ON pk.column_name = c.column_name
WHERE c.table_schema = $1 AND c.table_name = $2
ORDER BY c.ordinal_position`

	result, err := p.conn.Query(ctx, q, schema, table)
	if err != nil {
		return nil, fmt.Errorf("schema: columns: %w", err)
	}
	defer result.Rows.Close()

	var cols []ColumnInfo
	for result.Rows.Next() {
		vals, err := result.Rows.Values()
		if err != nil {
			return nil, fmt.Errorf("schema: columns row: %w", err)
		}
		cols = append(cols, ColumnInfo{
			Name:     asString(vals[0]),
			TypeName: asString(vals[1]),
			Nullable: asBool(vals[2]),
			Default:  asString(vals[3]),
			IsPK:     asBool(vals[4]),
			Position: int(asInt64(vals[5])),
		})
	}
	if err := result.Rows.Err(); err != nil {
		return nil, fmt.Errorf("schema: columns iter: %w", err)
	}
	return cols, nil
}

func (p *pgInspector) Indexes(ctx context.Context, schemaName, table string) ([]Index, error) {
	schema := p.resolveSchema(schemaName)
	const q = `
SELECT
	i.indexname,
	i.indexdef,
	ix.indisunique,
	am.amname,
	COALESCE(pg_size_pretty(pg_relation_size(quote_ident(i.schemaname) || '.' || quote_ident(i.indexname))), '') AS size
FROM pg_indexes i
JOIN pg_class c ON c.relname = i.indexname
JOIN pg_namespace n ON n.oid = c.relnamespace AND n.nspname = i.schemaname
JOIN pg_index ix ON ix.indexrelid = c.oid
JOIN pg_am am ON am.oid = c.relam
WHERE i.schemaname = $1 AND i.tablename = $2
ORDER BY i.indexname`

	result, err := p.conn.Query(ctx, q, schema, table)
	if err != nil {
		return nil, fmt.Errorf("schema: indexes: %w", err)
	}
	defer result.Rows.Close()

	var indexes []Index
	for result.Rows.Next() {
		vals, err := result.Rows.Values()
		if err != nil {
			return nil, fmt.Errorf("schema: indexes row: %w", err)
		}
		def := asString(vals[1])
		indexes = append(indexes, Index{
			Name:       asString(vals[0]),
			Definition: def,
			Unique:     asBool(vals[2]),
			Type:       asString(vals[3]),
			Size:       asString(vals[4]),
			Columns:    parseIndexColumns(def),
		})
	}
	if err := result.Rows.Err(); err != nil {
		return nil, fmt.Errorf("schema: indexes iter: %w", err)
	}
	return indexes, nil
}

func (p *pgInspector) Constraints(ctx context.Context, schemaName, table string) ([]Constraint, error) {
	schema := p.resolveSchema(schemaName)
	const q = `
SELECT
	con.conname,
	CASE con.contype
		WHEN 'p' THEN 'PRIMARY KEY'
		WHEN 'f' THEN 'FOREIGN KEY'
		WHEN 'u' THEN 'UNIQUE'
		WHEN 'c' THEN 'CHECK'
		WHEN 'x' THEN 'EXCLUDE'
		ELSE con.contype::text
	END AS constraint_type,
	pg_get_constraintdef(con.oid) AS definition,
	array_to_string(ARRAY(
		SELECT a.attname
		FROM unnest(con.conkey) WITH ORDINALITY AS k(attnum, ord)
		JOIN pg_attribute a ON a.attrelid = con.conrelid AND a.attnum = k.attnum
		ORDER BY k.ord
	), ',') AS columns
FROM pg_constraint con
JOIN pg_class c ON c.oid = con.conrelid
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = $1 AND c.relname = $2
ORDER BY
	CASE con.contype WHEN 'p' THEN 0 WHEN 'u' THEN 1 WHEN 'f' THEN 2 WHEN 'c' THEN 3 ELSE 4 END,
	con.conname`

	result, err := p.conn.Query(ctx, q, schema, table)
	if err != nil {
		return nil, fmt.Errorf("schema: constraints: %w", err)
	}
	defer result.Rows.Close()

	var constraints []Constraint
	for result.Rows.Next() {
		vals, err := result.Rows.Values()
		if err != nil {
			return nil, fmt.Errorf("schema: constraints row: %w", err)
		}
		colStr := asString(vals[3])
		var cols []string
		if colStr != "" {
			cols = strings.Split(colStr, ",")
		}
		constraints = append(constraints, Constraint{
			Name:       asString(vals[0]),
			Type:       asString(vals[1]),
			Definition: asString(vals[2]),
			Columns:    cols,
		})
	}
	if err := result.Rows.Err(); err != nil {
		return nil, fmt.Errorf("schema: constraints iter: %w", err)
	}
	return constraints, nil
}

func (p *pgInspector) ForeignKeys(ctx context.Context, schemaName, table string) ([]ForeignKey, error) {
	schema := p.resolveSchema(schemaName)
	const q = `
SELECT
	con.conname,
	array_to_string(ARRAY(
		SELECT a.attname
		FROM unnest(con.conkey) WITH ORDINALITY AS k(attnum, ord)
		JOIN pg_attribute a ON a.attrelid = con.conrelid AND a.attnum = k.attnum
		ORDER BY k.ord
	), ',') AS columns,
	ref_cls.relname AS ref_table,
	ref_ns.nspname AS ref_schema,
	array_to_string(ARRAY(
		SELECT a.attname
		FROM unnest(con.confkey) WITH ORDINALITY AS k(attnum, ord)
		JOIN pg_attribute a ON a.attrelid = con.confrelid AND a.attnum = k.attnum
		ORDER BY k.ord
	), ',') AS ref_columns,
	CASE con.confdeltype
		WHEN 'a' THEN 'NO ACTION'
		WHEN 'r' THEN 'RESTRICT'
		WHEN 'c' THEN 'CASCADE'
		WHEN 'n' THEN 'SET NULL'
		WHEN 'd' THEN 'SET DEFAULT'
		ELSE ''
	END AS on_delete,
	CASE con.confupdtype
		WHEN 'a' THEN 'NO ACTION'
		WHEN 'r' THEN 'RESTRICT'
		WHEN 'c' THEN 'CASCADE'
		WHEN 'n' THEN 'SET NULL'
		WHEN 'd' THEN 'SET DEFAULT'
		ELSE ''
	END AS on_update
FROM pg_constraint con
JOIN pg_class c ON c.oid = con.conrelid
JOIN pg_namespace n ON n.oid = c.relnamespace
JOIN pg_class ref_cls ON ref_cls.oid = con.confrelid
JOIN pg_namespace ref_ns ON ref_ns.oid = ref_cls.relnamespace
WHERE con.contype = 'f' AND n.nspname = $1 AND c.relname = $2
ORDER BY con.conname`

	result, err := p.conn.Query(ctx, q, schema, table)
	if err != nil {
		return nil, fmt.Errorf("schema: foreign keys: %w", err)
	}
	defer result.Rows.Close()

	var fks []ForeignKey
	for result.Rows.Next() {
		vals, err := result.Rows.Values()
		if err != nil {
			return nil, fmt.Errorf("schema: foreign keys row: %w", err)
		}
		fks = append(fks, ForeignKey{
			Name:              asString(vals[0]),
			Columns:           splitCSV(asString(vals[1])),
			ReferencedTable:   asString(vals[2]),
			ReferencedSchema:  asString(vals[3]),
			ReferencedColumns: splitCSV(asString(vals[4])),
			OnDelete:          asString(vals[5]),
			OnUpdate:          asString(vals[6]),
		})
	}
	if err := result.Rows.Err(); err != nil {
		return nil, fmt.Errorf("schema: foreign keys iter: %w", err)
	}
	return fks, nil
}

// parseIndexColumns extracts column names from a CREATE INDEX definition.
func parseIndexColumns(def string) []string {
	start := strings.Index(def, "(")
	if start == -1 {
		return nil
	}
	// Find the matching closing paren
	end := strings.LastIndex(def, ")")
	// If there's a WHERE clause, find the paren before it
	if whereIdx := strings.Index(def, " WHERE "); whereIdx != -1 {
		end = strings.LastIndex(def[:whereIdx], ")")
	}
	if end == -1 || end <= start {
		return nil
	}
	inner := def[start+1 : end]
	parts := strings.Split(inner, ",")
	cols := make([]string, 0, len(parts))
	for _, p := range parts {
		cols = append(cols, strings.TrimSpace(p))
	}
	return cols
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

func asString(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func asInt64(v any) int64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int64:
		return val
	case int32:
		return int64(val)
	case int:
		return int64(val)
	case float64:
		return int64(val)
	case float32:
		return int64(val)
	default:
		return 0
	}
}

func asBool(v any) bool {
	if v == nil {
		return false
	}
	b, ok := v.(bool)
	if ok {
		return b
	}
	return false
}
