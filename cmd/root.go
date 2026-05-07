package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"video-editor/pkg/frame"
	"video-editor/pkg/video"
)

func NewRootCmd(editor video.Editor) *cobra.Command {
	return &cobra.Command{
		Use:   "video-editor <video-path> <target-frame> <replacement-frame>",
		Short: "Replace specific frames inside a video file.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			target := args[1]
			replacement := args[2]

			if _, err := editor.ReplaceFrame(path, target, replacement); err != nil {
				return fmt.Errorf("replacing frames in %s: %w", path, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "video updated: %s\n", path)
			return nil
		},
	}
}

var rootCmd = NewRootCmd(video.NewEditor(video.OSFileSystem{}, frame.StringReplacer{}))

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
