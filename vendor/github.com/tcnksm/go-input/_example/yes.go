package main

import (
	"fmt"
	"log"
	"os"

	"github.com/tcnksm/go-input"
)

func main() {
	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}

	query := "Do you love golang [Y/n]"
	name, err := ui.Ask(query, &input.Options{
		Required: true,
		// Validate input
		ValidateFunc: func(s string) error {
			if s != "Y" && s != "n" {
				return fmt.Errorf("input must be Y or n")
			}

			return nil
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Answer is %s\n", name)
}
