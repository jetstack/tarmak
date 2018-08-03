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
	wingclientv1alpha1 "github.com/jetstack/tarmak/pkg/wing/client/typed/wing/v1alpha1"
)

func (c *Cluster) ensureWingClientset() error {
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
			return err
		}
	}

	return nil
}

func (c *Cluster) wingInstanceClient() (wingclientv1alpha1.InstanceInterface, error) {
	err := c.ensureWingClientset()
	if err != nil {
		return nil, err
	}

	return c.wingClientset.WingV1alpha1().Instances(c.ClusterName()), nil
}

func (c *Cluster) wingJobClient() (wingclientv1alpha1.WingJobInterface, error) {
	err := c.ensureWingClientset()
	if err != nil {
		return nil, err
	}

	return c.wingClientset.WingV1alpha1().WingJobs(c.ClusterName()), nil
}

func (c *Cluster) listInstances() (instances []*wingv1alpha1.Instance, err error) {
	// connect to wing
	client, err := c.wingInstanceClient()
	if err != nil {
		return instances, fmt.Errorf("failed to connect to wing API on bastion: %s", err)
	}

	// list all instances in Provider
	providerInstances, err := c.ListHosts()
	providerInstaceMap := make(map[string]interfaces.Host)
	if err != nil {
		return instances, fmt.Errorf("failed to list provider's instances: %s", err)
	}

	for pos, _ := range providerInstances {
		providerInstaceMap[providerInstances[pos].ID()] = providerInstances[pos]
	}

	// list all instances in wing
	wingInstances, err := client.List(metav1.ListOptions{})
	if err != nil {
		return instances, err
	}

	// loop through instances
	for pos, _ := range wingInstances.Items {
		instance := &wingInstances.Items[pos]

		// removes instances not in AWS
		if _, ok := providerInstaceMap[instance.Name]; !ok {
			c.log.Debugf("deleting unused instance %s in wing API", instance.Name)
			if err := client.Delete(instance.Name, &metav1.DeleteOptions{}); err != nil {
				c.log.Warnf("error deleting instance %s in wing API: %s", instance.Name, err)
			}
			continue
		}
		instances = append(instances, instance)
	}

	return instances, nil

}

func (c *Cluster) checkAllInstancesConverged(byState map[wingv1alpha1.InstanceManifestState][]*wingv1alpha1.Instance) error {
	instancesNotConverged := []*wingv1alpha1.Instance{}
	for key, instances := range byState {
		if len(instances) == 0 {
			continue
		}
		if key != wingv1alpha1.InstanceManifestStateConverged {
			instancesNotConverged = append(instancesNotConverged, instances...)
		}
		c.Log().Debugf("%d instances in state %s: %s", len(instances), key, outputInstances(instances))
	}

	if len(instancesNotConverged) > 0 {
		return fmt.Errorf("not all instances have converged yet %s", outputInstances(instancesNotConverged))
	}

	return nil
}

func (c *Cluster) checkAllJobsCompleted(jobs []wingv1alpha1.WingJob) error {
	jobsNotCompleted := []wingv1alpha1.WingJob{}
	for _, job := range jobs {
		if !job.Status.Completed {
			jobsNotCompleted = append(jobsNotCompleted, job)
		}
	}

	c.Log().Debugf("%d jobs not completed: %s", len(jobs), outputWingJobs(jobs))

	if len(jobsNotCompleted) > 0 {
		return fmt.Errorf("not all jobs have completed yet %s", outputWingJobs(jobsNotCompleted))
	}

	return nil
}

func outputInstances(instances []*wingv1alpha1.Instance) string {
	var output []string
	for _, instance := range instances {
		output = append(output, instance.Name)
	}
	return strings.Join(output, ", ")
}

func outputWingJobs(jobs []wingv1alpha1.WingJob) string {
	var output []string
	for _, job := range jobs {
		output = append(output, job.Name)
	}
	return strings.Join(output, ", ")
}

func (c *Cluster) listJobsForInstance(instance *wingv1alpha1.Instance) ([]wingv1alpha1.WingJob, error) {
	err := c.ensureWingClientset()
	if err != nil {
		return nil, err
	}

	client, err := c.wingJobClient()
	if err != nil {
		return nil, err
	}

	list, err := client.List(metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.instanceName=%s", instance.Name),
	})

	if err != nil {
		return nil, err
	}

	return list.Items, err
}
