// Copyright Jetstack Ltd. See LICENSE for details.
package tarmak

import (
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

func (t *Tarmak) SSH() interfaces.SSH {
	return t.ssh
}

func (t *Tarmak) SSHPassThrough(host string, argsAdditional []string) error {
	if err := t.ssh.WriteConfig(t.Cluster()); err != nil {
		return err
	}

	if err := t.ssh.Validate(); err != nil {
		return err
	}

	if err := t.ssh.PassThrough(host, argsAdditional); err != nil {
		return err
	}

	return nil
}
