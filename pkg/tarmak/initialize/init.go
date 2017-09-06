package initialize

import (
	"fmt"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/tcnksm/go-input"

	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type Init struct {
	log    *logrus.Entry
	tarmak interfaces.Tarmak
}

func New(t interfaces.Tarmak) *Init {
	return &Init{
		log:    t.Log(),
		tarmak: t,
	}
}

func parseContextName(in string) (environment string, context string, err error) {
	in = strings.ToLower(in)

	splitted := false

	for i, c := range in {
		if !splitted && c == '-' {
			splitted = true
			environment = in[0:i]
			context = in[i+1 : len(in)]
		} else if c < '0' || (c > '9' && c < 'a') || c > 'z' {
			return "", "", fmt.Errorf("invalid char '%c' in string '%s' at pos %d", c, in, i)
		}
	}

	if !splitted {
		return "", "", fmt.Errorf("string '%s' did not contain '-'", in)
	}
	return environment, context, nil
}

func (i *Init) Run() error {
	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}

	var query string

	/* TODO: support multiple cluster in one env
	query = "What kind of cluster do you want to initialise?"
	options = []string{"create new single cluster environment", "create new multi cluster environment", "add new cluster to existing multi cluster environment"}
	kind, err := ui.Select(query, options, &input.Options{
		Default: options[0],
		Loop:    true,
		ValidateFunc: func(s string) error {
			if s != "1" {
				return fmt.Errorf(`option "%s" is currently not supported`, s)
			}
			return nil
		},
	})
	if err != nil {
		return err
	}
	*/

	query = "What should be the name of the cluster?\n\nThe name consists of two parts seperated by a dash. First part is the environment name, second part the cluster name. Both names should be matching [a-z0-9]+\n"
	combinedName, err := ui.Ask(query, &input.Options{
		Loop: true,
		ValidateFunc: func(s string) error {
			environment, context, err := parseContextName(s)
			if err != nil {
				return err
			}
			i.log.WithField("environment", environment).WithField("context", context).Debug("")
			// TODO verify environment name not taken yet
			// TODO ensure max length of both is not longer than 24 chars (verify that limit from AWS)
			return nil
		},
	})
	if err != nil {
		return err
	}
	environmentName, contextName, err := parseContextName(combinedName)
	if err != nil {
		return err
	}

	/* TODO: Support multiple providers
	query = "Which provider do you want to use?"
	options = []string{"AWS"}
	provider, err := ui.Ask(query, &input.Options{
		Loop:    true,
		Default: options[0],
	})
	if err != nil {
		return err
	}
	*/

	query = "Do you want to use vault to get credentials for AWS? [Y/N] "
	vault, err := ui.Ask(query, &input.Options{
		Required: true,
		Default:  "N",
		// Validate input
		ValidateFunc: func(s string) error {
			s = strings.ToLower(s)
			if s != "y" && s != "n" {
				return fmt.Errorf("input must be Y or N")
			}

			return nil
		},
	})
	if err != nil {
		return err
	}

	vaultPrefix := ""
	if s := strings.ToLower(vault); s == "y" {
		query = "Which path should be used for AWS credentials?"
		vaultPrefix, err = ui.Ask(query, &input.Options{
			Required: true,
			Default:  "jetstack/aws/jetstack-dev/sts/admin",
		})
		if err != nil {
			return err
		}
	}

	query = "Which region should be used?"
	awsRegion, err := ui.Ask(query, &input.Options{
		Required: true,
		Default:  "eu-west-1",
	})
	if err != nil {
		return err
	}
	// TODO: validate region
	// TODO: add dialog to allow custom AWS azs

	query = "What bucket prefix should be used?"
	bucketPrefix, err := ui.Ask(query, &input.Options{
		Required: true,
		Default:  "tarmak-",
	})
	if err != nil {
		return err
	}
	// TODO: verify bucket prefix [a-z0-9-_]

	query = "What public zone should be used?\n\nPlease make sure you can delegate this zone to AWS!\n"
	publicZone, err := ui.Ask(query, &input.Options{
		Required: true,
	})
	if err != nil {
		return err
	}
	// TODO: verify domain name

	query = "What private zone should be used?"
	privateZone, err := ui.Ask(query, &input.Options{
		Required: true,
		Default:  "tarmak.local",
	})
	if err != nil {
		return err
	}
	// TODO: verify domain name

	query = "What is the mail address of someone responsible?"
	contact, err := ui.Ask(query, &input.Options{
		Required: true,
		Default:  "",
	})
	if err != nil {
		return err
	}
	// TODO: use default from existing config
	// TODO: verify mail

	query = "What is the project name?"
	project, err := ui.Ask(query, &input.Options{
		Required: true,
		Default:  "k8s-playground",
	})
	if err != nil {
		return err
	}

	env := config.Environment{
		Contact: contact,
		Project: project,
		AWS: &config.AWSConfig{
			VaultPath: vaultPrefix,
			Region:    awsRegion,
		},
		Name: environmentName,
		Contexts: []config.Context{
			config.Context{
				Name:      contextName,
				BaseImage: "centos-puppet-agent",
				Stacks: []config.Stack{
					config.Stack{
						State: &config.StackState{
							BucketPrefix: bucketPrefix,
							PublicZone:   publicZone,
						},
					},
					config.Stack{
						Network: &config.StackNetwork{
							NetworkCIDR: "10.98.0.0/20",
							PrivateZone: privateZone,
						},
					},
					config.Stack{
						Tools: &config.StackTools{},
					},
					config.Stack{
						Vault: &config.StackVault{},
					},
					config.Stack{
						Kubernetes: &config.StackKubernetes{},
						NodeGroups: config.DefaultKubernetesNodeGroupAWSOneMasterThreeEtcdThreeWorker(),
					},
				},
			},
		},
	}

	return i.tarmak.MergeEnvironment(env)
}
