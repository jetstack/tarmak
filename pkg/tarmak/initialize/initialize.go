// Copyright Jetstack Ltd. See LICENSE for details.
package initialize

import (
	"fmt"
	"io"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/sirupsen/logrus"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/cluster"
	"github.com/jetstack/tarmak/pkg/tarmak/environment"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/provider"
	"github.com/jetstack/tarmak/pkg/tarmak/utils/input"
)

var _ interfaces.Initialize = &Initialize{}

type Initialize struct {
	log    *logrus.Entry
	tarmak interfaces.Tarmak

	currentEnvironment interfaces.Environment
	currentProvider    interfaces.Provider

	input *input.Input
}

func New(t interfaces.Tarmak, in io.ReadCloser, out io.Writer) *Initialize {
	return &Initialize{
		log:    t.Log(),
		tarmak: t,
		input:  input.New(in, out),
	}
}

func (i *Initialize) newProvider(providerConf *tarmakv1alpha1.Provider) (interfaces.Provider, error) {
	return provider.NewFromConfig(i.tarmak, providerConf)
}

func (i *Initialize) newEnvironment(environmentConf *tarmakv1alpha1.Environment) (interfaces.Environment, error) {
	return environment.NewFromConfig(i.tarmak, environmentConf, i.Tarmak().Config().Clusters(environmentConf.Name))
}

func (i *Initialize) newCluster(clusterConf *clusterv1alpha1.Cluster) (interfaces.Cluster, error) {
	return cluster.NewFromConfig(i.CurrentEnvironment(), clusterConf)
}

func (i *Initialize) InitProvider() (providerObj interfaces.Provider, err error) {

creationLoop:
	for {
		providerConf, err := provider.Init(i)
		if err != nil {
			return nil, err
		}

		// create provider from config
		providerObj, err = i.newProvider(providerConf)
		if err != nil {
			i.input.Warn("error creating provider: ", err)
			continue
		}

		// show provider options
		question := []string{""}
		for key, value := range providerObj.Parameters() {
			question = append(question, fmt.Sprintf("%s: %s", key, value))
		}
		question = append(question, "\nContinue with this provider?")

		persist, err := i.input.AskYesNo(&input.AskYesNo{
			Default: true,
			Query:   strings.Join(question, "\n"),
		})
		if err != nil {
			return nil, err
		}

		if !persist {
			i.input.Warn("aborting provider creation")
			continue
		}

		for {
			err := providerObj.Validate()
			if err != nil {
				i.input.Warnf("validation failed: %s", err)

				choice, err := i.input.AskSelection(&input.AskSelection{
					Query: "What do you want to do now?",
					Choices: []string{
						"Retry validation",
						"Ignore validation error",
						"Start over again with provider creation",
					},
					Default: 0,
				})
				if err != nil {
					return nil, err
				}

				if choice == 0 {
					continue
				}

				if choice == 2 {
					continue creationLoop
				}

			}

			err = i.tarmak.Config().AppendProvider(providerConf)
			if err != nil {
				return nil, err
			}
			break
		}
		break
	}

	return providerObj, nil
}

func (i *Initialize) CurrentEnvironment() interfaces.Environment {
	return i.currentEnvironment
}

func (i *Initialize) CurrentProvider() interfaces.Provider {
	return i.currentProvider
}

func (i *Initialize) InitEnvironment() (environmentObj interfaces.Environment, err error) {
	providerObj, err := i.GetProvider()
	if err != nil {
		return nil, fmt.Errorf("error getting a provider: %s", err)
	}
	i.currentProvider = providerObj

creationLoop:
	for {
		var err error
		environmentConf, err := environment.Init(i)
		if err != nil {
			return nil, err
		}
		// create environment from config
		environmentObj, err = environment.NewFromConfig(i.tarmak, environmentConf, []*clusterv1alpha1.Cluster{})
		if err != nil {
			i.input.Warn("error creating environment: ", err)
			continue
		}

		// show environment options
		question := []string{""}
		for key, value := range environmentObj.Parameters() {
			question = append(question, fmt.Sprintf("%s: %s", key, value))
		}
		question = append(question, "\nContinue with this environment?")

		persist, err := i.input.AskYesNo(&input.AskYesNo{
			Default: true,
			Query:   strings.Join(question, "\n"),
		})
		if err != nil {
			return nil, err
		}

		if !persist {
			i.input.Warn("aborting environment creation")
			continue
		}

		for {
			err := environmentObj.Validate()
			if err != nil {
				i.input.Warnf("validation failed: %s", err)

				choice, err := i.input.AskSelection(&input.AskSelection{
					Query: "What do you want to do now?",
					Choices: []string{
						"retry validation",
						"ignore validation error",
						"start over again with environment creation",
					},
					Default: 0,
				})
				if err != nil {
					return nil, err
				}

				if choice == 0 {
					continue
				}

				if choice == 2 {
					continue creationLoop
				}

			}

			err = i.tarmak.Config().AppendEnvironment(environmentConf)
			if err != nil {
				return nil, err
			}
			break
		}
		break
	}

	return environmentObj, nil
}

func (i *Initialize) Input() *input.Input {
	return i.input
}

