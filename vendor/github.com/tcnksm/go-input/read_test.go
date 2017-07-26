package input

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"
)

func TestRead(t *testing.T) {
	cases := []struct {
		opts      *readOptions
		userInput io.Reader
		expect    string
	}{
		{
			opts: &readOptions{
				mask:    false,
				maskVal: "",
			},
			userInput: bytes.NewBufferString("passw0rd"),
			expect:    "passw0rd",
		},

		{
			opts: &readOptions{
				mask:    false,
				maskVal: "",
			},
			userInput: bytes.NewBufferString("taichi nakashima"),
			expect:    "taichi nakashima",
		},

		// No good way to test masking...
	}

	for i, tc := range cases {
		ui := &UI{
			Writer: ioutil.Discard,
			Reader: tc.userInput,
		}

		out, err := ui.read(tc.opts)
		if err != nil {
			t.Fatalf("#%d expect not to be error: %s", i, err)
		}

		if out != tc.expect {
			t.Fatalf("#%d expect %q to be eq %q", i, out, tc.expect)
		}
	}
}
