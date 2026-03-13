package editor

import "testing"

func TestGenerateUpdate(t *testing.T) {
	pk := PKValue{Columns: []string{"id"}, Values: []any{42}}
	r := GenerateUpdate("public", "users", pk, "name", "Bob")

	want := `UPDATE "public"."users" SET "name" = $1 WHERE "id" = $2`
	if r.SQL != want {
		t.Errorf("SQL:\n got  %s\n want %s", r.SQL, want)
	}
	if len(r.Args) != 2 || r.Args[0] != "Bob" || r.Args[1] != 42 {
		t.Errorf("args: got %v, want [Bob 42]", r.Args)
	}
}

func TestGenerateUpdate_NullPK(t *testing.T) {
	pk := PKValue{Columns: []string{"id", "seq"}, Values: []any{1, nil}}
	r := GenerateUpdate("public", "t", pk, "val", "x")

	want := `UPDATE "public"."t" SET "val" = $1 WHERE "id" = $2 AND "seq" IS NULL`
	if r.SQL != want {
		t.Errorf("SQL:\n got  %s\n want %s", r.SQL, want)
	}
	if len(r.Args) != 2 || r.Args[0] != "x" || r.Args[1] != 1 {
		t.Errorf("args: got %v, want [x 1]", r.Args)
	}
}

func TestGenerateUpdate_NullNewValue(t *testing.T) {
	pk := PKValue{Columns: []string{"id"}, Values: []any{1}}
	r := GenerateUpdate("public", "t", pk, "name", nil)

	want := `UPDATE "public"."t" SET "name" = $1 WHERE "id" = $2`
	if r.SQL != want {
		t.Errorf("SQL:\n got  %s\n want %s", r.SQL, want)
	}
	if len(r.Args) != 2 {
		t.Errorf("args: got %v", r.Args)
	}
	if r.Args[0] != nil {
		t.Errorf("expected nil arg for NULL value, got %v", r.Args[0])
	}
}

func TestGenerateInsert(t *testing.T) {
	vals := map[string]any{"name": "Alice", "age": 30}
	r := GenerateInsert("public", "users", vals)

	// map iteration order is non-deterministic, so check structure
	if len(r.Args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(r.Args))
	}
	if r.SQL == "" {
		t.Fatal("expected non-empty SQL")
	}
}

func TestGenerateInsert_Empty(t *testing.T) {
	r := GenerateInsert("public", "t", nil)
	want := `INSERT INTO "public"."t" DEFAULT VALUES`
	if r.SQL != want {
		t.Errorf("SQL:\n got  %s\n want %s", r.SQL, want)
	}
	if len(r.Args) != 0 {
		t.Errorf("expected 0 args, got %d", len(r.Args))
	}
}

func TestGenerateInsert_NullValue(t *testing.T) {
	vals := map[string]any{"name": nil}
	r := GenerateInsert("public", "t", vals)

	want := `INSERT INTO "public"."t" ("name") VALUES (NULL)`
	if r.SQL != want {
		t.Errorf("SQL:\n got  %s\n want %s", r.SQL, want)
	}
	if len(r.Args) != 0 {
		t.Errorf("expected 0 args for NULL insert, got %d", len(r.Args))
	}
}

func TestGenerateDelete(t *testing.T) {
	pk := PKValue{Columns: []string{"id"}, Values: []any{42}}
	r := GenerateDelete("public", "users", pk)

	want := `DELETE FROM "public"."users" WHERE "id" = $1`
	if r.SQL != want {
		t.Errorf("SQL:\n got  %s\n want %s", r.SQL, want)
	}
	if len(r.Args) != 1 || r.Args[0] != 42 {
		t.Errorf("args: got %v, want [42]", r.Args)
	}
}

func TestGenerateDelete_CompositePK(t *testing.T) {
	pk := PKValue{Columns: []string{"a", "b"}, Values: []any{1, 2}}
	r := GenerateDelete("public", "t", pk)

	want := `DELETE FROM "public"."t" WHERE "a" = $1 AND "b" = $2`
	if r.SQL != want {
		t.Errorf("SQL:\n got  %s\n want %s", r.SQL, want)
	}
	if len(r.Args) != 2 {
		t.Errorf("args: got %v, want [1 2]", r.Args)
	}
}

func TestQuoteIdent(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"users", `"users"`},
		{`my"table`, `"my""table"`},
	}
	for _, tt := range tests {
		got := quoteIdent(tt.in)
		if got != tt.want {
			t.Errorf("quoteIdent(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
