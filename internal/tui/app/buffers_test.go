package app

import (
	"strings"
	"testing"

	"github.com/mdjarv/db/internal/tui/core"
)

func TestBufferManager_InitialState(t *testing.T) {
	bm := NewBufferManager()
	if bm.Count() != 1 {
		t.Errorf("count = %d, want 1", bm.Count())
	}
	if bm.ActiveIndex() != 1 {
		t.Errorf("active = %d, want 1", bm.ActiveIndex())
	}
	if bm.Active().Query != "" {
		t.Error("initial buffer should be empty")
	}
}

func TestBufferManager_New(t *testing.T) {
	bm := NewBufferManager()
	bm.Active().Query = "SELECT 1"

	ok := bm.New()
	if !ok {
		t.Fatal("New() returned false")
	}
	if bm.Count() != 2 {
		t.Errorf("count = %d, want 2", bm.Count())
	}
	if bm.ActiveIndex() != 2 {
		t.Errorf("active = %d, want 2", bm.ActiveIndex())
	}
	if bm.Active().Query != "" {
		t.Error("new buffer should be empty")
	}
}

func TestBufferManager_NewMaxLimit(t *testing.T) {
	bm := NewBufferManager()
	bm.max = 3
	bm.New()
	bm.New()
	ok := bm.New()
	if ok {
		t.Error("New() should return false at max")
	}
	if bm.Count() != 3 {
		t.Errorf("count = %d, want 3", bm.Count())
	}
}

func TestBufferManager_Close(t *testing.T) {
	bm := NewBufferManager()
	bm.Active().Query = "first"
	bm.New()
	bm.Active().Query = "second"

	ok := bm.Close()
	if !ok {
		t.Fatal("Close() returned false")
	}
	if bm.Count() != 1 {
		t.Errorf("count = %d, want 1", bm.Count())
	}
	if bm.Active().Query != "first" {
		t.Errorf("query = %q, want %q", bm.Active().Query, "first")
	}
}

func TestBufferManager_CloseLastBuffer(t *testing.T) {
	bm := NewBufferManager()
	ok := bm.Close()
	if ok {
		t.Error("Close() should return false for last buffer")
	}
	if bm.Count() != 1 {
		t.Errorf("count = %d, want 1", bm.Count())
	}
}

func TestBufferManager_NextPrev(t *testing.T) {
	bm := NewBufferManager()
	bm.Active().Query = "one"
	bm.New()
	bm.Active().Query = "two"
	bm.New()
	bm.Active().Query = "three"

	// at buffer 3, next wraps to 1
	bm.Next()
	if bm.ActiveIndex() != 1 {
		t.Errorf("after Next from 3: active = %d, want 1", bm.ActiveIndex())
	}

	// at buffer 1, prev wraps to 3
	bm.Prev()
	if bm.ActiveIndex() != 3 {
		t.Errorf("after Prev from 1: active = %d, want 3", bm.ActiveIndex())
	}

	// prev to 2
	bm.Prev()
	if bm.ActiveIndex() != 2 {
		t.Errorf("after Prev from 3: active = %d, want 2", bm.ActiveIndex())
	}
}

func TestBufferManager_SwitchTo(t *testing.T) {
	bm := NewBufferManager()
	bm.Active().Query = "one"
	bm.New()
	bm.Active().Query = "two"
	bm.New()
	bm.Active().Query = "three"

	ok := bm.SwitchTo(1)
	if !ok || bm.ActiveIndex() != 1 {
		t.Errorf("SwitchTo(1): ok=%v, active=%d", ok, bm.ActiveIndex())
	}

	ok = bm.SwitchTo(0)
	if ok {
		t.Error("SwitchTo(0) should return false")
	}

	ok = bm.SwitchTo(4)
	if ok {
		t.Error("SwitchTo(4) should return false")
	}
}

func TestBufferManager_StatePreservation(t *testing.T) {
	bm := NewBufferManager()
	bm.Active().Query = "SELECT 1"
	bm.Active().Modified = true
	bm.Active().Columns = []core.ResultColumn{{Name: "id"}}
	bm.Active().Rows = [][]string{{"1"}}
	bm.Active().HasData = true
	bm.Active().CursorRow = 5
	bm.Active().ColOffset = 2

	bm.New()
	bm.Active().Query = "SELECT 2"

	// switch back
	bm.Prev()
	buf := bm.Active()
	if buf.Query != "SELECT 1" {
		t.Errorf("query = %q, want %q", buf.Query, "SELECT 1")
	}
	if !buf.Modified {
		t.Error("modified should be preserved")
	}
	if !buf.HasData {
		t.Error("hasData should be preserved")
	}
	if buf.CursorRow != 5 {
		t.Errorf("cursorRow = %d, want 5", buf.CursorRow)
	}
	if buf.ColOffset != 2 {
		t.Errorf("colOffset = %d, want 2", buf.ColOffset)
	}
}

func TestBufferManager_List(t *testing.T) {
	bm := NewBufferManager()
	bm.Active().Query = "SELECT 1"
	bm.Active().Modified = true
	bm.New()
	bm.Active().Query = "SELECT 2"

	list := bm.List()
	if !strings.Contains(list, "SELECT 1") {
		t.Error("list should contain first query")
	}
	if !strings.Contains(list, "SELECT 2") {
		t.Error("list should contain second query")
	}
	if !strings.Contains(list, "+") {
		t.Error("list should show modified indicator")
	}
	if !strings.Contains(list, "%") {
		t.Error("list should show active indicator")
	}
}

func TestBufferManager_CloseMiddle(t *testing.T) {
	bm := NewBufferManager()
	bm.Active().Query = "one"
	bm.New()
	bm.Active().Query = "two"
	bm.New()
	bm.Active().Query = "three"

	// switch to middle buffer and close it
	bm.SwitchTo(2)
	bm.Close()

	if bm.Count() != 2 {
		t.Errorf("count = %d, want 2", bm.Count())
	}
	// should now be on "three" (shifted to index 2)
	if bm.Active().Query != "three" {
		t.Errorf("query = %q, want %q", bm.Active().Query, "three")
	}
}

func TestBufferManager_CloseLastIndex(t *testing.T) {
	bm := NewBufferManager()
	bm.Active().Query = "one"
	bm.New()
	bm.Active().Query = "two"

	// active is buffer 2 (last), close it
	bm.Close()
	if bm.ActiveIndex() != 1 {
		t.Errorf("active = %d, want 1", bm.ActiveIndex())
	}
	if bm.Active().Query != "one" {
		t.Errorf("query = %q, want %q", bm.Active().Query, "one")
	}
}
