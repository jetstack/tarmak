package main

import (
	"log"
	"os"

	"github.com/tcnksm/go-input"
)

func main() {
	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}

	query := "Which language do you prefer to use?"
	lang, err := ui.Select(query, []string{"go", "Go", "golang"}, &input.Options{
		Default: "Go",
		Loop:    true,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Answer is %s\n", lang)
}
