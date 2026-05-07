package video

import (
	"errors"
	"io/fs"
	"testing"
)

func TestVideoEditorReplaceFrame(t *testing.T) {
	fs := &fakeFS{data: []byte("frame1 frame2")}
	replacer := &fakeReplacer{output: []byte("frameX frame2")}

	editor := NewEditor(fs, replacer)

	got, err := editor.ReplaceFrame("video.mp4", "frame1", "frameX")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(got) != string(replacer.output) {
		t.Fatalf("got %q, want %q", got, replacer.output)
	}

	if fs.wrotePath != "video.mp4" {
		t.Fatalf("wrote path %s", fs.wrotePath)
	}

	if string(fs.wroteData) != string(replacer.output) {
		t.Fatalf("written data mismatch")
	}

	if fs.wrotePerm != 0o644 {
		t.Fatalf("wrong perm: %o", fs.wrotePerm)
	}
}

func TestVideoEditorReplaceFrameReadError(t *testing.T) {
	fs := &fakeFS{readErr: errors.New("missing file")}
	replacer := &fakeReplacer{}
	editor := NewEditor(fs, replacer)

	if _, err := editor.ReplaceFrame("video.mp4", "frame1", "frameX"); err == nil {
		t.Fatal("expected error")
	}

	if fs.wrotePath != "" {
		t.Fatalf("unexpected write on read failure")
	}
}

func TestVideoEditorReplaceFrameWriteError(t *testing.T) {
	fs := &fakeFS{data: []byte("frame1 frame2"), writeErr: errors.New("perm denied")}
	replacer := &fakeReplacer{output: []byte("frameX frame2")}
	editor := NewEditor(fs, replacer)

	if _, err := editor.ReplaceFrame("video.mp4", "frame1", "frameX"); err == nil {
		t.Fatal("expected error")
	}

	if fs.wrotePath != "video.mp4" {
		t.Fatalf("did not attempt write")
	}
}

func TestVideoEditorReplaceFrameEmptyTarget(t *testing.T) {
	input := []byte("frame1 frame2")
	fs := &fakeFS{data: input}
	replacer := &fakeReplacer{output: input}
	editor := NewEditor(fs, replacer)

	got, err := editor.ReplaceFrame("video.mp4", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(got) != string(input) {
		t.Fatalf("got %q, want %q", got, input)
	}

	if string(replacer.input) != string(input) {
		t.Fatalf("replacer received incorrect data")
	}

	if fs.wrotePath == "" {
		t.Fatalf("expected write even when target empty")
	}
}

type fakeFS struct {
	data      []byte
	readErr   error
	writeErr  error
	wrotePath string
	wroteData []byte
	wrotePerm fs.FileMode
}

func (f *fakeFS) ReadFile(path string) ([]byte, error) {
	if f.readErr != nil {
		return nil, f.readErr
	}
	return f.data, nil
}

func (f *fakeFS) WriteFile(path string, data []byte, perm fs.FileMode) error {
	f.wrotePath = path
	f.wroteData = data
	f.wrotePerm = perm
	if f.writeErr != nil {
		return f.writeErr
	}
	return nil
}

type fakeReplacer struct {
	output []byte
	input  []byte
}

func (f *fakeReplacer) Replace(input []byte, _, _ string) []byte {
	f.input = append([]byte(nil), input...)
	return f.output
}
