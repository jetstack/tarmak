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
	client, err := c.wingMachineClient()
	if err != nil {
		return fmt.Errorf("failed to connect to wing API on bastion: %s", err)
	}

	// list machines
	machines, err := c.listMachines()
	if err != nil {
		return fmt.Errorf("failed to list machines: %s", err)
	}

	for pos, _ := range machines {
		machine := machines[pos]
		if machine.Spec == nil {
			machine.Spec = &wingv1alpha1.MachineSpec{}
		}
		machine.Spec.Converge = &wingv1alpha1.MachineSpecManifest{}

		if _, err := client.Update(machine); err != nil {
			c.log.Warnf("error updating machine %s in wing API: %s", machine.Name, err)
		}
	}

	// TODO: solve this on the API server side
	time.Sleep(time.Second * 5)

	return nil
}

// This waits until all machines have congverged successfully
func (c *Cluster) WaitForConvergance() error {
	c.log.Debugf("making sure all machine have converged using puppet")

	retries := retries
	for {
		machines, err := c.listMachines()
		if err != nil {
			return fmt.Errorf("failed to list machines: %s", err)
		}

		machineByState := make(map[wingv1alpha1.MachineManifestState][]*wingv1alpha1.Machine)

		for pos, _ := range machines {
			machine := machines[pos]

			// index by machine convergance state
			if machine.Status == nil || machine.Status.Converge == nil || machine.Status.Converge.State == "" {
				continue
			}

			state := machine.Status.Converge.State
			if _, ok := machineByState[state]; !ok {
				machineByState[state] = []*wingv1alpha1.Machine{}
			}

			machineByState[state] = append(
				machineByState[state],
				machine,
			)
		}

		err = c.checkAllMachinesConverged(machineByState)
		if err == nil {
			c.log.Info("all machines converged")
			return nil
		} else {
			c.log.Debug(err)
		}

		select {
		case <-c.ctx.Done():
			return c.ctx.Err()
		default:
		}

		retries--
		if retries == 0 {
			break
		}
		time.Sleep(time.Second * 5)

	}

	return fmt.Errorf("machines failed to converge in time")
}
