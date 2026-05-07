package cmd

import (
	"testing"

	"video-editor/pkg/frame"
	"video-editor/pkg/video"
)

func TestNewRootCmdHasSubcommands(t *testing.T) {
	editor := video.NewEditor(video.OSFileSystem{}, frame.StringReplacer{})
	extractor := video.NewExtractor(video.OSRunner{})
	rootCmd := NewRootCmd(editor, extractor)

	if len(rootCmd.Commands()) != 2 {
		t.Fatalf("expected 2 subcommands, got %d", len(rootCmd.Commands()))
	}

	commands := make(map[string]bool)
	for _, cmd := range rootCmd.Commands() {
		commands[cmd.Name()] = true
	}

	if !commands["replace"] {
		t.Fatal("missing replace command")
	}

	if !commands["extract"] {
		t.Fatal("missing extract command")
	}
}
