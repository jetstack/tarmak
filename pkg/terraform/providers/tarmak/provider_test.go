// Copyright Jetstack Ltd. See LICENSE for details.
package tarmak

import (
	"fmt"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/sirupsen/logrus"

	"github.com/jetstack/tarmak/pkg/tarmak/connector"
	"github.com/jetstack/tarmak/pkg/tarmak/mocks"
	"github.com/jetstack/tarmak/pkg/terraform/providers/tarmak/rpc"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"tarmak": testAccProvider,
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

type fakeCalls struct {
}

func (f *fakeCalls) Ping(args *rpc.PingArgs, reply *rpc.PingReply) error {
	reply.Version = "0.0.1"
	return nil
}

func (f *fakeCalls) BastionInstanceStatus(args *rpc.BastionInstanceStatusArgs, reply *rpc.BastionInstanceStatusReply) error {
	reply.Status = "test-status"
	return nil
}

type rpcServer struct {
	proxy      *connector.Proxy
	containerR io.Reader
	containerW io.Writer

	socketPath string

	ctrl       *gomock.Controller
	fakeTarmak *mocks.MockTarmak
}

func newRPCServer(t *testing.T) *rpcServer {
	r := &rpcServer{
		ctrl:       gomock.NewController(t),
		socketPath: fmt.Sprintf("tarmak-connector-%s.sock", randStringRunes(6)),
	}
	r.proxy = connector.NewProxy(r.socketPath)

	r.proxy.Reader, r.containerW = io.Pipe()
	r.containerR, r.proxy.Writer = io.Pipe()

	return r
}

func (r *rpcServer) Start() error {
	log := logrus.WithField("app", "test")
	r.proxy.Start()
	go rpc.Bind(log, &fakeCalls{}, r.containerR, r.containerW, &testClose{})
	return nil
}

func (r *rpcServer) Stop() error {
	r.proxy.Stop()
	return nil
}

type testClose struct {
}

func (t *testClose) Close() error {
	return nil
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
