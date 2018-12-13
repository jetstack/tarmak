// Copyright Jetstack Ltd. See LICENSE for details.
package cluster

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

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
	hasher := md5.New()
	hasher.Write(buffer.Bytes())

	return c.Environment().Provider().UploadConfiguration(
		c,
		reader,
		hex.EncodeToString(hasher.Sum(nil)),
	)
}

// This enforces a reapply of the puppet.tar.gz on every machine in the cluster
func (c *Cluster) ReapplyConfiguration() error {
	c.log.Infof("making sure all machines apply the latest manifest")

	// connect to wing
	//client, err := c.wingMachineDeploymentClient()
	//if err != nil {
	//	return fmt.Errorf("failed to connect to wing API on bastion: %s", err)
	//}

	if err := c.deleteUnusedMachines(); err != nil {
		return err
	}

	if err := c.updateMachineDeployments(); err != nil {
		return err
	}

	//// list machine deployments
	_, err := c.listMachineDeployments()
	if err != nil {
		return fmt.Errorf("failed to list machines: %s", err)
	}

	//for pos, _ := range deployments {
	//	deployment := deployments[pos]
	//	if deployment.Spec == nil {
	//		//machine.Spec = &wingv1alpha1.MachineSpec{}
	//	}
	//	//machine.Spec.Converge = &wingv1alpha1.MachineSpecManifest{}

	//	if _, err := client.Update(deployment); err != nil {
	//		c.log.Warnf("error updating machine deployment %s in wing API: %s", deployment.Name, err)
	//	}
	//}

	//fmt.Printf("%+#v\n", deployments)

	// TODO: solve this on the API server side
	time.Sleep(time.Second * 5)

	return nil
}

// This waits until all machines have congverged successfully
func (c *Cluster) WaitForConvergance() error {
	c.log.Debugf("making sure all machine have converged using puppet")

	retries := retries
	for {
		deployments, err := c.listMachineDeployments()
		if err != nil {
			return fmt.Errorf("failed to list machines: %s", err)
		}

		var converged []*wingv1alpha1.MachineDeployment
		var converging []*wingv1alpha1.MachineDeployment
		for pos, _ := range deployments {
			deployment := deployments[pos]

			if deployment.Status == nil {
				converging = append(converging, deployment)
				continue
			}

			if deployment.Status.ReadyReplicas >= deployment.Status.Replicas &&
				deployment.Status.Replicas >= *deployment.Spec.MinReplicas {
				converged = append(converged, deployment)
				continue
			}

			converging = append(converging, deployment)
		}

		if len(converging) == 0 {
			c.log.Info("all deployments converged")
			return nil
		}

		c.log.Debug("--------")
		var convergedStr string
		for _, d := range converged {
			convergedStr = fmt.Sprintf("%s %s", convergedStr, d.Name)
		}
		if convergedStr != "" {
			c.log.Debugf("converged deployments [%s]", convergedStr)
		}

		for _, d := range converging {
			var readyReplicas int32
			if d.Status != nil {
				readyReplicas = d.Status.Replicas
			}
			c.log.Debugf("converging %s [%v/%v]", d.Name, readyReplicas, *d.Spec.MinReplicas)
		}

		retries--
		if retries == 0 {
			break
		}

		tok := time.Tick(time.Second * 5)

		select {
		case <-c.ctx.Done():
			return c.ctx.Err()
		case <-tok:
		}
	}

	return fmt.Errorf("machines failed to converge in time")
}
