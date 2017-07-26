package kubernetes

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

const FlagMaxValidityAdmin = "max-validity-admin"
const FlagMaxValidityCA = "max-validity-ca"
const FlagMaxValidityComponents = "max-validity-components"

func (k *Kubernetes) Run(cmd *cobra.Command, args []string) error {

	if value, err := cmd.PersistentFlags().GetDuration(FlagMaxValidityComponents); err != nil {
		if err != nil {
			return fmt.Errorf("error parsing %s '%s': %s", FlagMaxValidityComponents, value, err)
		}
		k.MaxValidityComponents = value
	}

	if value, err := cmd.PersistentFlags().GetDuration(FlagMaxValidityAdmin); err != nil {
		if err != nil {
			return fmt.Errorf("error parsing %s '%s': %s", FlagMaxValidityAdmin, value, err)
		}
		k.MaxValidityAdmin = value
	}

	if value, err := cmd.PersistentFlags().GetDuration(FlagMaxValidityCA); err != nil {
		if err != nil {
			return fmt.Errorf("error parsing %s '%s': %s", FlagMaxValidityCA, value, err)
		}
		k.MaxValidityCA = value
	}

	// TODO: ensure CA >> COMPONENTS/ADMIN

	if len(args) > 0 {
		k.clusterID = args[0]
	} else {
		return errors.New("no cluster id was given")
	}

	return k.Ensure()
}
