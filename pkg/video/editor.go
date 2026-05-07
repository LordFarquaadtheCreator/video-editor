package video

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"video-editor/pkg/logger"

	"github.com/corona10/goimagehash"
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
	ReplaceFrame(path string, timestamp float64, replacement, output string, debug bool) ([]byte, error)
}

// Extractor extracts frames from video files.
type Extractor interface {
	ExtractFrame(videoPath, outputDir string, seconds float64) error
}

// ImageProcessor handles image normalization via FFmpeg.
type ImageProcessor interface {
	GetDimensions(imagePath string) (width, height int, err error)
	ConvertToPNG(inputPath, outputPath string) error
	Resize(inputPath, outputPath string, width, height int) error
}

// FrameExtractor extracts all frames from a video to a directory.
type FrameExtractor interface {
	ExtractAllFrames(videoPath, outputDir string) (frameCount int, fps float64, err error)
}

// FrameComparator compares two images for similarity.
type FrameComparator interface {
	Compare(img1Path, img2Path string, threshold float64) (match bool, err error)
}

// VideoEncoder encodes a video from a sequence of images.
type VideoEncoder interface {
	EncodeFromFrames(framesDir, outputPath string, fps float64, originalVideoPath string) error
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
func NewEditor(fs FileSystem, processor ImageProcessor, extractor FrameExtractor, comparator FrameComparator, encoder VideoEncoder) Editor {
	return &videoEditor{
		FS:         fs,
		Processor:  processor,
		Extractor:  extractor,
		Comparator: comparator,
		Encoder:    encoder,
		Perm:       0o644,
	}
}

type videoEditor struct {
	FS         FileSystem
	Processor  ImageProcessor
	Extractor  FrameExtractor
	Comparator FrameComparator
	Encoder    VideoEncoder
	Perm       fs.FileMode
}

func (v *videoEditor) ReplaceFrame(path string, timestamp float64, replacement, output string, debug bool) ([]byte, error) {
	if timestamp < 0 {
		return nil, fmt.Errorf("timestamp must be non-negative")
	}

	logger.Info("creating temporary directory")
	tmpDir, err := os.MkdirTemp("", "video-editor-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)
	framesDir := filepath.Join(tmpDir, "frames")

	logger.Info("extracting frames from video")
	frameCount, fps, err := v.Extractor.ExtractAllFrames(path, framesDir)
	if err != nil {
		return nil, fmt.Errorf("extract frames: %w", err)
	}
	logger.Info("extracted %d frames at %.2f fps", frameCount, fps)

	targetFrameNum := int(timestamp*fps) + 1
	targetFramePath := filepath.Join(framesDir, fmt.Sprintf("frame_%d.png", targetFrameNum))

	if _, err := os.Stat(targetFramePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("target frame at timestamp %.2fs (frame %d) does not exist", timestamp, targetFrameNum)
	}

	replacementBytes, err := v.FS.ReadFile(replacement)
	if err != nil {
		return nil, fmt.Errorf("read replacement: %w", err)
	}

	if len(replacementBytes) < 8 || !bytes.HasPrefix(replacementBytes, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
		return nil, fmt.Errorf("replacement must be a valid PNG image")
	}

	targetWidth, targetHeight, err := v.Processor.GetDimensions(targetFramePath)
	if err != nil {
		return nil, fmt.Errorf("get target dimensions: %w", err)
	}
	logger.Debug("target frame dimensions: %dx%d", targetWidth, targetHeight)

	logger.Info("processing replacement frame")
	replacementPNG := filepath.Join(tmpDir, "replacement_normalized.png")
	if err := v.Processor.ConvertToPNG(replacement, replacementPNG); err != nil {
		return nil, fmt.Errorf("convert replacement to PNG: %w", err)
	}
	defer os.Remove(replacementPNG)

	replacementResized := filepath.Join(tmpDir, "replacement_resized.png")
	if err := v.Processor.Resize(replacementPNG, replacementResized, targetWidth, targetHeight); err != nil {
		return nil, fmt.Errorf("resize replacement: %w", err)
	}
	defer os.Remove(replacementResized)

	replacedCount := 0
	logger.Info("comparing frames to find similar ones")
	entries, err := os.ReadDir(framesDir)
	if err != nil {
		return nil, fmt.Errorf("read frames dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "frame_") || !strings.HasSuffix(name, ".png") {
			continue
		}
		framePath := filepath.Join(framesDir, name)
		// Skip comparing target frame to itself to avoid self-replacement during comparison
		if framePath == targetFramePath {
			continue
		}
		match, err := v.Comparator.Compare(targetFramePath, framePath, 10.0)
		if err != nil {
			logger.Debug("compare failed for %s: %v", name, err)
			continue
		}
		if match {
			data, err := os.ReadFile(replacementResized)
			if err != nil {
				return nil, fmt.Errorf("read replacement frame: %w", err)
			}
			if err := os.WriteFile(framePath, data, 0o644); err != nil {
				return nil, fmt.Errorf("replace frame %s: %w", entry.Name(), err)
			}
			replacedCount++
		}
	}

	// Also replace the target frame itself
	data, err := os.ReadFile(replacementResized)
	if err != nil {
		return nil, fmt.Errorf("read replacement frame: %w", err)
	}
	if err := os.WriteFile(targetFramePath, data, 0o644); err != nil {
		return nil, fmt.Errorf("replace target frame: %w", err)
	}
	replacedCount++

	if debug {
		logger.Debug("replaced %d frames", replacedCount)
		logger.Debug("encoding from %s to %s at %.2f fps", framesDir, output, fps)
	}

	logger.Info("encoding video from frames")
	if err := v.Encoder.EncodeFromFrames(framesDir, output, fps, path); err != nil {
		return nil, fmt.Errorf("encode video: %w", err)
	}
	logger.Info("video encoding complete")

	result, err := v.FS.ReadFile(output)
	if err != nil {
		return nil, fmt.Errorf("read result: %w", err)
	}

	return result, nil
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

// NewImageProcessor returns a ready-to-use image processor.
func NewImageProcessor(runner CmdRunner) ImageProcessor {
	return &ffmpegImageProcessor{Runner: runner}
}

type ffmpegImageProcessor struct {
	Runner CmdRunner
}

func (f *ffmpegImageProcessor) GetDimensions(imagePath string) (width, height int, err error) {
	args := []string{
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=s=x:p=0",
		imagePath,
	}
	cmd := exec.Command("ffprobe", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, 0, fmt.Errorf("ffprobe get dimensions: %w", err)
	}

	var w, h int
	_, err = fmt.Sscanf(string(output), "%dx%d", &w, &h)
	if err != nil {
		return 0, 0, fmt.Errorf("parse dimensions: %w", err)
	}
	return w, h, nil
}

func (f *ffmpegImageProcessor) ConvertToPNG(inputPath, outputPath string) error {
	args := []string{
		"-i", inputPath,
		outputPath,
	}
	if err := f.Runner.Run("ffmpeg", args...); err != nil {
		return fmt.Errorf("ffmpeg convert to png: %w", err)
	}
	return nil
}

func (f *ffmpegImageProcessor) Resize(inputPath, outputPath string, width, height int) error {
	args := []string{
		"-i", inputPath,
		"-vf", fmt.Sprintf("scale=%d:%d", width, height),
		outputPath,
	}
	if err := f.Runner.Run("ffmpeg", args...); err != nil {
		return fmt.Errorf("ffmpeg resize: %w", err)
	}
	return nil
}

// NewFrameExtractor returns a ready-to-use frame extractor.
func NewFrameExtractor(runner CmdRunner) FrameExtractor {
	return &ffmpegFrameExtractor{Runner: runner}
}

type ffmpegFrameExtractor struct {
	Runner CmdRunner
}

func (f *ffmpegFrameExtractor) ExtractAllFrames(videoPath, outputDir string) (frameCount int, fps float64, err error) {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return 0, 0, fmt.Errorf("create output dir: %w", err)
	}

	args := []string{
		"-y",
		"-accurate_seek",
		"-i", videoPath,
		filepath.Join(outputDir, "frame_%d.png"),
	}
	if err := f.Runner.Run("ffmpeg", args...); err != nil {
		return 0, 0, fmt.Errorf("ffmpeg extract frames: %w", err)
	}

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return 0, 0, fmt.Errorf("read output dir: %w", err)
	}
	frameCount = len(entries)

	if frameCount == 0 {
		return 0, 0, fmt.Errorf("no frames extracted")
	}

	probeArgs := []string{
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=r_frame_rate",
		"-of", "csv=s=x:p=0",
		videoPath,
	}
	cmd := exec.Command("ffprobe", probeArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return frameCount, 30, nil
	}

	_, err = fmt.Sscanf(string(output), "%f", &fps)
	if err != nil {
		return frameCount, 30, nil
	}

	return frameCount, fps, nil
}

// NewFrameComparator returns a ready-to-use frame comparator.
func NewFrameComparator() FrameComparator {
	return &pHashFrameComparator{}
}

type pHashFrameComparator struct{}

func (p *pHashFrameComparator) Compare(img1Path, img2Path string, threshold float64) (match bool, err error) {
	img1, err := os.Open(img1Path)
	if err != nil {
		return false, fmt.Errorf("open image1: %w", err)
	}
	defer img1.Close()

	img2, err := os.Open(img2Path)
	if err != nil {
		return false, fmt.Errorf("open image2: %w", err)
	}
	defer img2.Close()

	decoded1, _, err := image.Decode(img1)
	if err != nil {
		return false, fmt.Errorf("decode image1: %w", err)
	}

	decoded2, _, err := image.Decode(img2)
	if err != nil {
		return false, fmt.Errorf("decode image2: %w", err)
	}

	hash1, err := goimagehash.PerceptionHash(decoded1)
	if err != nil {
		return false, fmt.Errorf("hash image1: %w", err)
	}

	hash2, err := goimagehash.PerceptionHash(decoded2)
	if err != nil {
		return false, fmt.Errorf("hash image2: %w", err)
	}

	distance, err := hash1.Distance(hash2)
	if err != nil {
		return false, fmt.Errorf("compare hashes: %w", err)
	}

	return distance <= int(threshold), nil
}

// NewVideoEncoder returns a ready-to-use video encoder.
func NewVideoEncoder(runner CmdRunner) VideoEncoder {
	return &ffmpegVideoEncoder{Runner: runner}
}

type ffmpegVideoEncoder struct {
	Runner CmdRunner
}

func (f *ffmpegVideoEncoder) EncodeFromFrames(framesDir, outputPath string, fps float64, originalVideoPath string) error {
	inputPattern := filepath.Join(framesDir, "frame_%d.png")
	args := []string{
		"-framerate", fmt.Sprintf("%.2f", fps),
		"-i", inputPattern,
		"-i", originalVideoPath,
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-map", "0:v:0",
		"-map", "1:a:0?",
		"-shortest",
		outputPath,
	}
	if err := f.Runner.Run("ffmpeg", args...); err != nil {
		return fmt.Errorf("ffmpeg encode video: %w", err)
	}
	return nil
}
