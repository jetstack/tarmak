// Copyright Jetstack Ltd. See LICENSE for details.
package connector

import (
	"fmt"
	"net"
	"net/rpc"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"

	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

type Connector struct {
	log *logrus.Entry

	client *rpc.Client
	server net.Listener

	stopCh chan struct{}
}

func NewConnector(log *logrus.Entry) *Connector {
	return &Connector{
		log:    log,
		stopCh: utils.BasicSignalHandler(log),
	}
}

func (c *Connector) InitiateConnection() error {
	if err := c.NewClient(); err != nil {
		return err
	}

	if err := c.NewServer(); err != nil {
		return err
	}

	reply, err := c.CallHandshake()
	if err != nil {
		return err
	}

	conn, err := c.AcceptProvider()
	if err != nil {
		return err
	}

	return c.SendProvider(conn, reply)
}

func (c *Connector) StartConnector() error {
	var result *multierror.Error

	if err := c.InitiateConnection(); err != nil {
		return fmt.Errorf("error initialising connection: %v", err)
	}

	go c.StartServer()

	<-c.stopCh

	if err := c.CloseClient(); err != nil {
		result = multierror.Append(result, err)
	}

	if err := c.CloseServer(); err != nil {
		result = multierror.Append(result, err)
	}

	return result.ErrorOrNil()
}
