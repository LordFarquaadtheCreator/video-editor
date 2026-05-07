package video

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/corona10/goimagehash"
)

func TestPHashDistance_RedVsBlue(t *testing.T) {
	skipIfNoFFmpeg(t)
	dir := t.TempDir()

	// Gradients with different luminance directions: bright-to-dark vs dark-to-bright
	leftBrightPNG := filepath.Join(dir, "leftbright.png")
	rightBrightPNG := filepath.Join(dir, "rightbright.png")
	makeGradientPNG(t, leftBrightPNG, color.RGBA{255, 255, 255, 255}, color.RGBA{0, 0, 0, 255})
	makeGradientPNG(t, rightBrightPNG, color.RGBA{0, 0, 0, 255}, color.RGBA{255, 255, 255, 255})

	pairs := []struct{ a, b, label string }{
		{leftBrightPNG, rightBrightPNG, "left-bright vs right-bright"},
		{leftBrightPNG, leftBrightPNG, "left-bright vs self"},
	}

	for _, hashFn := range []struct {
		name string
		fn   func(image.Image) (*goimagehash.ImageHash, error)
	}{
		{"PerceptionHash", goimagehash.PerceptionHash},
		{"AverageHash", goimagehash.AverageHash},
		{"DifferenceHash", goimagehash.DifferenceHash},
	} {
		for _, p := range pairs {
			f1, _ := os.Open(p.a)
			f2, _ := os.Open(p.b)
			img1, _, _ := image.Decode(f1)
			img2, _, _ := image.Decode(f2)
			f1.Close()
			f2.Close()
			h1, _ := hashFn.fn(img1)
			h2, _ := hashFn.fn(img2)
			d, _ := h1.Distance(h2)
			fmt.Printf("%s | %s distance=%d\n", hashFn.name, p.label, d)
		}
	}
}
