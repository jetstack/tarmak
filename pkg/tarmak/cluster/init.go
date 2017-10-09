// Copyright Jetstack Ltd. See LICENSE for details.
package cluster

import (
	"errors"
	"fmt"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/utils/input"
)

func Init(init interfaces.Initialize) (cluster *clusterv1alpha1.Cluster, err error) {
	environment := init.CurrentEnvironment()
	if environment == nil {
		return nil, errors.New("no environment given")
	}

	var clusterType string

	// determine cluster type to add
	if environment.Type() == tarmakv1alpha1.EnvironmentTypeSingle {
		return nil, fmt.Errorf("can't add a cluster to the single cluster environment '%s'", environment.Name())
	}
	if environment.Type() == tarmakv1alpha1.EnvironmentTypeEmpty {
		clusterType, err = askClusterType(init)
		if err != nil {
			return nil, err
		}
	}

	// add multi cluster
	if environment.Type() == tarmakv1alpha1.EnvironmentTypeMulti {
		clusterType = tarmakv1alpha1.EnvironmentTypeMulti
	}

	// add single cluster

	if clusterType == tarmakv1alpha1.EnvironmentTypeMulti {
		clusterName, err := askClusterName(init)
		if err != nil {
			return nil, err
		}
		cluster = config.NewClusterMulti(environment.Name(), clusterName)

		// adds hub if neccessary
		hubExists := false
		for _, cluster := range init.Config().Clusters(environment.Name()) {
			if cluster.Name == "hub" {
				hubExists = true
				break
			}
		}
		if !hubExists {
			err := init.Config().AppendCluster(config.NewHub(environment.Name()))
			if err != nil {
				return nil, err
			}
		}
	}

	if clusterType == tarmakv1alpha1.EnvironmentTypeSingle {
		cluster = config.NewClusterSingle(environment.Name(), "cluster")
	}

	availabilityZones, err := init.CurrentEnvironment().Provider().AskInstancePoolZones(init)
	if err != nil {
		return nil, err
	}
	addAvailabilityZones(cluster, availabilityZones)

	return cluster, nil
}

func askClusterName(init interfaces.Initialize) (clusterName string, err error) {
	for {
		clusterName, err = init.Input().AskOpen(&input.AskOpen{
			Query: "Enter a unique name for this cluster [a-z0-9-]+",
		})
		if err != nil {
			return "", err
		}

		nameValid := input.RegexpName.MatchString(clusterName)
		nameUnique := init.Config().UniqueProviderName(clusterName) == nil

		if clusterName == clusterv1alpha1.ClusterTypeHub {
			init.Input().Warnf("cluster name is not valid: %s", clusterName)
		}

		if !nameValid {
			init.Input().Warnf("cluster name is not valid: %s", clusterName)
		}

		if !nameUnique {
			init.Input().Warnf("cluster name is not unique: %s", clusterName)
		}

		if nameValid && nameUnique {
			break
		}
	}
	return clusterName, nil
}

func askClusterType(init interfaces.Initialize) (clusterType string, err error) {
	clusterTypePos, err := init.Input().AskSelection(&input.AskSelection{
		Query:   "What kind of cluster do you want to add?",
		Choices: []string{"Single cluster environment", "Multi-cluster environment"},
		Default: 0,
	})
	if err != nil {
		return "", err
	}

	if clusterTypePos == 0 {
		return tarmakv1alpha1.EnvironmentTypeSingle, nil
	}
	if clusterTypePos == 1 {
		return tarmakv1alpha1.EnvironmentTypeMulti, nil
	}

	return "", errors.New("no valid selection")
}

func addAvailabilityZones(cluster *clusterv1alpha1.Cluster, zones []string) {

	subnets := make([]*clusterv1alpha1.Subnet, len(zones))

	for i, zone := range zones {
		subnets[i] = &clusterv1alpha1.Subnet{
			Zone: zone,
		}
	}

	for i := range cluster.InstancePools {
		cluster.InstancePools[i].Subnets = subnets
	}
}