func (i *Initialize) Config() interfaces.Config {
	return i.tarmak.Config()
}

func (i *Initialize) Tarmak() interfaces.Tarmak {
	return i.tarmak
}

func (i *Initialize) AskProjectName() (projectName string, err error) {
	projectName, err = i.Input().AskOpen(&input.AskOpen{
		Query:   "What is the project name?",
		Default: "tarmak-playground",
	})
	if err != nil {
		return "", err
	}
	return projectName, nil
}

func (i *Initialize) AskContact() (contact string, err error) {
	for {
		contact, err = i.Input().AskOpen(&input.AskOpen{
			AllowEmpty: true,
			Query:      "Provide a contact mail address for the project administrator",
		})
		if err != nil {
			return "", err
		}

		if contact != "" {
			if err := validation.Validate(contact, validation.Required, is.Email); err != nil {
				i.Input().Warn("invalid contact mail address: ", err)
				continue
			}

		}

		break
	}
	return contact, nil
}

func (i *Initialize) InitCluster() (clusterObj interfaces.Cluster, err error) {
	environmentObj, err := i.GetEnvironment()
	if err != nil {
		return nil, fmt.Errorf("error getting the environment: %s", err)
	}
	i.currentEnvironment = environmentObj
	environmentObj.Provider().Reset()

	var clusterConf *clusterv1alpha1.Cluster

creationLoop:
	for {
		var err error
		clusterConf, err = cluster.Init(i)
		if err != nil {
			return nil, err
		}

		clusterObj, err = i.newCluster(clusterConf)
		if err != nil {
			return nil, err
		}

		// show cluster options
		question := []string{""}
		for key, value := range clusterObj.Parameters() {
			question = append(question, fmt.Sprintf("%s: %s", key, value))
		}
		question = append(question, "\nContinue with this cluster?")

		persist, err := i.input.AskYesNo(&input.AskYesNo{
			Default: true,
			Query:   strings.Join(question, "\n"),
		})
		if err != nil {
			return nil, err
		}

		if !persist {
			i.input.Warn("aborting cluster creation")
			continue
		}

		for {
			err := clusterObj.Validate()
			if err != nil {
				i.input.Warnf("validation failed: %s", err)

				choice, err := i.input.AskSelection(&input.AskSelection{
					Query: "What do you want to do now?",
					Choices: []string{
						"Retry validation",
						"Ignore validation error",
						"Start over again with cluster creation",
					},
					Default: 0,
				})
				if err != nil {
					return nil, err
				}

				if choice == 0 {
					continue
				}

				if choice == 2 {
					continue creationLoop
				}

			}

			err = i.tarmak.Config().AppendCluster(clusterConf)
			if err != nil {
				return nil, err
			}
			break
		}
		break
	}
	return clusterObj, nil
}

func (i *Initialize) GetProvider() (providerObj interfaces.Provider, err error) {
	if len(i.tarmak.Config().Providers()) == 0 {
		i.input.Warn("no providers found in configuration...\n")
		return i.InitProvider()
	} else {

		providers := i.tarmak.Providers()
		providerStrings := make([]string, len(providers))
		providerNames := make([]string, len(providers))
		for pos, provider := range providers {
			providerNames[pos] = provider.Name()
			providerStrings[pos] = provider.String()
		}

		providerPos, err := i.input.AskSelection(&input.AskSelection{
			Query:   "Select an existing provider or create a new one",
			Choices: append(providerStrings, "Create a new provider"),
			Default: len(providerNames),
		})
		if err != nil {
			return nil, err
		}

		// init new provider if this option has been selected
		if providerPos == len(providers) {
			return i.InitProvider()
		}

		return providers[providerPos], nil
	}
}

func (i *Initialize) GetEnvironment() (environmentObj interfaces.Environment, err error) {
	if len(i.tarmak.Config().Environments()) == 0 {
		i.input.Warn("no environments found in configuration...\n")
		return i.InitEnvironment()
	}

	for {
		environments := i.tarmak.Environments()
		environmentNames := make([]string, len(environments))
		environmentChoices := make([]string, len(environments))
		environmentSingle := make([]bool, len(environments))
		for pos, environment := range environments {
			environmentNames[pos] = environment.Name()
			environmentChoices[pos] = environment.Name()
			if environment.Type() == tarmakv1alpha1.EnvironmentTypeSingle {
				environmentSingle[pos] = true
				environmentChoices[pos] = fmt.Sprintf("%s (unavailable as single cluster env)", environmentChoices[pos])
			}
		}

		environmentPos, err := i.input.AskSelection(&input.AskSelection{
			Query:   "Select environment or create new",
			Choices: append(environmentChoices, "create new environment"),
			Default: len(environmentNames),
		})
		if err != nil {
			return nil, err
		}

		// init new environment if this option has been selected
		if environmentPos == len(environments) {
			return i.InitEnvironment()
		}

		if environmentSingle[environmentPos] {
			i.Input().Warn("you cannot add a cluster to a single cluster environment")
			continue
		}
		environmentObj = environments[environmentPos]
		break

	}
	return environmentObj, nil
}
