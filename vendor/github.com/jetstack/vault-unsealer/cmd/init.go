package cmd

import (
	"github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/jetstack/vault-unsealer/pkg/vault"
)

const cfgInitRootToken = "init-root-token"
const cfgStoreRootToken = "store-root-token"
const cfgOverwriteExisting = "overwrite-existing"

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialise the target Vault instance",
	Long: `This command will verify the Cloud KMS service is accessible, then
run "vault init" against the target Vault instance, before encrypting and
storing the keys in the Cloud KMS keyring.

It will not unseal the Vault instance after initialising.`,
	Run: func(cmd *cobra.Command, args []string) {
		appConfig.BindPFlag(cfgInitRootToken, cmd.PersistentFlags().Lookup(cfgInitRootToken))
		appConfig.BindPFlag(cfgStoreRootToken, cmd.PersistentFlags().Lookup(cfgStoreRootToken))
		appConfig.BindPFlag(cfgOverwriteExisting, cmd.PersistentFlags().Lookup(cfgOverwriteExisting))

		store, err := kvStoreForConfig(appConfig)

		if err != nil {
			logrus.Fatalf("error creating kv store: %s", err.Error())
		}

		cl, err := api.NewClient(nil)

		if err != nil {
			logrus.Fatalf("error connecting to vault: %s", err.Error())
		}

		vaultConfig, err := vaultConfigForConfig(appConfig)

		if err != nil {
			logrus.Fatalf("error building vault config: %s", err.Error())
		}

		v, err := vault.New(store, cl, vaultConfig)

		if err != nil {
			logrus.Fatalf("error creating vault helper: %s", err.Error())
		}

		if err = v.Init(); err != nil {
			logrus.Fatalf("error initialising vault: %s", err.Error())
		}
	},
}

func init() {
	initCmd.PersistentFlags().String(cfgInitRootToken, "", "root token for the new vault cluster")
	initCmd.PersistentFlags().Bool(cfgStoreRootToken, true, "should the root token be stored in the key store")
	initCmd.PersistentFlags().Bool(cfgOverwriteExisting, false, "overwrite existing unseal keys and root tokens, possibly dangerous!")

	RootCmd.AddCommand(initCmd)
}
