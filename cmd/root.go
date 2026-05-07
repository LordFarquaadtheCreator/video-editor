package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"video-editor/pkg/logger"
	"video-editor/pkg/video"
)

func NewRootCmd(editor video.Editor, extractor video.Extractor) *cobra.Command {
	root := &cobra.Command{
		Use:   "video-editor",
		Short: "Video editing CLI tool.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			debug, _ := cmd.Flags().GetBool("debug")
			logger.SetDebug(debug)
		},
	}
	root.PersistentFlags().Bool("debug", false, "enable debug output")

	replaceCmd := &cobra.Command{
		Use:   "replace <video-path> <timestamp> <replacement-frame> [output]",
		Short: "Replace specific frames inside a video file.",
		Args:  cobra.RangeArgs(3, 4),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			timestampStr := args[1]
			replacement := args[2]
			output := "."
			if len(args) == 4 {
				output = args[3]
			}

			timestamp, err := strconv.ParseFloat(timestampStr, 64)
			if err != nil {
				return fmt.Errorf("parse timestamp: %w", err)
			}

			debug, _ := cmd.Flags().GetBool("debug")

			logger.Info("starting frame replacement for %s at timestamp %.2fs", path, timestamp)

			if _, err := editor.ReplaceFrame(path, timestamp, replacement, output, debug); err != nil {
				logger.Error("failed to replace frames: %v", err)
				return fmt.Errorf("replacing frames in %s: %w", path, err)
			}

			logger.Success("video updated")
			return nil
		},
	}

	extractCmd := NewExtractCmd(extractor)

	root.AddCommand(replaceCmd)
	root.AddCommand(extractCmd)
	return root
}

var rootCmd = NewRootCmd(
	video.NewEditor(
		video.OSFileSystem{},
		video.NewImageProcessor(video.OSRunner{}),
		video.NewFrameExtractor(video.OSRunner{}),
		video.NewFrameComparator(),
		video.NewVideoEncoder(video.OSRunner{}),
	),
	video.NewExtractor(video.OSRunner{}),
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Error("command failed: %v", err)
		os.Exit(1)
	}
}
