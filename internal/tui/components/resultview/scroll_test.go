package resultview

import "testing"

func TestScrollState_SetRows(t *testing.T) {
	s := NewScrollState()
	rows := [][]string{{"a"}, {"b"}, {"c"}}
	s.SetRows(rows)

	if s.LoadedRows != 3 {
		t.Errorf("LoadedRows = %d, want 3", s.LoadedRows)
	}
	if len(s.Rows()) != 3 {
		t.Errorf("len(Rows) = %d, want 3", len(s.Rows()))
	}
}

func TestScrollState_AppendRows(t *testing.T) {
	s := NewScrollState()
	s.SetRows([][]string{{"a"}, {"b"}})
	s.AppendRows([][]string{{"c"}, {"d"}})

	if s.LoadedRows != 4 {
		t.Errorf("LoadedRows = %d, want 4", s.LoadedRows)
	}
}

func TestScrollState_Reset(t *testing.T) {
	s := NewScrollState()
	s.SetRows([][]string{{"a"}})
	s.AllLoaded = true
	s.TotalEstimate = 100
	s.Reset()

	if s.LoadedRows != 0 {
		t.Errorf("LoadedRows = %d, want 0", s.LoadedRows)
	}
	if s.TotalEstimate != -1 {
		t.Errorf("TotalEstimate = %d, want -1", s.TotalEstimate)
	}
	if s.AllLoaded {
		t.Error("AllLoaded should be false after reset")
	}
}

func TestScrollState_Eviction(t *testing.T) {
	s := NewScrollState()
	s.CacheLimit = 3

	rows := make([][]string, 5)
	for i := range rows {
		rows[i] = []string{string(rune('a' + i))}
	}
	s.SetRows(rows)

	if s.LoadedRows != 3 {
		t.Errorf("LoadedRows = %d, want 3 (capped)", s.LoadedRows)
	}
	if s.Rows()[0][0] != "c" {
		t.Errorf("first row = %q, want 'c' (oldest evicted)", s.Rows()[0][0])
	}
}

func TestScrollState_NeedsFetch(t *testing.T) {
	s := NewScrollState()
	s.PageSize = 100
	s.SetRows(make([][]string, 100))

	if !s.NeedsFetch(90) {
		t.Error("should need fetch when viewport near edge")
	}
	if s.NeedsFetch(50) {
		t.Error("should not need fetch when viewport far from edge")
	}

	s.AllLoaded = true
	if s.NeedsFetch(99) {
		t.Error("should not need fetch when all loaded")
	}
}

func TestScrollState_Defaults(t *testing.T) {
	s := NewScrollState()
	if s.PageSize != DefaultPageSize {
		t.Errorf("PageSize = %d, want %d", s.PageSize, DefaultPageSize)
	}
	if s.CacheLimit != DefaultCacheLimit {
		t.Errorf("CacheLimit = %d, want %d", s.CacheLimit, DefaultCacheLimit)
	}
	if s.TotalEstimate != -1 {
		t.Errorf("TotalEstimate = %d, want -1", s.TotalEstimate)
	}
}
