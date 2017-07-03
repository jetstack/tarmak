package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/terraform"
)

// tfdestroyCmd represents the tfdestroy command
var tfdestroyCmd = &cobra.Command{
	Use:   "tfdestroy",
	Short: "This applies the set of stacks in the current context",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("tfdestroy called")

		myConfig := config.DefaultConfigSingle()

		err := myConfig.Validate()
		if err != nil {
			log.Fatal(err)
		}

		context, err := myConfig.GetContext()
		if err != nil {
			log.Fatal(err)
		}
		log.Debugf("current context: %#+v", context)

		tf := terraform.New(nil, context)

		for posStack, _ := range context.Stacks {
			err = tf.Destroy(&context.Stacks[len(context.Stacks)-posStack-1])
			if err != nil {
				log.Fatal(err)
			}
		}

	},
}

func init() {
	RootCmd.AddCommand(tfdestroyCmd)
}
