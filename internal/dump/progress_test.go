package dump

import (
	"strings"
	"testing"
)

func TestParseProgressBasic(t *testing.T) {
	input := `pg_dump: last built-in OID is 16383
pg_dump: reading extensions
pg_dump: identifying extension members
pg_dump: reading schemas
pg_dump: reading user-defined tables
pg_dump: dumping contents of table "public.users"
pg_dump: dumping contents of table "public.posts"
pg_dump: dumping contents of table "public.comments"
`
	ch := ParseProgress(strings.NewReader(input), 3)
	var events []ProgressEvent
	for ev := range ch {
		events = append(events, ev)
	}

	// 3 table events + 1 done
	if len(events) != 4 {
		t.Fatalf("got %d events, want 4", len(events))
	}

	// Check table events.
	wantTables := []string{"public.users", "public.posts", "public.comments"}
	for i, want := range wantTables {
		ev := events[i]
		if ev.Object != want {
			t.Errorf("event[%d].Object = %q, want %q", i, ev.Object, want)
		}
		if ev.Index != i+1 {
			t.Errorf("event[%d].Index = %d, want %d", i, ev.Index, i+1)
		}
		if ev.Total != 3 {
			t.Errorf("event[%d].Total = %d, want 3", i, ev.Total)
		}
		if ev.Done {
			t.Errorf("event[%d] should not be Done", i)
		}
	}

	// Check done event.
	done := events[3]
	if !done.Done {
		t.Error("last event should be Done")
	}
	if done.Index != 3 {
		t.Errorf("done.Index = %d, want 3", done.Index)
	}
}

func TestParseProgressNoTables(t *testing.T) {
	input := `pg_dump: last built-in OID is 16383
pg_dump: reading extensions
pg_dump: reading schemas
`
	ch := ParseProgress(strings.NewReader(input), 0)
	var events []ProgressEvent
	for ev := range ch {
		events = append(events, ev)
	}

	if len(events) != 1 {
		t.Fatalf("got %d events, want 1 (done only)", len(events))
	}
	if !events[0].Done {
		t.Error("expected Done event")
	}
	if events[0].Index != 0 {
		t.Errorf("done.Index = %d, want 0", events[0].Index)
	}
}

func TestParseProgressEmptyInput(t *testing.T) {
	ch := ParseProgress(strings.NewReader(""), 5)
	var events []ProgressEvent
	for ev := range ch {
		events = append(events, ev)
	}

	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	if !events[0].Done {
		t.Error("expected Done event on empty input")
	}
}

func TestParseProgressUnknownTotal(t *testing.T) {
	input := `pg_dump: dumping contents of table "t1"
pg_dump: dumping contents of table "t2"
`
	ch := ParseProgress(strings.NewReader(input), 0)
	var events []ProgressEvent
	for ev := range ch {
		events = append(events, ev)
	}

	if len(events) != 3 {
		t.Fatalf("got %d events, want 3", len(events))
	}
	for _, ev := range events[:2] {
		if ev.Total != 0 {
			t.Errorf("Total = %d, want 0 for unknown", ev.Total)
		}
	}
}

func TestParseProgressSchemaQualified(t *testing.T) {
	input := `pg_dump: dumping contents of table "myschema.mytable"
`
	ch := ParseProgress(strings.NewReader(input), 1)
	var events []ProgressEvent
	for ev := range ch {
		events = append(events, ev)
	}

	if len(events) < 1 {
		t.Fatal("no events")
	}
	if events[0].Object != "myschema.mytable" {
		t.Errorf("Object = %q, want %q", events[0].Object, "myschema.mytable")
	}
}

func TestParseProgressErrorEvent(t *testing.T) {
	r := &errReader{err: strings.NewReader("partial"), failAfter: 0}
	ch := ParseProgress(r, 1)
	var events []ProgressEvent
	for ev := range ch {
		events = append(events, ev)
	}

	// Should get an Err event from the scanner error.
	found := false
	for _, ev := range events {
		if ev.Err != nil {
			found = true
		}
	}
	if !found {
		t.Error("expected at least one error event")
	}
}

// errReader returns an error on Read after initial data.
type errReader struct {
	err       *strings.Reader
	failAfter int
	calls     int
}

func (r *errReader) Read(p []byte) (int, error) {
	r.calls++
	if r.calls > 1 {
		return 0, &testError{}
	}
	return r.err.Read(p)
}

type testError struct{}

func (e *testError) Error() string { return "test read error" }
