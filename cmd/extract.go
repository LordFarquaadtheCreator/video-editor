package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"video-editor/pkg/logger"
	"video-editor/pkg/video"
)

func NewExtractCmd(extractor video.Extractor) *cobra.Command {
	return &cobra.Command{
		Use:   "extract <video-path> <seconds> <output-dir>",
		Short: "Extract a frame from video at timestamp.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			videoPath := args[0]
			secondsStr := args[1]
			outputDir := args[2]

			seconds, err := strconv.ParseFloat(secondsStr, 64)
			if err != nil {
				return fmt.Errorf("invalid seconds value %q: %w", secondsStr, err)
			}

			if seconds < 0 {
				return fmt.Errorf("seconds must be non-negative, got %f", seconds)
			}

			logger.Info("extracting frame from %s at %.2fs", videoPath, seconds)

			if err := extractor.ExtractFrame(videoPath, outputDir, seconds); err != nil {
				logger.Error("failed to extract frame: %v", err)
				return fmt.Errorf("extracting frame from %s: %w", videoPath, err)
			}

			outputPath := outputDir + "/target.png"
			logger.Success("frame extracted: %s", outputPath)
			return nil
		},
	}
}
