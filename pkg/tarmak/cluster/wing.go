// Copyright Jetstack Ltd. See LICENSE for details.
package cluster

import (
	"fmt"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	wingv1alpha1 "github.com/jetstack/tarmak/pkg/apis/wing/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	wingclientv1alpha1 "github.com/jetstack/tarmak/pkg/wing/client/clientset/versioned/typed/wing/v1alpha1"
)

func (c *Cluster) wingMachineClient() (wingclientv1alpha1.MachineInterface, error) {
	var err error

	if c.wingClientset == nil {
		// connect to wing

		wingClientsetTry := func() error {
			c.wingClientset, c.wingTunnel, err = c.Environment().WingClientset()
			if err != nil {
				return fmt.Errorf("failed to connect to wing API on bastion: %s", err)
			}

			return nil
		}

		expBackoff := backoff.NewExponentialBackOff()
		expBackoff.InitialInterval = time.Second
		expBackoff.MaxElapsedTime = time.Minute * 2

		b := backoff.WithContext(expBackoff, context.Background())

		if err := backoff.Retry(wingClientsetTry, b); err != nil {
			return nil, err
		}

	}

	return c.wingClientset.WingV1alpha1().Machines(c.ClusterName()), nil
}

func (c *Cluster) listMachines() (machines []*wingv1alpha1.Machine, err error) {
	// connect to wing
	client, err := c.wingMachineClient()
	if err != nil {
		return machines, fmt.Errorf("failed to connect to wing API on bastion: %s", err)
	}

	// list all machines in Provider
	providerMachines, err := c.ListHosts()
	providerInstaceMap := make(map[string]interfaces.Host)
	if err != nil {
		return machines, fmt.Errorf("failed to list provider's machines: %s", err)
	}

	for pos, _ := range providerMachines {
		providerInstaceMap[providerMachines[pos].ID()] = providerMachines[pos]
	}

	// list all machines in wing
	wingMachines, err := client.List(metav1.ListOptions{})
	if err != nil {
		return machines, err
	}

	// loop through machines
	for pos, _ := range wingMachines.Items {
		machine := &wingMachines.Items[pos]

		// removes machines not in AWS
		if _, ok := providerInstaceMap[machine.Name]; !ok {
			c.log.Debugf("deleting unused machine %s in wing API", machine.Name)
			if err := client.Delete(machine.Name, &metav1.DeleteOptions{}); err != nil {
				c.log.Warnf("error deleting machine %s in wing API: %s", machine.Name, err)
			}
			continue
		}
		machines = append(machines, machine)
	}

	return machines, nil

}

func (c *Cluster) checkAllMachinesConverged(byState map[wingv1alpha1.MachineManifestState][]*wingv1alpha1.Machine) error {
	machinesNotConverged := []*wingv1alpha1.Machine{}
	for key, machines := range byState {
		if len(machines) == 0 {
			continue
		}
		if key != wingv1alpha1.MachineManifestStateConverged {
			machinesNotConverged = append(machinesNotConverged, machines...)
		}
		c.Log().Debugf("%d machines in state %s: %s", len(machines), key, outputMachines(machines))
	}

	if len(machinesNotConverged) > 0 {
		return fmt.Errorf("not all machines have converged yet %s", outputMachines(machinesNotConverged))
	}

	return nil
}

func outputMachines(machines []*wingv1alpha1.Machine) string {
	var output []string
	for _, machine := range machines {
		output = append(output, machine.Name)
	}
	return strings.Join(output, ", ")
}
