package input

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"testing"
)

func TestAsk(t *testing.T) {
	cases := []struct {
		opts      *Options
		userInput io.Reader
		expect    string
	}{
		{
			opts:      &Options{},
			userInput: bytes.NewBufferString("Taichi\n"),
			expect:    "Taichi",
		},

		{
			opts: &Options{
				Default: "Nakashima",
			},
			userInput: bytes.NewBufferString("\n"),
			expect:    "Nakashima",
		},

		// Loop & Required
		{
			opts: &Options{
				Required: true,
				Loop:     true,
			},
			userInput: bytes.NewBufferString("\nNakashima\n"),
			expect:    "Nakashima",
		},
	}

	for i, c := range cases {
		ui := &UI{
			Writer: ioutil.Discard,
			Reader: c.userInput,
		}

		ans, err := ui.Ask("", c.opts)
		if err != nil {
			t.Fatalf("#%d expect not to occurr error: %s", i, err)
		}

		if ans != c.expect {
			t.Fatalf("#%d expect %q to be eq %q", i, ans, c.expect)
		}
	}
}

func ExampleUI_Ask() {
	ui := &UI{
		// In real world, Reader is os.Stdin and input comes
		// from user actual input.
		Reader: bytes.NewBufferString("tcnksm"),
		Writer: ioutil.Discard,
	}

	query := "What is your name?"
	name, _ := ui.Ask(query, &Options{})

	fmt.Println(name)
	// Output: tcnksm
}
