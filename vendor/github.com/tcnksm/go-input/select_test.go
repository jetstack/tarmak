package input

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"testing"
)

func TestSelect(t *testing.T) {
	cases := []struct {
		list      []string
		opts      *Options
		userInput io.Reader
		expect    string
	}{
		{
			list:      []string{"A", "B", "C"},
			opts:      &Options{},
			userInput: bytes.NewBufferString("1\n"),
			expect:    "A",
		},

		{
			list: []string{"A", "B", "C"},
			opts: &Options{
				Default: "A",
			},
			userInput: bytes.NewBufferString("\n"),
			expect:    "A",
		},

		{
			list: []string{"A", "B", "C"},
			opts: &Options{
				Default: "A",
			},
			userInput: bytes.NewBufferString("3\n"),
			expect:    "C",
		},

		// Loop
		{
			list: []string{"A", "B", "C"},
			opts: &Options{
				Loop: true,
			},
			userInput: bytes.NewBufferString("\n3\n"),
			expect:    "C",
		},

		// Loop
		{
			list: []string{"A", "B", "C"},
			opts: &Options{
				Loop: true,
			},
			userInput: bytes.NewBufferString("\n\n\n\n\n2\n"),
			expect:    "B",
		},

		// Loop
		{
			list: []string{"A", "B", "C"},
			opts: &Options{
				Loop: true,
			},
			userInput: bytes.NewBufferString("4\n3\n"),
			expect:    "C",
		},

		// Loop
		{
			list: []string{"A", "B", "C"},
			opts: &Options{
				Loop: true,
			},
			userInput: bytes.NewBufferString("A\n3\n"),
			expect:    "C",
		},
	}

	for i, c := range cases {
		ui := &UI{
			Writer: ioutil.Discard,
			Reader: c.userInput,
		}

		ans, err := ui.Select("", c.list, c.opts)
		if err != nil {
			t.Fatalf("#%d expect not to occurr error: %s", i, err)
		}

		if ans != c.expect {
			t.Fatalf("#%d expect %q to be eq %q", i, ans, c.expect)
		}
	}
}

func TestSelect_invalidDefault(t *testing.T) {
	ui := &UI{
		Writer: ioutil.Discard,
	}
	_, err := ui.Select("Which?", []string{"A", "B", "C"}, &Options{
		// "D" is not in select target list
		Default: "D",
	})

	if err == nil {
		t.Fatal("expect err to be occurr")
	}
}

func ExampleUI_Select() {
	ui := &UI{
		// In real world, Reader is os.Stdin and input comes
		// from user actual input.
		Reader: bytes.NewBufferString("3\n"),
		Writer: ioutil.Discard,
	}

	query := "Which language do you prefer to use?"
	lang, _ := ui.Select(query, []string{"go", "Go", "golang"}, &Options{
		Default: "Go",
	})

	fmt.Println(lang)
	// Output: golang
}
