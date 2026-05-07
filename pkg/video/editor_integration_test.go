package video

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func skipIfNoFFmpeg(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found")
	}
}

// makeSolidColorPNG writes a solid-color 16x16 PNG to path.
func makeSolidColorPNG(t *testing.T, path string, c color.RGBA) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.Set(x, y, c)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create png: %v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
}

// makeGradientPNG writes a horizontal gradient from c1 (left) to c2 (right).
func makeGradientPNG(t *testing.T, path string, c1, c2 color.RGBA) {
	t.Helper()
	const size = 256
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			frac := float64(x) / float64(size-1)
			r := uint8(float64(c1.R)*(1-frac) + float64(c2.R)*frac)
			g := uint8(float64(c1.G)*(1-frac) + float64(c2.G)*frac)
			b := uint8(float64(c1.B)*(1-frac) + float64(c2.B)*frac)
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create png: %v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
}

// makeCheckerboardPNG writes an 8x8 checkerboard of two colors to path.
func makeCheckerboardPNG(t *testing.T, path string, c1, c2 color.RGBA) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			if (x/8+y/8)%2 == 0 {
				img.Set(x, y, c1)
			} else {
				img.Set(x, y, c2)
			}
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create png: %v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
}
func makeVideoFromPNG(t *testing.T, imgPath, videoPath string) {
	t.Helper()
	cmd := exec.Command("ffmpeg", "-y",
		"-loop", "1",
		"-i", imgPath,
		"-t", "1",
		"-r", "5",
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		videoPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("make video: %v\n%s", err, out)
	}
}

// extractFrameAtTime pulls a single frame from videoPath at the given second.
func extractFrameAtTime(t *testing.T, videoPath, outPath string, sec float64) {
	t.Helper()
	cmd := exec.Command("ffmpeg", "-y",
		"-ss", "0",
		"-i", videoPath,
		"-vframes", "1",
		outPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("extract frame: %v\n%s", err, out)
	}
}

type brightnessResult struct{ leftBright, rightBright bool }

// brightnessAt samples left and right edge pixels of the PNG and reports which sides are bright.
func brightnessAt(t *testing.T, path string) brightnessResult {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open png: %v", err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		t.Fatalf("decode png: %v", err)
	}
	bounds := img.Bounds()
	lR, lG, lB, _ := img.At(bounds.Min.X, bounds.Min.Y).RGBA()
	rR, rG, rB, _ := img.At(bounds.Max.X-1, bounds.Min.Y).RGBA()
	lLuma := (lR + lG + lB) / 3
	rLuma := (rR + rG + rB) / 3
	// threshold: >50% of max (65535) is "bright"
	return brightnessResult{leftBright: lLuma > 32767, rightBright: rLuma > 32767}
}

// dominantColor returns the color of the top-left pixel of the PNG at path.
func dominantColor(t *testing.T, path string) color.RGBA {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open png: %v", err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		t.Fatalf("decode png: %v", err)
	}
	r, g, b, a := img.At(0, 0).RGBA()
	return color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}

// TestIntegration_AllFramesReplaced creates a 1s video of left-bright gradient,
// replaces all frames with a right-bright gradient, and verifies output is right-bright.
func TestIntegration_AllFramesReplaced(t *testing.T) {
	skipIfNoFFmpeg(t)
	dir := t.TempDir()

	// left-bright: white on left, black on right
	// right-bright: black on left, white on right
	leftBrightPNG := filepath.Join(dir, "leftbright.png")
	rightBrightPNG := filepath.Join(dir, "rightbright.png")
	videoPath := filepath.Join(dir, "input.mp4")
	outputPath := filepath.Join(dir, "output.mp4")

	makeGradientPNG(t, leftBrightPNG, color.RGBA{255, 255, 255, 255}, color.RGBA{0, 0, 0, 255})
	makeGradientPNG(t, rightBrightPNG, color.RGBA{0, 0, 0, 255}, color.RGBA{255, 255, 255, 255})
	makeVideoFromPNG(t, leftBrightPNG, videoPath)

	editor := NewEditor(
		OSFileSystem{},
		NewImageProcessor(OSRunner{}),
		NewFrameExtractor(OSRunner{}),
		NewFrameComparator(),
		NewVideoEncoder(OSRunner{}),
	)

	// All frames are left-bright; timestamp 0.1s targets one; pHash matches all
	_, err := editor.ReplaceFrame(videoPath, 0.1, rightBrightPNG, outputPath, false)
	if err != nil {
		t.Fatalf("ReplaceFrame: %v", err)
	}

	// First frame of output should be right-bright: left pixel dark, right pixel bright
	outFrame := filepath.Join(dir, "out_frame.png")
	extractFrameAtTime(t, outputPath, outFrame, 0)

	got := brightnessAt(t, outFrame)
	// right-bright: pixel at x=0 dark, pixel at x=width-1 bright
	if got.leftBright {
		t.Errorf("expected right-bright frame but left side is still bright")
	}
}

