package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"video-editor/pkg/video"
)

type mockFrameExtractor struct{}

func (m *mockFrameExtractor) ExtractAllFrames(videoPath, outputDir string) (frameCount int, fps float64, err error) {
	if _, err := os.Stat(videoPath); err != nil {
		return 0, 0, err
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return 0, 0, err
	}
	frame1 := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x66, 0x72, 0x61, 0x6D, 0x65, 0x31}
	frame2 := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x66, 0x72, 0x61, 0x6D, 0x65, 0x32}
	if err := os.WriteFile(filepath.Join(outputDir, "frame_1.png"), frame1, 0o644); err != nil {
		return 0, 0, err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "frame_2.png"), frame2, 0o644); err != nil {
		return 0, 0, err
	}
	return 2, 30.0, nil
}

func (m *mockFrameExtractor) ExtractFrame(videoPath, outputDir string, seconds float64) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}
	frame := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x66, 0x72, 0x61, 0x6D, 0x65, 0x31}
	return os.WriteFile(filepath.Join(outputDir, "target_frame.png"), frame, 0o644)
}

type mockFrameComparator struct{}

func (m *mockFrameComparator) Compare(img1Path, img2Path string, threshold float64) (match bool, err error) {
	return true, nil
}

type mockVideoEncoder struct{}

func (m *mockVideoEncoder) EncodeFromFrames(framesDir, outputPath string, fps float64, originalVideoPath string) error {
	return os.WriteFile(outputPath, []byte("dummy video data"), 0o644)
}

type mockImageProcessor struct{}

func (m *mockImageProcessor) GetDimensions(imagePath string) (width, height int, err error) {
	return 1920, 1080, nil
}

func (m *mockImageProcessor) ConvertToPNG(inputPath, outputPath string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, data, 0o644)
}

func (m *mockImageProcessor) Resize(inputPath, outputPath string, width, height int) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, data, 0o644)
}

func TestRootCmdReplaceFrameE2E(t *testing.T) {
	tmp := t.TempDir()
	videoPath := filepath.Join(tmp, "video.bin")
	replacementPath := filepath.Join(tmp, "replacement.png")

	replacementFrame := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x66, 0x72, 0x61, 0x6D, 0x65, 0x58}
	videoData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x66, 0x72, 0x61, 0x6D, 0x65, 0x31, 0x00, 0x00, 0x00}

	if err := os.WriteFile(videoPath, videoData, 0o644); err != nil {
		t.Fatalf("write video: %v", err)
	}
	if err := os.WriteFile(replacementPath, replacementFrame, 0o644); err != nil {
		t.Fatalf("write replacement: %v", err)
	}

	var buf bytes.Buffer
	cmd := NewRootCmd(video.NewEditor(video.OSFileSystem{}, &mockImageProcessor{}, &mockFrameExtractor{}, &mockFrameComparator{}, &mockVideoEncoder{}), video.NewExtractor(video.OSRunner{}))
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"replace", videoPath, "0", replacementPath, videoPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
}

func TestRootCmdReplaceFrameMissingVideoE2E(t *testing.T) {
	var buf bytes.Buffer
	cmd := NewRootCmd(video.NewEditor(video.OSFileSystem{}, &mockImageProcessor{}, &mockFrameExtractor{}, &mockFrameComparator{}, &mockVideoEncoder{}), video.NewExtractor(video.OSRunner{}))
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"replace", "missing.mp4", "1.5", "/replace.txt", "missing.mp4"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error")
	} else if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRootCmdReplaceFrameInvalidTimestampE2E(t *testing.T) {
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
	cmd := NewRootCmd(video.NewEditor(video.OSFileSystem{}, &mockImageProcessor{}, &mockFrameExtractor{}, &mockFrameComparator{}, &mockVideoEncoder{}), video.NewExtractor(video.OSRunner{}))
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"replace", videoPath, "invalid", replacementPath, videoPath})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for invalid timestamp")
	}
}
