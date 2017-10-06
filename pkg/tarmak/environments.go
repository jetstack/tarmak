// Copyright Jetstack Ltd. See LICENSE for details.
package tarmak

import (
	"fmt"

	"github.com/jetstack/tarmak/pkg/tarmak/environment"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

func (t *Tarmak) newEnvironment(environmentName string) (interfaces.Environment, error) {
	environmentConfig, err := t.config.Environment(environmentName)
	if err != nil {
		return nil, fmt.Errorf("error finding environment '%s'", environmentName)
	}

	clusterConfigs := t.config.Clusters(environmentName)

	// init environment
	env, err := environment.NewFromConfig(t, environmentConfig, clusterConfigs)
	if err != nil {
		return nil, fmt.Errorf("error initializing environment '%s': %s", environmentName, err)
	}
	return env, nil
}

func (t *Tarmak) Environment() interfaces.Environment {
	if t.init != nil && t.init.CurrentEnvironment() != nil {
		return t.init.CurrentEnvironment()
	}
	return t.environment
}

func (t *Tarmak) Environments() (envs []interfaces.Environment) {
	for _, envConfig := range t.Config().Environments() {
		env, err := t.newEnvironment(envConfig.Name)
		if err != nil {
			t.log.Warnf("error listing environment '%s': %s", envConfig.Name, err)
			continue
		}
		envs = append(envs, env)
	}
	return envs
}