// TestIntegration_OnlyMatchingFramesReplaced creates a 2s video:
// first second left-bright gradient, second right-bright gradient.
// Targets left-bright frames and replaces with a solid white image.
// Verifies early frames are white and late frames remain right-bright.
func TestIntegration_OnlyMatchingFramesReplaced(t *testing.T) {
	skipIfNoFFmpeg(t)
	dir := t.TempDir()

	leftBrightPNG := filepath.Join(dir, "leftbright.png")
	rightBrightPNG := filepath.Join(dir, "rightbright.png")
	whitePNG := filepath.Join(dir, "white.png")
	videoPath := filepath.Join(dir, "input.mp4")
	outputPath := filepath.Join(dir, "output.mp4")

	makeGradientPNG(t, leftBrightPNG, color.RGBA{255, 255, 255, 255}, color.RGBA{0, 0, 0, 255})
	makeGradientPNG(t, rightBrightPNG, color.RGBA{0, 0, 0, 255}, color.RGBA{255, 255, 255, 255})
	makeSolidColorPNG(t, whitePNG, color.RGBA{255, 255, 255, 255})

	// Build 2-second video: 5 frames left-bright, 5 frames right-bright
	// Generate PNG sequence directly to avoid concat issues
	framesDir := filepath.Join(dir, "input_frames")
	os.MkdirAll(framesDir, 0o755)
	for i := 1; i <= 5; i++ {
		src := leftBrightPNG
		dst := filepath.Join(framesDir, fmt.Sprintf("frame_%d.png", i))
		data, _ := os.ReadFile(src)
		os.WriteFile(dst, data, 0o644)
	}
	for i := 6; i <= 10; i++ {
		src := rightBrightPNG
		dst := filepath.Join(framesDir, fmt.Sprintf("frame_%d.png", i))
		data, _ := os.ReadFile(src)
		os.WriteFile(dst, data, 0o644)
	}

	cmd := exec.Command("ffmpeg", "-y",
		"-framerate", "5",
		"-i", filepath.Join(framesDir, "frame_%d.png"),
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		videoPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("encode video: %v\n%s", err, out)
	}

	editor := NewEditor(
		OSFileSystem{},
		NewImageProcessor(OSRunner{}),
		NewFrameExtractor(OSRunner{}),
		NewFrameComparator(),
		NewVideoEncoder(OSRunner{}),
	)

	// Target left-bright at 0.1s — pHash matches left-bright only, not right-bright
	_, err := editor.ReplaceFrame(videoPath, 0.1, whitePNG, outputPath, false)
	if err != nil {
		t.Fatalf("ReplaceFrame: %v", err)
	}

	// Frame at 0.1s (was left-bright) should now be white: both sides bright
	earlyFrame := filepath.Join(dir, "early_frame.png")
	cmd = exec.Command("ffmpeg", "-y",
		"-ss", "0.1",
		"-i", outputPath,
		"-vframes", "1",
		earlyFrame,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("extract early frame: %v\n%s", err, out)
	}
	earlyB := brightnessAt(t, earlyFrame)
	// white replacement: both left and right should be bright
	if !earlyB.leftBright || !earlyB.rightBright {
		t.Errorf("early frame should be all-white, got leftBright=%v rightBright=%v", earlyB.leftBright, earlyB.rightBright)
	}

	// Frame at 1.5s (was right-bright) should be unchanged: left dark, right bright
	lateFrame := filepath.Join(dir, "late_frame.png")
	cmd = exec.Command("ffmpeg", "-y",
		"-ss", "1.5",
		"-i", outputPath,
		"-vframes", "1",
		lateFrame,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("extract late frame: %v\n%s", err, out)
	}
	lateB := brightnessAt(t, lateFrame)
	// right-bright: left pixel dark, right pixel bright
	if lateB.leftBright {
		t.Errorf("late frame should be right-bright (left dark), got leftBright=%v", lateB.leftBright)
	}
	if !lateB.rightBright {
		t.Errorf("late frame should be right-bright (right bright), got rightBright=%v", lateB.rightBright)
	}
}
