package postgres

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

func oidToTypeName(oid uint32) string {
	if name, ok := oidTypeNames[oid]; ok {
		return name
	}
	return "unknown"
}
