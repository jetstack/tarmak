// Copyright Jetstack Ltd. See LICENSE for details.
package tarmak

import (
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

func (t *Tarmak) SSH() interfaces.SSH {
	return t.ssh
}

func (t *Tarmak) SSHPassThrough(argsAdditional []string) error {
	if err := t.ssh.WriteConfig(t.Cluster()); err != nil {
		return err
	}

	if err := t.ssh.Validate(); err != nil {
		return err
	}

	if err := t.ssh.PassThrough(argsAdditional); err != nil {
		return err
	}

	return nil
}
