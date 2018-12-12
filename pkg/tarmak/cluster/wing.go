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
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
	wingclient "github.com/jetstack/tarmak/pkg/wing/client/clientset/versioned"
	wingclientv1alpha1 "github.com/jetstack/tarmak/pkg/wing/client/clientset/versioned/typed/wing/v1alpha1"
)

func (c *Cluster) listMachineDeployments() ([]*wingv1alpha1.MachineDeployment, error) {

	client, err := c.wingMachineDeploymentClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to wing API on bastion: %s", err)
	}

	machineDeploymentsList, err := client.List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var deployments []*wingv1alpha1.MachineDeployment
	for _, d := range machineDeploymentsList.DeepCopy().Items {
		deployments = append(deployments, &d)
	}

	return deployments, nil
}

func (c *Cluster) updateMachineDeployments() error {
	client, err := c.wingMachineDeploymentClient()
	if err != nil {
		return fmt.Errorf("failed to connect to wing API on bastion: %s", err)
	}

	// list all deployments in wing
	deployments, err := client.List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	deploymentsMap := make(map[string]*wingv1alpha1.MachineDeployment)
	if err != nil {
		return fmt.Errorf("failed to list provider's machines: %s", err)
	}

	for _, d := range deployments.DeepCopy().Items {
		deploymentsMap[d.Name] = &d
	}

	for _, i := range c.InstancePools() {
		if i.Role().Name() == "bastion" {
			continue
		}

		spec := c.deploymentSpec(i)
		d, ok := deploymentsMap[i.Role().Name()]
		if !ok {

			dep := &wingv1alpha1.MachineDeployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      i.Role().Name(),
					Namespace: i.Config().ClusterName,
					Labels: map[string]string{
						"role":    i.Role().Name(),
						"cluster": c.ClusterName(),
					},
				},
				Spec: c.deploymentSpec(i),
			}
			_, err := client.Create(dep)
			if err != nil {
				return fmt.Errorf("failed to create new deployment %s: %s", i.Role().Name(), err)
			}

			c.log.Infof("created new machine deployment %s", i.Role().Name())

			continue
		}

		d = d.DeepCopy()
		d.Spec = spec

		_, err := client.Update(d)
		if err != nil {
			return fmt.Errorf("failed to update deployment sepc %s: %s", i.Role().Name(), err)
		}
	}

	return nil
}

func (c *Cluster) deploymentSpec(instancePool interfaces.InstancePool) *wingv1alpha1.MachineDeploymentSpec {
	a := int32(instancePool.Config().MaxCount)
	b := int32(instancePool.Config().MinCount)
	return &wingv1alpha1.MachineDeploymentSpec{
		//MaxReplicas:     utils.PointerInt32(instancePool.Config().MaxCount),
		//MinReplicas:     utils.PointerInt32(instancePool.Config().MinCount),
		MaxReplicas:     &a,
		MinReplicas:     &b,
		MinReadySeconds: utils.PointerInt32(1000),
		Paused:          false,
		Selector: metav1.LabelSelector{
			MatchLabels: map[string]string{
				"role":    instancePool.Role().Name(),
				"cluster": c.ClusterName(),
			},
		},
	}
}

func (c *Cluster) deleteUnusedMachines() error {
	// connect to wing
	client, err := c.wingMachineClient()
	if err != nil {
		return fmt.Errorf("failed to connect to wing API on bastion: %s", err)
	}

	// list all machines in Provider
	providerMachines, err := c.ListHosts()
	providerInstaceMap := make(map[string]interfaces.Host)
	if err != nil {
		return fmt.Errorf("failed to list provider's machines: %s", err)
	}

	for pos, _ := range providerMachines {
		providerInstaceMap[providerMachines[pos].ID()] = providerMachines[pos]
	}

	// list all machines in wing
	wingMachines, err := client.List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	// loop through machines
	for pos, _ := range wingMachines.Items {
		machine := &wingMachines.Items[pos]
		fmt.Printf("%+#v\n", machine)

		// removes machines not in AWS
		if _, ok := providerInstaceMap[machine.Name]; !ok {
			c.log.Debugf("deleting unused machine %s in wing API", machine.Name)
			if err := client.Delete(machine.Name, &metav1.DeleteOptions{}); err != nil {
				c.log.Warnf("error deleting machine %s in wing API: %s", machine.Name, err)
			}
			continue
		}
	}

	return nil
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

func (c *Cluster) wingClient() (*wingclient.Clientset, error) {
	if c.wingClientset != nil {
		return c.wingClientset, nil
	}

	// connect to wing
	var err error
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

	return c.wingClientset, nil
}

func (c *Cluster) wingMachineClient() (wingclientv1alpha1.MachineInterface, error) {
	client, err := c.wingClient()
	if err != nil {
		return nil, err
	}

	return client.WingV1alpha1().Machines(c.ClusterName()), nil
}

func (c *Cluster) wingMachineDeploymentClient() (wingclientv1alpha1.MachineDeploymentInterface, error) {
	client, err := c.wingClient()
	if err != nil {
		return nil, err
	}

	return client.WingV1alpha1().MachineDeployments(c.ClusterName()), nil
}
