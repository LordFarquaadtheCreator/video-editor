package cmd

import (
	"bytes"
	"errors"
	"testing"

	"video-editor/pkg/video"
)

type fakeCmdRunner struct {
	name string
	args []string
	err  error
}

func (f *fakeCmdRunner) Run(name string, args ...string) error {
	f.name = name
	f.args = args
	return f.err
}

func TestExtractCmdSuccess(t *testing.T) {
	runner := &fakeCmdRunner{}
	extractor := video.NewExtractor(runner)
	cmd := NewExtractCmd(extractor)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"/video.mp4", "50", "/output"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExtractCmdInvalidSeconds(t *testing.T) {
	runner := &fakeCmdRunner{}
	extractor := video.NewExtractor(runner)
	cmd := NewExtractCmd(extractor)

	cmd.SetArgs([]string{"/video.mp4", "invalid", "/output"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error")
	}
}

func TestExtractCmdNegativeSeconds(t *testing.T) {
	runner := &fakeCmdRunner{}
	extractor := video.NewExtractor(runner)
	cmd := NewExtractCmd(extractor)

	cmd.SetArgs([]string{"/video.mp4", "-5", "/output"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error")
	}
}

func TestExtractCmdFFmpegError(t *testing.T) {
	runner := &fakeCmdRunner{err: errors.New("ffmpeg failed")}
	extractor := video.NewExtractor(runner)
	cmd := NewExtractCmd(extractor)

	cmd.SetArgs([]string{"/video.mp4", "50", "/output"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error")
	}
}
