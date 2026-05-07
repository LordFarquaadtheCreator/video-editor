package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"video-editor/pkg/frame"
	"video-editor/pkg/video"
)

func TestRootCmdReplaceFrameE2E(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "video.txt")
	if err := os.WriteFile(path, []byte("frame1 frame2"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	var buf bytes.Buffer
	cmd := NewRootCmd(video.NewEditor(video.OSFileSystem{}, frame.StringReplacer{}))
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{path, "frame1", "frameX"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read result: %v", err)
	}

	if string(got) != "frameX frame2" {
		t.Fatalf("unexpected file: %s", got)
	}

	if !strings.Contains(buf.String(), "video updated") {
		t.Fatalf("missing success message")
	}
}

func TestRootCmdReplaceFrameMissingVideoE2E(t *testing.T) {
	var buf bytes.Buffer
	cmd := NewRootCmd(video.NewEditor(video.OSFileSystem{}, frame.StringReplacer{}))
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"missing.mp4", "frame", "X"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error")
	} else if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("unexpected error: %v", err)
	}
}
