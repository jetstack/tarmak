// Copyright Jetstack Ltd. See LICENSE for details.
package terraform

import (
	"github.com/Sirupsen/logrus"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type ConnectorContainer struct {
	t      *Terraform
	log    *logrus.Entry
	tarmak interfaces.Tarmak
}

func (t *Terraform) NewConnector() *ConnectorContainer {
	c := &ConnectorContainer{
		t:   t,
		log: t.log,
	}

	return c
}
