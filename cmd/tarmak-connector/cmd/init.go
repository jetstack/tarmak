// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const FlagLogLevel = "log-level"

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var RootCmd = &cobra.Command{
	Use:   "connector",
	Short: "tarmak connector to facilitate tarmak and terraform communications",
}

func init() {
	RootCmd.PersistentFlags().IntP(FlagLogLevel, "l", 1, "Set the log level of output. 0-Fatal 1-Info 2-Debug")

	startCmd := NewCommandStartConnector()
	startCmd.Use = "start"
	RootCmd.AddCommand(startCmd)
}

func LogLevel(cmd *cobra.Command) *logrus.Entry {
	logger := logrus.New()

	i, err := RootCmd.PersistentFlags().GetInt("log-level")
	if err != nil {
		logrus.Fatalf("failed to get log level of flag: %s", err)
	}
	if i < 0 || i > 2 {
		logrus.Fatalf("invalid valid log level")
	}
	switch i {
	case 0:
		logger.Level = logrus.FatalLevel
	case 1:
		logger.Level = logrus.InfoLevel
	case 2:
		logger.Level = logrus.DebugLevel
	}

	return logrus.NewEntry(logger)
}
