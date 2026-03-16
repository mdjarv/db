package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Common PostgreSQL OIDs mapped to human-readable type names.
// See: https://github.com/postgres/postgres/blob/master/src/include/catalog/pg_type.dat
var oidTypeNames = map[uint32]string{
	16:   "bool",
	17:   "bytea",
	18:   "char",
	19:   "name",
	20:   "int8",
	21:   "int2",
	23:   "int4",
	24:   "regproc",
	25:   "text",
	26:   "oid",
	114:  "json",
	142:  "xml",
	700:  "float4",
	701:  "float8",
	790:  "money",
	1042: "bpchar",
	1043: "varchar",
	1082: "date",
	1083: "time",
	1114: "timestamp",
	1184: "timestamptz",
	1186: "interval",
	1266: "timetz",
	1560: "bit",
	1562: "varbit",
	1700: "numeric",
	2950: "uuid",
	3802: "jsonb",
	3904: "int4range",
	3906: "numrange",
	3908: "tsrange",
	3910: "tstzrange",
	3912: "daterange",
	3926: "int8range",
}

// CompositeField describes a single field within a composite type.
type CompositeField struct {
	Name     string
	TypeName string
	TypeOID  uint32
}

// TypeMap resolves OIDs to type names, combining hardcoded built-in types
// with dynamically loaded types from pg_type.
type TypeMap struct {
	dynamic         map[uint32]string
	enumValues      map[uint32][]string
	compositeFields map[uint32][]CompositeField
	arrayElemOID    map[uint32]uint32 // array OID → element OID
	Warnings        []string          // scan errors during loading
}

// NewTypeMap creates a TypeMap with only hardcoded entries (fallback).
func NewTypeMap() *TypeMap {
	return &TypeMap{}
}

// LoadTypeMap queries pg_type, pg_enum, and pg_attribute to build a
// complete type map. Falls back to hardcoded-only on error.
// All "char" catalog columns are cast to text to avoid pgx scan issues.
func LoadTypeMap(ctx context.Context, pool *pgxpool.Pool) *TypeMap {
	tm := &TypeMap{
		dynamic:         make(map[uint32]string),
		enumValues:      make(map[uint32][]string),
		compositeFields: make(map[uint32][]CompositeField),
		arrayElemOID:    make(map[uint32]uint32),
	}
	if err := tm.loadTypes(ctx, pool); err != nil {
		return NewTypeMap()
	}
	tm.loadEnums(ctx, pool)
	tm.loadComposites(ctx, pool)
	return tm
}

func (tm *TypeMap) loadTypes(ctx context.Context, pool *pgxpool.Pool) error {
	const q = `
SELECT t.oid,
       t.typname::text,
       t.typcategory::text,
       t.typelem,
       e.typname::text AS elem_name
FROM pg_type t
LEFT JOIN pg_type e ON e.oid = t.typelem AND t.typcategory = 'A'
WHERE t.typnamespace NOT IN (
    SELECT oid FROM pg_namespace WHERE nspname = 'pg_toast'
)`

	rows, err := pool.Query(ctx, q)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var oid, typelem uint32
		var typname, category string
		var elemName *string
		if err := rows.Scan(&oid, &typname, &category, &typelem, &elemName); err != nil {
			tm.Warnings = append(tm.Warnings, fmt.Sprintf("types: scan: %v", err))
			continue
		}
		if category == "A" && elemName != nil {
			tm.dynamic[oid] = *elemName + "[]"
			tm.arrayElemOID[oid] = typelem
		} else {
			tm.dynamic[oid] = typname
		}
	}
	return rows.Err()
}

func (tm *TypeMap) loadEnums(ctx context.Context, pool *pgxpool.Pool) {
	const q = `
SELECT e.enumtypid, e.enumlabel::text
FROM pg_enum e
ORDER BY e.enumtypid, e.enumsortorder`

	rows, err := pool.Query(ctx, q)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var oid uint32
		var label string
		if err := rows.Scan(&oid, &label); err != nil {
			tm.Warnings = append(tm.Warnings, fmt.Sprintf("enums: scan: %v", err))
			continue
		}
		tm.enumValues[oid] = append(tm.enumValues[oid], label)
	}
}

func (tm *TypeMap) loadComposites(ctx context.Context, pool *pgxpool.Pool) {
	// Key by type OID (ct.oid), not relation OID (a.attrelid).
	// pg_attribute.attrelid = pg_type.typrelid, not pg_type.oid.
	const q = `
SELECT ct.oid,
       a.attname::text,
       t.typname::text,
       a.atttypid
FROM pg_attribute a
JOIN pg_type ct ON ct.typrelid = a.attrelid
JOIN pg_type t ON t.oid = a.atttypid
WHERE ct.typcategory = 'C'
  AND a.attnum > 0
  AND NOT a.attisdropped
ORDER BY ct.oid, a.attnum`

	rows, err := pool.Query(ctx, q)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var typeOID, atttypid uint32
		var attname, typname string
		if err := rows.Scan(&typeOID, &attname, &typname, &atttypid); err != nil {
			tm.Warnings = append(tm.Warnings, fmt.Sprintf("composites: scan: %v", err))
			continue
		}
		fieldType := typname
		if name, ok := oidTypeNames[atttypid]; ok {
			fieldType = name
		} else if name, ok := tm.dynamic[atttypid]; ok {
			fieldType = name
		}
		tm.compositeFields[typeOID] = append(tm.compositeFields[typeOID], CompositeField{
			Name:     attname,
			TypeName: fieldType,
			TypeOID:  atttypid,
		})
	}
}

// Resolve returns the type name for a given OID.
func (tm *TypeMap) Resolve(oid uint32) string {
	if name, ok := oidTypeNames[oid]; ok {
		return name
	}
	if tm.dynamic != nil {
		if name, ok := tm.dynamic[oid]; ok {
			return name
		}
	}
	return "unknown"
}

// ElemOID returns the element type OID for an array type.
func (tm *TypeMap) ElemOID(oid uint32) (uint32, bool) {
	if tm.arrayElemOID == nil {
		return 0, false
	}
	elem, ok := tm.arrayElemOID[oid]
	return elem, ok
}

// CompositeFields returns the fields of a composite type, or nil.
func (tm *TypeMap) CompositeFields(oid uint32) []CompositeField {
	if tm.compositeFields == nil {
		return nil
	}
	fields, ok := tm.compositeFields[oid]
	if !ok {
		return nil
	}
	out := make([]CompositeField, len(fields))
	copy(out, fields)
	return out
}

// EnumValues returns the enum labels for a given OID, or nil if not an enum.
// For array types, returns the element type's enum values.
func (tm *TypeMap) EnumValues(oid uint32) []string {
	if tm.enumValues == nil {
		return nil
	}
	vals, ok := tm.enumValues[oid]
	if !ok {
		if elemOID, aok := tm.ElemOID(oid); aok {
			vals, ok = tm.enumValues[elemOID]
			if !ok {
				return nil
			}
		} else {
			return nil
		}
	}
	out := make([]string, len(vals))
	copy(out, vals)
	return out
}
