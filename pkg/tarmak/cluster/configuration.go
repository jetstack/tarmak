// Copyright Jetstack Ltd. See LICENSE for details.
package cluster

import (
	"bytes"
	"fmt"
	"time"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	wingv1alpha1 "github.com/jetstack/tarmak/pkg/apis/wing/v1alpha1"
)

const (
	retries = 100
)

// This upload the puppet.tar.gz to the cluster, warning there is some duplication as terraform is also uploading this puppet.tar.gz
func (c *Cluster) UploadConfiguration() error {

	buffer := new(bytes.Buffer)

	// get puppet config
	err := c.Environment().Tarmak().Puppet().TarGz(buffer)
	if err != nil {
		return err
	}

	// build reader from config
	reader := bytes.NewReader(buffer.Bytes())

	return c.Environment().Provider().UploadConfiguration(
		c,
		reader,
	)
}

// This enforces a reapply of the puppet.tar.gz on every instance in the cluster
func (c *Cluster) ReapplyConfiguration() error {
	c.log.Infof("making sure all instances apply the latest manifest")

	// connect to wing
	client, err := c.wingInstanceClient()
	if err != nil {
		return fmt.Errorf("failed to connect to wing API on bastion: %s", err)
	}

	// list instances
	instances, err := c.listInstances()
	if err != nil {
		return fmt.Errorf("failed to list instances: %s", err)
	}

	for pos, _ := range instances {
		instance := instances[pos]
		if instance.Spec == nil {
			instance.Spec = &wingv1alpha1.InstanceSpec{}
		}
		instance.Status.Converged = false

		if _, err := client.Update(instance); err != nil {
			c.log.Warnf("error updating instance %s in wing API: %s", instance.Name, err)
		}
	}

	// TODO: solve this on the API server side
	time.Sleep(time.Second * 5)

	return nil
}

// This waits until all instances have congverged successfully
func (c *Cluster) WaitForConvergance() error {
	c.log.Debugf("making sure all instances have converged using puppet")

	retries := retries
	for {
		instances, err := c.listInstances()
		if err != nil {
			return fmt.Errorf("failed to list instances: %s", err)
		}

		errors := []error{}

		for pos, _ := range instances {
			instance := instances[pos]

			// index by instance convergance state
			if instance.Status == nil || instance.Status.Converged == false {
				continue
			}

			jobs, err := c.listJobsForInstance(instance)
			if err != nil {
				return err
			}

			err = c.checkAllJobsCompleted(jobs)
			if err != nil {
				errors = append(errors, err)
			}

		}

		if len(errors) == 0 {
			c.log.Info("all instances converged")
			return nil
		} else {
			c.log.Debug(utilerrors.NewAggregate(errors))
		}

		retries--
		if retries == 0 {
			break
		}
		time.Sleep(time.Second * 5)

	}

	return fmt.Errorf("instances failed to converge in time")
}
