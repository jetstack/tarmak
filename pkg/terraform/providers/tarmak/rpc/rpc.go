// Copyright Jetstack Ltd. See LICENSE for details.
package rpc

import (
	"net"
	"net/rpc"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

const (
	ConnectorSocket = "/tmp/tarmak-connector.sock"
	RPCName         = "Tarmak"
	Retries         = 60
)

type tarmakInterface struct{}

type tarmakRPC struct {
	cluster interfaces.Cluster
	tarmak  interfaces.Tarmak
}

func (r *tarmakRPC) log() *logrus.Entry {
	return r.tarmak.Log()
}

func New(cluster interfaces.Cluster) Tarmak {
	return &tarmakRPC{
		tarmak:  cluster.Environment().Tarmak(),
		cluster: cluster,
	}
}

type Tarmak interface {
	BastionInstanceStatus(*BastionInstanceStatusArgs, *BastionInstanceStatusReply) error
	VaultClusterStatus(*VaultClusterStatusArgs, *VaultClusterStatusReply) error
	VaultClusterInitStatus(*VaultClusterStatusArgs, *VaultClusterStatusReply) error
	VaultInstanceRole(*VaultInstanceRoleArgs, *VaultInstanceRoleReply) error
	Ping(*PingArgs, *PingReply) error
}

// listen to a unix socket
func ListenUnixSocket(log *logrus.Entry, tarmak Tarmak, socketPath string, stopCh chan struct{}) error {
	s := rpc.NewServer()
	s.RegisterName(RPCName, tarmak)
	log.Debugf("rpc server started")

	err := os.Remove(socketPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	unixListener, err := net.Listen("unix", socketPath)
	if err != nil {
		return err
	}

	go func() {
		<-stopCh
		err := unixListener.Close()
		if err != nil {
			log.Debugf("error stoppingn rpc server: %s", err)
		}
	}()

	go func() {
		for {
			fd, err := unixListener.Accept()
			if err != nil {
				log.Errorf("failed to accept unix socket: %s", err)
				break
			}

			// handle new connection in new go routine
			go s.ServeConn(fd)

			s.Accept(unixListener)
		}
		log.Debugf("rpc server stopped")
	}()

	return nil
}
