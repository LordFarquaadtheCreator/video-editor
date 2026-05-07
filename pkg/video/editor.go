package video

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"video-editor/pkg/frame"
)

// FileSystem abstracts disk interactions for easier testing.
type FileSystem interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm fs.FileMode) error
}

// CmdRunner abstracts command execution for testability.
type CmdRunner interface {
	Run(name string, args ...string) error
}

// OSRunner wraps os/exec.Command for real execution.
type OSRunner struct{}

func (OSRunner) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

// Editor encapsulates frame replacement within videos.
type Editor interface {
	ReplaceFrame(path, target, replacement string) ([]byte, error)
}

// Extractor extracts frames from video files.
type Extractor interface {
	ExtractFrame(videoPath, outputDir string, seconds float64) error
}

// OSFileSystem wraps os-level helpers.
type OSFileSystem struct{}

func (OSFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (OSFileSystem) WriteFile(path string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(path, data, perm)
}

// NewEditor returns a ready-to-use video editor.
func NewEditor(fs FileSystem, replacer frame.Replacer) Editor {
	return &videoEditor{
		FS:       fs,
		Replacer: replacer,
		// 0o644 means owner can read/write, group and others can only read.
		Perm: 0o644,
	}
}

type videoEditor struct {
	FS       FileSystem
	Replacer frame.Replacer
	Perm     fs.FileMode
}

func (v *videoEditor) ReplaceFrame(path, target, replacement string) ([]byte, error) {
	data, err := v.FS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read video: %w", err)
	}

	replaced, err := v.Replacer.Replace(data, target, replacement, v.FS)
	if err != nil {
		return nil, fmt.Errorf("replace frames: %w", err)
	}

	if err := v.FS.WriteFile(path, replaced, v.Perm); err != nil {
		return nil, fmt.Errorf("write video: %w", err)
	}

	return replaced, nil
}

// NewExtractor returns a ready-to-use frame extractor.
func NewExtractor(runner CmdRunner) Extractor {
	return &ffmpegExtractor{
		Runner: runner,
	}
}

type ffmpegExtractor struct {
	Runner CmdRunner
}

func (f *ffmpegExtractor) ExtractFrame(videoPath, outputDir string, seconds float64) error {
	outputPath := filepath.Join(outputDir, "target.png")
	args := []string{
		"-ss", fmt.Sprintf("%.2f", seconds),
		"-i", videoPath,
		"-vframes", "1",
		"-q:v", "2",
		outputPath,
	}
	if err := f.Runner.Run("ffmpeg", args...); err != nil {
		return fmt.Errorf("ffmpeg extract frame: %w", err)
	}
	return nil
}
