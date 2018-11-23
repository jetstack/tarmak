// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak"
)

var tunnelCmd = &cobra.Command{
	Use: "tunnel [destination] [port]",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf(
				"expecting only a destination and port argument, got=%s", args)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Errorf("HERE")
		t := tarmak.New(globalFlags)
		tunnel := t.SSH().Tunnel(args[0], args[1], false)

		retries := 10
		for {
			if retries == 0 {
				t.Cleanup()
				os.Exit(1)
			}

			err := tunnel.Start()
			if err == nil {
				break
			}

			t.Log().Errorf("failed to start tunnel: %s", err)
			time.Sleep(time.Second * 2)
			retries--
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
