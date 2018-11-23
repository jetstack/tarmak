// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var tunnelCmd = &cobra.Command{
	Use: "tunnel [destination] [destination port] [local port]",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return fmt.Errorf(
				"expecting only a destination, destination and local port argument, got=%s", args)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		t := tarmak.New(globalFlags)
		tunnel := t.SSH().Tunnel(args[0], args[1], args[2], false)

		retries := 5
		for {
			err := tunnel.Start()
			if err == nil {
				t.Log().Infof("tunnel started: %s", args)
				break
			}

			t.Log().Errorf("failed to start tunnel: %s", err)
			retries--
			if retries == 0 {
				t.Log().Error("failed to start tunnel after 5 attempts")
				t.Cleanup()
				os.Exit(1)
			}

			time.Sleep(time.Second * 2)
		}

		time.Sleep(time.Minute * 10)
		t.Cleanup()
		os.Exit(0)
	},
	Hidden:             true,
	DisableFlagParsing: true,
}

func init() {
	RootCmd.AddCommand(tunnelCmd)
}
