package sqlstorage

import (
	"testing"
)

func TestNew(t *testing.T) {
	dsn := "postgres://user:pass@localhost:5432/db"
	s := New(dsn)
	if s == nil {
		t.Fatal("expected storage to be non-nil")
	}
	if s.dsn != dsn {
		t.Errorf("expected dsn %s, got %s", dsn, s.dsn)
	}
}
