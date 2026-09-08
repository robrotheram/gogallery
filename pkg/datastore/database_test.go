package datastore

import (
	"os"
	"path/filepath"
	"testing"

	"gogallery/pkg/monitor"
)

func TestOpenUsesRequestedDatabasePathAndPrivatePermissions(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "custom.db")
	store, err := Open(databasePath, monitor.NewCMDMonitor())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })

	info, err := os.Stat(databasePath)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("database permissions = %o, want 600", got)
	}
}

func TestOpenRejectsEmptyDatabasePath(t *testing.T) {
	if _, err := Open(" ", monitor.NewCMDMonitor()); err == nil {
		t.Fatal("Open() accepted an empty database path")
	}
}
