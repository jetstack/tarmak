// Copyright Jetstack Ltd. See LICENSE for details.
package cluster

import (
	"fmt"
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
	for i := range machineDeploymentsList.Items {
		deployments = append(deployments, &machineDeploymentsList.Items[i])
	}

	return deployments, nil
}

func (c *Cluster) listMachines() ([]*wingv1alpha1.Machine, error) {

	client, err := c.wingMachineClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to wing API on bastion: %s", err)
	}

	mList, err := client.List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var machines []*wingv1alpha1.Machine
	for i := range mList.Items {
		machines = append(machines, &mList.Items[i])
	}

	return machines, nil
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

	for i := range deployments.Items {
		deploymentsMap[deployments.Items[i].Name] = &deployments.Items[i]
	}

	for _, i := range c.InstancePools() {
		if i.Role().Name() == "bastion" || i.Role().Name() == "vault" {
			continue
		}

		d, ok := deploymentsMap[i.Name()]
		if !ok {

			md := &wingv1alpha1.MachineDeployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      i.Role().Name(),
					Namespace: i.Config().ClusterName,
					Labels: map[string]string{
						"pool":    i.Name(),
						"cluster": c.ClusterName(),
					},
				},
				Spec: c.deploymentSpec(i),
			}

			_, err := client.Create(md)
			if err != nil {
				return fmt.Errorf("failed to create new deployment %s: %s", i.Name(), err)
			}

			c.log.Debugf("created new machine deployment %s", i.Name())

			continue
		}

		d = d.DeepCopy()
		d.Spec = c.deploymentSpec(i)

		_, err = client.Update(d)
		if err != nil {
			return fmt.Errorf("failed to update deployment sepc %s: %s", i.Name(), err)
		}
	}

	return nil
}

func (c *Cluster) deploymentSpec(instancePool interfaces.InstancePool) *wingv1alpha1.MachineDeploymentSpec {
	return &wingv1alpha1.MachineDeploymentSpec{
		MaxReplicas:             utils.PointerInt32(instancePool.Config().MaxCount),
		MinReplicas:             utils.PointerInt32(instancePool.Config().MinCount),
		ProgressDeadlineSeconds: new(int32),
		RevisionHistoryLimit:    new(int32),
		Strategy:                &wingv1alpha1.MachineDeploymentStrategy{},
		Paused:                  false,
		Selector: metav1.LabelSelector{
			MatchLabels: map[string]string{
				"pool":    instancePool.Name(),
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
		m := &wingMachines.Items[pos]

		// removes machines not in AWS
		if _, ok := providerInstaceMap[m.Name]; !ok {
			c.log.Debugf("deleting unused machine %s in wing API", m.Name)
			if err := client.Delete(m.Name, &metav1.DeleteOptions{}); err != nil {
				c.log.Warnf("error deleting machine %s in wing API: %s", m.Name, err)
			}
			continue
		}
	}

	return nil
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
