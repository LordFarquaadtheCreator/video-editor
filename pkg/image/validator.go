package image

import (
	"bytes"
	"errors"
	"fmt"
)

var (
	ErrInvalidImage = errors.New("not a valid image format")
)

// ValidateImage checks if data is a valid image format using magic bytes.
func ValidateImage(data []byte) (string, error) {
	if len(data) < 2 {
		return "", ErrInvalidImage
	}

	// PNG: \x89PNG\r\n\x1a\n
	if bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
		return "PNG", nil
	}

	// JPG/JPEG: \xff\xd8\xff
	if bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}) {
		return "JPEG", nil
	}

	// GIF: GIF87a or GIF89a
	if bytes.HasPrefix(data, []byte("GIF87a")) || bytes.HasPrefix(data, []byte("GIF89a")) {
		return "GIF", nil
	}

	// WEBP: RIFF....WEBP
	if len(data) >= 12 && bytes.HasPrefix(data, []byte("RIFF")) && bytes.HasPrefix(data[8:], []byte("WEBP")) {
		return "WEBP", nil
	}

	// BMP: BM
	if bytes.HasPrefix(data, []byte("BM")) {
		return "BMP", nil
	}

	return "", ErrInvalidImage
}

// ValidateFile reads file and validates it's an image.
type FileReader interface {
	ReadFile(path string) ([]byte, error)
}

func ValidateFile(path string, fs FileReader) (string, error) {
	data, err := fs.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}

	format, err := ValidateImage(data)
	if err != nil {
		return "", fmt.Errorf("%s: %w", path, ErrInvalidImage)
	}

	return format, nil
}
