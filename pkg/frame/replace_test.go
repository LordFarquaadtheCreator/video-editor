package frame

import "testing"

func TestReplaceFrames(t *testing.T) {
	cases := []struct {
		name        string
		input       string
		target      string
		replacement string
		want        string
	}{
		{name: "simple", input: "frame1 frame2 frame1", target: "frame1", replacement: "frameX", want: "frameX frame2 frameX"},
		{name: "missing", input: "frameA", target: "notfound", replacement: "x", want: "frameA"},
		{name: "empty target", input: "frame", target: "", replacement: "x", want: "frame"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ReplaceFrames(tc.input, tc.target, tc.replacement)
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
