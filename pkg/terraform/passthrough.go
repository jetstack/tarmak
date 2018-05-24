// Copyright Jetstack Ltd. See LICENSE for details.
package terraform

import (
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/hashicorp/terraform/command"
	"github.com/hashicorp/terraform/plugin"
	"github.com/mitchellh/cli"
	provideraws "github.com/terraform-providers/terraform-provider-aws/aws"
	providerrandom "github.com/terraform-providers/terraform-provider-random/random"
	providertemplate "github.com/terraform-providers/terraform-provider-template/template"
	providertls "github.com/terraform-providers/terraform-provider-tls/tls"

	providerawstag "github.com/jetstack/tarmak/pkg/terraform/providers/awstag"
	providertarmak "github.com/jetstack/tarmak/pkg/terraform/providers/tarmak"
)

var InternalProviders = map[string]plugin.ProviderFunc{
	"aws":      provideraws.Provider,
	"random":   providerrandom.Provider,
	"tls":      providertls.Provider,
	"tarmak":   providertarmak.Provider,
	"awstag":   providerawstag.Provider,
	"template": providertemplate.Provider,
}

// create new terraform ui
func newUI(out io.Writer, err io.Writer) cli.Ui {

	outPrefix := ""
	errPrefix := ""

	return &cli.PrefixedUi{
		AskPrefix:    outPrefix,
		OutputPrefix: outPrefix,
		InfoPrefix:   outPrefix,
		ErrorPrefix:  errPrefix,
		Ui:           &cli.BasicUi{Writer: out, ErrorWriter: err},
	}
}

func newErrUI(out io.Writer, errOut io.Writer) cli.Ui {

	outPrefix := "OUT"
	errPrefix := "ERR"

	return &cli.PrefixedUi{
		AskPrefix:    outPrefix,
		OutputPrefix: outPrefix,
		InfoPrefix:   outPrefix,
		ErrorPrefix:  errPrefix,
		Ui:           &cli.BasicUi{Writer: out, ErrorWriter: errOut},
	}
}

func newMeta(ui cli.Ui) command.Meta {
	if os.Getenv("TF_LOG") == "" {
		log.SetOutput(ioutil.Discard)
		os.Stderr = nil
	}

	var inAutomation bool
	if v := os.Getenv("TF_IN_AUTOMATION"); v != "" {
		inAutomation = true
	}

	dataDir := os.Getenv("TF_DATA_DIR")

	return command.Meta{
		Color: true,
		Ui:    ui,

		RunningInAutomation: inAutomation,
		OverrideDataDir:     dataDir,
	}
}

// passthrough parameters
func InternalPlugin(args []string) int {
	if os.Getenv("TF_LOG") == "" {
		log.SetOutput(ioutil.Discard)
	}

	if len(args) != 2 {
		log.Printf("Wrong number of args; expected: terraform internal-plugin pluginType pluginName")
		return 1
	}

	pluginType := args[0]
	pluginName := args[1]

	if pluginType == "provider" {
		pluginFunc, found := InternalProviders[pluginName]
		if !found {
			log.Printf("[ERROR] Could not load provider: %s", pluginName)
			return 1
		}
		log.Printf("[INFO] Starting provider plugin %s", pluginName)
		plugin.Serve(&plugin.ServeOpts{
			ProviderFunc: pluginFunc,
		})
	} else {
		c := &command.InternalPluginCommand{}
		return c.Run(args)
	}

	return 0
}

func Plan(args []string) int {
	c := &command.PlanCommand{
		Meta: newMeta(newUI(os.Stdout, os.Stderr)),
	}
	return c.Run(args)
}

func Apply(args []string) int {
	c := &command.ApplyCommand{
		Meta: newMeta(newUI(os.Stdout, os.Stderr)),
	}
	return c.Run(args)
}

func Destroy(args []string) int {
	c := &command.ApplyCommand{
		Meta:    newMeta(newUI(os.Stdout, os.Stderr)),
		Destroy: true,
	}
	return c.Run(args)
}

func Init(args []string) int {
	c := &command.InitCommand{
		Meta: newMeta(newUI(os.Stdout, os.Stderr)),
	}
	return c.Run(args)
}

func Output(args []string) int {
	c := &command.OutputCommand{
		Meta: newMeta(newUI(os.Stdout, os.Stderr)),
	}
	return c.Run(args)
}

func Unlock(args []string) int {
	c := &command.UnlockCommand{
		Meta: newMeta(newUI(os.Stdout, os.Stderr)),
	}
	return c.Run(args)
}
