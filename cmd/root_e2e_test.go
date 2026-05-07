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
	videoPath := filepath.Join(tmp, "video.bin")
	targetPath := filepath.Join(tmp, "target.png")
	replacementPath := filepath.Join(tmp, "replacement.png")

	targetFrame := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x66, 0x72, 0x61, 0x6D, 0x65, 0x31}
	replacementFrame := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x66, 0x72, 0x61, 0x6D, 0x65, 0x58}
	videoData := append(append([]byte{}, targetFrame...), []byte{0x00, 0x00, 0x00}...)
	videoData = append(videoData, targetFrame...)

	if err := os.WriteFile(videoPath, videoData, 0o644); err != nil {
		t.Fatalf("write video: %v", err)
	}
	if err := os.WriteFile(targetPath, targetFrame, 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	if err := os.WriteFile(replacementPath, replacementFrame, 0o644); err != nil {
		t.Fatalf("write replacement: %v", err)
	}

	var buf bytes.Buffer
	cmd := NewRootCmd(video.NewEditor(video.OSFileSystem{}, frame.StringReplacer{}), video.NewExtractor(video.OSRunner{}))
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"replace", videoPath, targetPath, replacementPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	got, err := os.ReadFile(videoPath)
	if err != nil {
		t.Fatalf("read result: %v", err)
	}

	expected := bytes.ReplaceAll(videoData, targetFrame, replacementFrame)
	if !bytes.Equal(got, expected) {
		t.Fatalf("replacement mismatch")
	}

	if !strings.Contains(buf.String(), "video updated") {
		t.Fatalf("missing success message")
	}
}

func TestRootCmdReplaceFrameMissingVideoE2E(t *testing.T) {
	var buf bytes.Buffer
	cmd := NewRootCmd(video.NewEditor(video.OSFileSystem{}, frame.StringReplacer{}), video.NewExtractor(video.OSRunner{}))
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"replace", "missing.mp4", "/target.txt", "/replace.txt"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error")
	} else if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRootCmdReplaceFrameMissingTargetE2E(t *testing.T) {
	tmp := t.TempDir()
	videoPath := filepath.Join(tmp, "video.bin")
	replacementPath := filepath.Join(tmp, "replacement.png")

	videoData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x66, 0x72, 0x61, 0x6D, 0x65, 0x31}
	replacementFrame := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x66, 0x72, 0x61, 0x6D, 0x65, 0x58}

	if err := os.WriteFile(videoPath, videoData, 0o644); err != nil {
		t.Fatalf("write video: %v", err)
	}
	if err := os.WriteFile(replacementPath, replacementFrame, 0o644); err != nil {
		t.Fatalf("write replacement: %v", err)
	}

	var buf bytes.Buffer
	cmd := NewRootCmd(video.NewEditor(video.OSFileSystem{}, frame.StringReplacer{}), video.NewExtractor(video.OSRunner{}))
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"replace", videoPath, "/missing.png", replacementPath})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error")
	}
}
