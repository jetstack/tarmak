// Copyright Jetstack Ltd. See LICENSE for details.
package tarmak

import (
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

func (t *Tarmak) SSH() interfaces.SSH {
	return t.ssh
}

func (t *Tarmak) SSHPassThrough(argsAdditional []string) {
	if err := t.ssh.WriteConfig(); err != nil {
		t.log.Fatal(err)
	}
	t.ssh.PassThrough(argsAdditional)
}
