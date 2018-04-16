// Copyright Jetstack Ltd. See LICENSE for details.
package tarmak

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/sirupsen/logrus"

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
	socketPath string

	ctrl            *gomock.Controller
	fakeTarmak      *mocks.MockTarmak
	fakeEnvironment *mocks.MockEnvironment
	fakeCluster     *mocks.MockCluster

	stopCh    chan struct{}
	waitGroup sync.WaitGroup
}

func newRPCServer(t *testing.T) *rpcServer {
	r := &rpcServer{
		stopCh:     make(chan struct{}),
		ctrl:       gomock.NewController(t),
		socketPath: fmt.Sprintf("tarmak-connector-%s.sock", randStringRunes(6)),
	}

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	log := logger.WithField("app", "test")

	r.fakeTarmak = mocks.NewMockTarmak(r.ctrl)
	r.fakeEnvironment = mocks.NewMockEnvironment(r.ctrl)
	r.fakeCluster = mocks.NewMockCluster(r.ctrl)

	r.fakeEnvironment.EXPECT().Tarmak().AnyTimes().Return(r.fakeTarmak)
	r.fakeCluster.EXPECT().Environment().AnyTimes().Return(r.fakeEnvironment)
	r.fakeTarmak.EXPECT().Log().AnyTimes().Return(log)

	return r
}

func (r *rpcServer) Finish() {
	r.Stop()
	r.ctrl.Finish()
}

func (r *rpcServer) Start() error {
	t := rpc.New(r.fakeCluster)
	r.waitGroup.Add(1)
	go func() {
		defer r.waitGroup.Done()
		rpc.ListenUnixSocket(t, r.socketPath, r.stopCh)
	}()
	return nil
}

func (r *rpcServer) Stop() error {
	if r.stopCh != nil {
		close(r.stopCh)
		r.stopCh = nil
	}
	r.waitGroup.Wait()
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
