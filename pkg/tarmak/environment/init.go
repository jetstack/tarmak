// Copyright Jetstack Ltd. See LICENSE for details.
package environment

import (
	"errors"

	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/utils/input"
)

func Init(init interfaces.Initialize) (environment *tarmakv1alpha1.Environment, err error) {
	provider := init.CurrentProvider()
	if provider == nil {
		return nil, errors.New("error no provider given")
	}

	environment = &tarmakv1alpha1.Environment{
		Provider: provider.Name(),
	}

	environment.Name, err = askEnvironmentName(init)
	if err != nil {
		return nil, err
	}

	environment.Project, err = init.AskProjectName()
	if err != nil {
		return nil, err
	}

	environment.Contact, err = init.AskContact()
	if err != nil {
		return nil, err
	}

	environment.Location, err = provider.AskEnvironmentLocation(init)
	if err != nil {
		return nil, err
	}

	return environment, nil

}

func askEnvironmentName(init interfaces.Initialize) (environmentName string, err error) {
	for {
		environmentName, err = init.Input().AskOpen(&input.AskOpen{
			Query: "Enter a unique name for this environment [a-z0-9-]+",
		})
		if err != nil {
			return "", err
		}

		nameValid := input.RegexpName.MatchString(environmentName)
		nameUnique := init.Config().UniqueProviderName(environmentName) == nil

		if !nameValid {
			init.Input().Warnf("environment name is not valid: %s", environmentName)
		}

		if !nameUnique {
			init.Input().Warnf("environment name is not unique: %s", environmentName)
		}

		if nameValid && nameUnique {
			break
		}
	}
	return environmentName, nil
}
