package cmd

import (
	"time"

	"github.com/Sirupsen/logrus"
	vault "github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"

	"github.com/jetstack-experimental/vault-helper/pkg/kubernetes"
)

// initCmd represents the init command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup kubernetes on a running vault server",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logrus.New()
		logger.Level = logrus.DebugLevel

		v, err := vault.NewClient(nil)
		if err != nil {
			logger.Fatal(err)
		}

		k := kubernetes.New(v)
		if err != nil {
			logger.Fatal(err)
		}

		if err := k.Run(cmd, args); err != nil {
			logger.Fatal(err)
		}

		for n, t := range k.InitTokens() {
			logrus.Infof(n + "-init_token := " + t)
		}

	},
}

func init() {
	setupCmd.PersistentFlags().Duration(kubernetes.FlagMaxValidityCA, time.Hour*24*365*20, "Maxium validity for CA certificates")
	setupCmd.PersistentFlags().Duration(kubernetes.FlagMaxValidityAdmin, time.Hour*24*365, "Maxium validity for admin certificates")
	setupCmd.PersistentFlags().Duration(kubernetes.FlagMaxValidityComponents, time.Hour*24*30, "Maxium validity for component certificates")
	RootCmd.AddCommand(setupCmd)
}
