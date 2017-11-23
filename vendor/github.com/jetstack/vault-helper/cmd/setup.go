package cmd

import (
	"fmt"
	"time"

	vault "github.com/hashicorp/vault/api"
	"github.com/spf13/cobra"

	"github.com/jetstack/vault-helper/pkg/kubernetes"
)

// initCmd represents the init command
var setupCmd = &cobra.Command{
	Use:   "setup [cluster ID]",
	Short: "Setup kubernetes on a running vault server.",
	Run: func(cmd *cobra.Command, args []string) {
		log := LogLevel(cmd)

		v, err := vault.NewClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		i, err := newInstanceToken(cmd)
		if err != nil {
			i.Log.Fatal(err)
		}

		if err := i.TokenRenewRun(); err != nil {
			i.Log.Fatal(err)
		}

		k := kubernetes.New(v, log)
		if err != nil {
			log.Fatal(err)
		}

		if len(args) > 0 {
			k.SetClusterID(args[0])
		} else {
			log.Fatal("no cluster id was given")
		}

		if err := setFlagsKubernetes(k, cmd); err != nil {
			log.Fatal(err)
		}

		if err := k.Ensure(); err != nil {
			log.Fatal(err)
		}

		for n, t := range k.InitTokens() {
			log.Infof(n + "-init_token := " + t)
		}

	},
}

func init() {
	setupCmd.PersistentFlags().Duration(kubernetes.FlagMaxValidityCA, time.Hour*24*365*20, "Maxium validity for CA certificates")
	setupCmd.PersistentFlags().Duration(kubernetes.FlagMaxValidityAdmin, time.Hour*24*365, "Maxium validity for admin certificates")
	setupCmd.PersistentFlags().Duration(kubernetes.FlagMaxValidityComponents, time.Hour*24*30, "Maxium validity for component certificates")

	setupCmd.PersistentFlags().String(kubernetes.FlagInitTokenEtcd, "", "Set init-token-etcd   (Default to new token)")
	setupCmd.PersistentFlags().String(kubernetes.FlagInitTokenWorker, "", "Set init-token-worker (Default to new token)")
	setupCmd.PersistentFlags().String(kubernetes.FlagInitTokenMaster, "", "Set init-token-master (Default to new token)")
	setupCmd.PersistentFlags().String(kubernetes.FlagInitTokenAll, "", "Set init-token-all    (Default to new token)")

	RootCmd.AddCommand(setupCmd)
}

func setFlagsKubernetes(k *kubernetes.Kubernetes, cmd *cobra.Command) error {
	if value, err := cmd.PersistentFlags().GetDuration(kubernetes.FlagMaxValidityComponents); err != nil {
		if err != nil {
			return fmt.Errorf("error parsing %s '%s': %s", kubernetes.FlagMaxValidityComponents, value, err)
		}
		k.MaxValidityComponents = value
	}

	if value, err := cmd.PersistentFlags().GetDuration(kubernetes.FlagMaxValidityAdmin); err != nil {
		if err != nil {
			return fmt.Errorf("error parsing %s '%s': %s", kubernetes.FlagMaxValidityAdmin, value, err)
		}
		k.MaxValidityAdmin = value
	}

	if value, err := cmd.PersistentFlags().GetDuration(kubernetes.FlagMaxValidityCA); err != nil {
		if err != nil {
			return fmt.Errorf("error parsing %s '%s': %s", kubernetes.FlagMaxValidityCA, value, err)
		}
		k.MaxValidityCA = value
	}

	// Init token flags
	value, err := cmd.PersistentFlags().GetString(kubernetes.FlagInitTokenEtcd)
	if err != nil {
		return fmt.Errorf("error parsing %s '%s': %s", kubernetes.FlagInitTokenEtcd, value, err)
	}
	k.FlagInitTokens.Etcd = value

	value, err = cmd.PersistentFlags().GetString(kubernetes.FlagInitTokenMaster)
	if err != nil {
		return fmt.Errorf("error parsing %s '%s': %s", kubernetes.FlagInitTokenMaster, value, err)
	}
	k.FlagInitTokens.Master = value

	value, err = cmd.PersistentFlags().GetString(kubernetes.FlagInitTokenWorker)
	if err != nil {
		return fmt.Errorf("error parsing %s '%s': %s", kubernetes.FlagInitTokenWorker, value, err)
	}
	k.FlagInitTokens.Worker = value

	value, err = cmd.PersistentFlags().GetString(kubernetes.FlagInitTokenAll)
	if err != nil {
		return fmt.Errorf("error parsing %s '%s': %s", kubernetes.FlagInitTokenAll, value, err)
	}
	k.FlagInitTokens.All = value

	return nil
}
