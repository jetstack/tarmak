// Copyright Jetstack Ltd. See LICENSE for details.
package connector

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
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
		stopCh: SignalHandler(log),
	}
}

func (c *Connector) InitiateConnection() error {
	if err := c.NewClient(); err != nil {
		return err
	}

	if err := c.NewServer(); err != nil {
		return err
	}

	reply, err := c.CallInit()
	if err != nil {
		return err
	}

	conn, err := c.AcceptProvider()
	if err != nil {
		return err
	}

	return c.SendProvider(conn, reply)
}

func (c *Connector) RunConnector() error {
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

func SignalHandler(log *logrus.Entry) chan struct{} {

	stopCh := make(chan struct{})
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	go func(log *logrus.Entry) {
		<-ch
		log.Infof("Connector received interupt. Shutting down...")
		close(stopCh)
		<-ch
		log.Infof("Force Closed.")
		os.Exit(1)
	}(log)

	return stopCh
}
