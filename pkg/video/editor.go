package video

import (
	"fmt"
	"io/fs"
	"os"

	"video-editor/pkg/frame"
)

// FileSystem abstracts disk interactions for easier testing.
type FileSystem interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm fs.FileMode) error
}

// Editor encapsulates frame replacement within videos.
type Editor interface {
	ReplaceFrame(path, target, replacement string) ([]byte, error)
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
		Perm:     0o644,
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

	replaced := v.Replacer.Replace(data, target, replacement)

	if err := v.FS.WriteFile(path, replaced, v.Perm); err != nil {
		return nil, fmt.Errorf("write video: %w", err)
	}

	return replaced, nil
}
