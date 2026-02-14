package fileutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpsertMarkerBlock_NewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testfile")

	content := "export FOO=bar\nexport BAZ=qux\n"
	err := UpsertMarkerBlock(path, content, "#")
	if err != nil {
		t.Fatalf("UpsertMarkerBlock failed: %v", err)
	}

	data, _ := os.ReadFile(path)
	got := string(data)
	if !strings.Contains(got, "# >>> ezproxy >>>") {
		t.Error("missing start marker")
	}
	if !strings.Contains(got, "# <<< ezproxy <<<") {
		t.Error("missing end marker")
	}
	if !strings.Contains(got, "export FOO=bar") {
		t.Error("missing content")
	}
}

func TestUpsertMarkerBlock_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testfile")

	existing := "# existing stuff\nPATH=/usr/bin\n"
	os.WriteFile(path, []byte(existing), 0644)

	content := "export PROXY=http://proxy:8080\n"
	err := UpsertMarkerBlock(path, content, "#")
	if err != nil {
		t.Fatalf("UpsertMarkerBlock failed: %v", err)
	}

	data, _ := os.ReadFile(path)
	got := string(data)
	if !strings.HasPrefix(got, "# existing stuff\n") {
		t.Error("existing content should be preserved at start")
	}
	if !strings.Contains(got, "export PROXY=http://proxy:8080") {
		t.Error("new content missing")
	}
}

func TestUpsertMarkerBlock_Replace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testfile")

	UpsertMarkerBlock(path, "OLD CONTENT\n", "#")
	UpsertMarkerBlock(path, "NEW CONTENT\n", "#")

	data, _ := os.ReadFile(path)
	got := string(data)
	if strings.Contains(got, "OLD CONTENT") {
		t.Error("old content should be replaced")
	}
	if !strings.Contains(got, "NEW CONTENT") {
		t.Error("new content missing")
	}
	if strings.Count(got, ">>> ezproxy >>>") != 1 {
		t.Error("should have exactly one start marker")
	}
}

func TestRemoveMarkerBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testfile")

	before := "line1\n"
	after := "line2\n"
	os.WriteFile(path, []byte(before), 0644)

	UpsertMarkerBlock(path, "PROXY STUFF\n", "#")

	f, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	f.WriteString(after)
	f.Close()

	err := RemoveMarkerBlock(path, "#")
	if err != nil {
		t.Fatalf("RemoveMarkerBlock failed: %v", err)
	}

	data, _ := os.ReadFile(path)
	got := string(data)
	if strings.Contains(got, "ezproxy") {
		t.Error("marker block should be removed")
	}
	if !strings.Contains(got, "line1") {
		t.Error("content before block should remain")
	}
	if !strings.Contains(got, "line2") {
		t.Error("content after block should remain")
	}
}

func TestHasMarkerBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testfile")

	os.WriteFile(path, []byte("nothing here\n"), 0644)
	if HasMarkerBlock(path, "#") {
		t.Error("should not have marker block")
	}

	UpsertMarkerBlock(path, "stuff\n", "#")
	if !HasMarkerBlock(path, "#") {
		t.Error("should have marker block")
	}
}

func TestGetMarkerBlockContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testfile")

	UpsertMarkerBlock(path, "FOO=bar\nBAZ=qux\n", "#")
	content, err := GetMarkerBlockContent(path, "#")
	if err != nil {
		t.Fatalf("GetMarkerBlockContent failed: %v", err)
	}
	if !strings.Contains(content, "FOO=bar") {
		t.Error("missing content")
	}
}
