package conn

import (
	"testing"

	"github.com/zalando/go-keyring"
)

func TestMemoryKeyring(t *testing.T) {
	kr := NewMemoryKeyring()

	if err := kr.Set("svc", "user", "pass"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, err := kr.Get("svc", "user")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != "pass" {
		t.Errorf("Get = %q, want %q", got, "pass")
	}

	if err := kr.Delete("svc", "user"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = kr.Get("svc", "user")
	if err != keyring.ErrNotFound {
		t.Errorf("Get after delete: err = %v, want ErrNotFound", err)
	}
}

func TestMemoryKeyringNotFound(t *testing.T) {
	kr := NewMemoryKeyring()

	_, err := kr.Get("svc", "missing")
	if err != keyring.ErrNotFound {
		t.Errorf("Get missing: err = %v, want ErrNotFound", err)
	}

	err = kr.Delete("svc", "missing")
	if err != keyring.ErrNotFound {
		t.Errorf("Delete missing: err = %v, want ErrNotFound", err)
	}
}

func TestCredentialStore(t *testing.T) {
	cs := NewCredentialStore(NewMemoryKeyring())

	if err := cs.SetPassword("myconn", "s3cret"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}

	got, err := cs.GetPassword("myconn")
	if err != nil {
		t.Fatalf("GetPassword: %v", err)
	}
	if got != "s3cret" {
		t.Errorf("GetPassword = %q, want %q", got, "s3cret")
	}

	if err := cs.DeletePassword("myconn"); err != nil {
		t.Fatalf("DeletePassword: %v", err)
	}

	_, err = cs.GetPassword("myconn")
	if err == nil {
		t.Fatal("expected error after delete")
	}
}
