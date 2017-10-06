// Copyright Jetstack Ltd. See LICENSE for details.
package provider

import (
	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/provider/amazon"
	"github.com/jetstack/tarmak/pkg/tarmak/utils/input"
)

func Init(init interfaces.Initialize) (provider *tarmakv1alpha1.Provider, err error) {
	provider = &tarmakv1alpha1.Provider{}

	for {
		providerName, err := init.Input().AskOpen(&input.AskOpen{
			Query: "Enter a unique name for this provider [a-z0-9-]+",
		})
		if err != nil {
			return nil, err
		}

		nameValid := input.RegexpName.MatchString(providerName)
		nameUnique := init.Config().UniqueProviderName(providerName) == nil

		if !nameValid {
			init.Input().Warnf("provider name is not valid: %s", providerName)
		}

		if !nameUnique {
			init.Input().Warnf("provider name is not unique: %s", providerName)
		}

		if nameValid && nameUnique {
			provider.Name = providerName
			break
		}
	}

providerloop:
	for {
		clouds := []string{clusterv1alpha1.CloudAmazon, clusterv1alpha1.CloudAzure}
		cloud, err := init.Input().AskSelection(&input.AskSelection{
			Query:   "Select a cloud",
			Choices: clouds,
			Default: 0,
		})
		if err != nil {
			return nil, err
		}

		switch clouds[cloud] {
		case clusterv1alpha1.CloudAmazon:
			err := amazon.Init(init.Input(), provider)
			if err != nil {
				return nil, err
			}
			break providerloop
		default:
			init.Input().Warn("unsupported cloud provider: ", clouds[cloud])
		}
	}

	return provider, nil
}
