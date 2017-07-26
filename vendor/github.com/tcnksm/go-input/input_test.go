package input

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestValidateFunc_implement(t *testing.T) {
	var _ ValidateFunc = defaultValidateFunc
}

func ExampleValidateFunc() {
	ui := &UI{
		// In real world, Reader is os.Stdin and input comes
		// from user actual input
		Reader: bytes.NewBufferString("Y\n"),
		Writer: ioutil.Discard,
	}

	query := "Do you love golang [Y/n]"
	ans, _ := ui.Ask(query, &Options{
		// Define validateFunc to validate user input is
		// 'Y' or 'n'. If not returns error.
		ValidateFunc: func(s string) error {
			if s != "Y" && s != "n" {
				return fmt.Errorf("input must be Y or n")
			}

			return nil
		},
	})

	fmt.Println(ans)
	// Output: Y
}
