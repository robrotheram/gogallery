package datastore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckEXTSupportsCommonImageExtensions(t *testing.T) {
	for _, name := range []string{"photo.jpg", "photo.JPEG", "photo.png", "photo.gif", "photo.webp"} {
		if !CheckEXT(name) {
			t.Errorf("expected %q to be supported", name)
		}
	}
	if CheckEXT("notes.txt") {
		t.Fatal("non-image extension was accepted")
	}
}

func TestPathWithinRootRejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.jpg")
	if err := os.WriteFile(outside, []byte("image"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "linked.jpg")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if pathWithinRoot(link, root) {
		t.Fatal("symlink escaping the gallery root was accepted")
	}
	if !pathWithinRoot(filepath.Join(root, "new.jpg"), root) {
		t.Fatal("new file under the gallery root was rejected")
	}
}

func TestSafeRemovalDirectoryRejectsSensitiveRoots(t *testing.T) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"", workingDirectory, string(os.PathSeparator)} {
		if _, err := safeRemovalDirectory(path); err == nil {
			t.Fatalf("unsafe directory %q was accepted", path)
		}
	}
}
