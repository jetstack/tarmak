// Copyright Jetstack Ltd. See LICENSE for details.
package tarmak

import (
	"fmt"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/provider"
)

func (t *Tarmak) newProvider(providerName string) (interfaces.Provider, error) {
	providerConfig, err := t.config.Provider(providerName)
	if err != nil {
		return nil, fmt.Errorf("error finding provider '%s'", providerName)
	}

	prov, err := provider.NewFromConfig(t, providerConfig)
	if err != nil {
		return nil, fmt.Errorf("error initializing provider '%s': %s", providerName, err)
	}
	return prov, nil
}

func (t *Tarmak) Provider() interfaces.Provider {
	return t.Environment().Provider()
}

func (t *Tarmak) Providers() (providers []interfaces.Provider) {
	for _, provConfig := range t.Config().Providers() {
		prov, err := t.newProvider(provConfig.Name)
		if err != nil {
			t.log.Warnf("error listing provider '%s': %s", provConfig.Name, err)
			continue
		}
		providers = append(providers, prov)
	}
	return providers
}
