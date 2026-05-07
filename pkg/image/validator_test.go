package image

import "testing"

func TestValidateImage(t *testing.T) {
	cases := []struct {
		name    string
		data    []byte
		want    string
		wantErr bool
	}{
		{name: "PNG", data: []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00}, want: "PNG", wantErr: false},
		{name: "JPEG", data: []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00}, want: "JPEG", wantErr: false},
		{name: "GIF87a", data: []byte("GIF87a\x00\x00"), want: "GIF", wantErr: false},
		{name: "GIF89a", data: []byte("GIF89a\x00\x00"), want: "GIF", wantErr: false},
		{name: "WEBP", data: []byte("RIFF\x00\x00\x00\x00WEBP"), want: "WEBP", wantErr: false},
		{name: "BMP", data: []byte("BM\x00\x00"), want: "BMP", wantErr: false},
		{name: "text", data: []byte("not an image"), want: "", wantErr: true},
		{name: "empty", data: []byte{}, want: "", wantErr: true},
		{name: "short", data: []byte{0x89}, want: "", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ValidateImage(tc.data)
			if tc.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
