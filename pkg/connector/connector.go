package connector

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Hello() {
	fmt.Println("Hello from the other side!")
}

func NewCommandStartConnector() *cobra.Command {
	cmd := &cobra.Command{
		Short: "Launch tarmak connector",
		Long:  "Launch tarmak connector",
		RunE: func(c *cobra.Command, args []string) error {
			Hello()

			return nil
		},
	}

	return cmd
}
