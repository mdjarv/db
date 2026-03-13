package conn

import (
	"path/filepath"
	"testing"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	return NewStore(filepath.Join(t.TempDir(), "connections.yaml"))
}

func TestStoreAddGet(t *testing.T) {
	s := testStore(t)

	cfg := ConnectionConfig{Name: "dev", Host: "localhost", Port: 5432, User: "app", DBName: "devdb"}
	if err := s.Add(cfg); err != nil {
		t.Fatalf("Add: %v", err)
	}

	got, err := s.Get("dev")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Host != "localhost" || got.User != "app" || got.DBName != "devdb" {
		t.Errorf("Get = %+v, want matching fields", got)
	}
}

func TestStoreAddRequiresName(t *testing.T) {
	s := testStore(t)
	err := s.Add(ConnectionConfig{Host: "localhost", Port: 5432})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestStoreList(t *testing.T) {
	s := testStore(t)

	for _, name := range []string{"a", "b", "c"} {
		if err := s.Add(ConnectionConfig{Name: name, Host: "h", Port: 5432}); err != nil {
			t.Fatal(err)
		}
	}

	list, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("List len = %d, want 3", len(list))
	}
}

func TestStoreRemove(t *testing.T) {
	s := testStore(t)
	cfg := ConnectionConfig{Name: "tmp", Host: "h", Port: 5432}
	if err := s.Add(cfg); err != nil {
		t.Fatal(err)
	}
	if err := s.Remove("tmp"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	_, err := s.Get("tmp")
	if err == nil {
		t.Fatal("expected error after remove")
	}
}

func TestStoreRemoveNotFound(t *testing.T) {
	s := testStore(t)
	if err := s.Remove("nonexistent"); err == nil {
		t.Fatal("expected error for nonexistent")
	}
}

func TestStoreDefault(t *testing.T) {
	s := testStore(t)

	cfg := ConnectionConfig{Name: "prod", Host: "db.prod", Port: 5432, User: "admin", DBName: "proddb"}
	if err := s.Add(cfg); err != nil {
		t.Fatal(err)
	}
	if err := s.SetDefault("prod"); err != nil {
		t.Fatalf("SetDefault: %v", err)
	}

	got, err := s.Default()
	if err != nil {
		t.Fatalf("Default: %v", err)
	}
	if got.Name != "prod" || got.Host != "db.prod" {
		t.Errorf("Default = %+v", got)
	}
}

func TestStoreDefaultUnset(t *testing.T) {
	s := testStore(t)
	_, err := s.Default()
	if err == nil {
		t.Fatal("expected error when no default set")
	}
}

func TestStoreSetDefaultNotFound(t *testing.T) {
	s := testStore(t)
	if err := s.SetDefault("missing"); err == nil {
		t.Fatal("expected error for nonexistent connection")
	}
}

func TestStoreRemoveClearsDefault(t *testing.T) {
	s := testStore(t)

	if err := s.Add(ConnectionConfig{Name: "x", Host: "h", Port: 5432}); err != nil {
		t.Fatal(err)
	}
	if err := s.SetDefault("x"); err != nil {
		t.Fatal(err)
	}
	if err := s.Remove("x"); err != nil {
		t.Fatal(err)
	}
	_, err := s.Default()
	if err == nil {
		t.Fatal("expected error after removing default connection")
	}
}

func TestStoreLoadEmpty(t *testing.T) {
	s := testStore(t)
	f, err := s.load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(f.Connections) != 0 {
		t.Errorf("expected empty connections, got %d", len(f.Connections))
	}
}
