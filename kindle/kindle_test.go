package kindle

import (
	"testing"

	"github.com/quelcom/homeink/config"
)

type testCases struct {
	input FBInk
	want  string
}

func TestFBInkString(t *testing.T) {
	config.EmbedTestConfig(`PathToFBInk = "/mnt/us/developer/fbink"`)

	var cases = []testCases{
		{
			input: FBInk{},
			want:  "/mnt/us/developer/fbink -q",
		},
		{
			input: FBInk{Verbose: true},
			want:  "/mnt/us/developer/fbink",
		},
		{
			input: FBInk{Text: "Terve", Size: 16, CenterX: true, CenterY: true},
			want:  "/mnt/us/developer/fbink -q -S 16 -m -M Terve",
		},
	}

	for _, c := range cases {
		got := c.input.String()
		if got != c.want {
			t.Errorf("got %q, want %q", got, c.want)
		}
	}

}
