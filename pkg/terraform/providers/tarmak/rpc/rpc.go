// Copyright Jetstack Ltd. See LICENSE for details.
package rpc

import (
	"io"
	"net/rpc"

	"github.com/alecthomas/multiplex"
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
	tarmak interfaces.Tarmak
	stack  interfaces.Stack
}

func (r *tarmakRPC) log() *logrus.Entry {
	return r.tarmak.Log()
}

func NewTarmak(tarmak interfaces.Tarmak, stack interfaces.Stack) Tarmak {
	return &tarmakRPC{tarmak: tarmak, stack: stack}
}

type Tarmak interface {
	BastionInstanceStatus(*BastionInstanceStatusArgs, *BastionInstanceStatusReply) error
	VaultClusterStatus(*VaultClusterStatusArgs, *VaultClusterStatusReply) error
	VaultClusterInitStatus(*VaultClusterStatusArgs, *VaultClusterStatusReply) error
	VaultInstanceRole(*VaultInstanceRoleArgs, *VaultInstanceRoleReply) error
	Ping(*PingArgs, *PingReply) error
}

// bind a new rpc server to socket
func Bind(log *logrus.Entry, tarmak Tarmak, reader io.Reader, writer io.Writer, closer io.Closer) {

	s := rpc.NewServer()
	s.RegisterName(RPCName, tarmak)

	log.Debugf("rpc server started")

	mx := multiplex.MultiplexedServer(struct {
		io.Reader
		io.Writer
		io.Closer
	}{reader, writer, closer},
	)

	for {
		c, err := mx.Accept()
		if err != nil {
			log.Warnf("error accepting rpc connection: %s", err)
			break
		}
		go func(c *multiplex.Channel) {
			log.Debugf("new rpc connection")
			s.ServeConn(c)
			log.Debugf("closed rpc connection")
		}(c)
	}

	log.Debugf("rpc server stopped")
}
