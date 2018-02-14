package vault_dev

import (
	"fmt"
	"net"
	"os/exec"
	"syscall"
	"time"

	vault "github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
)

const RootTokenDev = "root-token-dev"

type VaultDev struct {
	client       *vault.Client
	server       *exec.Cmd
	vaultRunning chan struct{}
	port         *int
}

func New() *VaultDev {
	return &VaultDev{}
}

func (v *VaultDev) Start() error {

	if v.port == nil {
		p := getUnusedPort()
		v.port = &p
	}

	args := []string{
		"server",
		"-dev",
		fmt.Sprintf("-dev-root-token-id=%s", RootTokenDev),
		fmt.Sprintf("-dev-listen-address=127.0.0.1:%d", *v.port),
	}

	logrus.Infof("starting vault: %#+v", args)

	v.server = exec.Command("vault", args...)

	err := v.server.Start()
	if err != nil {
		return err
	}

	// this channel will close once vault is stopped
	v.vaultRunning = make(chan struct{}, 0)

	go func() {
		err := v.server.Wait()
		if err != nil {
			logrus.Warn("vault stopped with error: ", err)

		} else {
			logrus.Info("vault stopped")
		}
		close(v.vaultRunning)
	}()

	v.client, err = vault.NewClient(&vault.Config{
		Address: fmt.Sprintf("http://127.0.0.1:%d", *v.port),
	})
	if err != nil {
		return err
	}
	v.client.SetToken(RootTokenDev)

	tries := 30
	for {
		select {
		case _, open := <-v.vaultRunning:
			if !open {
				return fmt.Errorf("vault could not be started")
			}
		default:
		}

		_, err := v.client.Auth().Token().LookupSelf()
		if err == nil {
			break
		}
		if tries <= 1 {
			return fmt.Errorf("vault dev server couldn't be started in time")
		}
		tries -= 1
		time.Sleep(time.Second)
	}

	return nil
}

func (v *VaultDev) Stop() {
	if err := v.server.Process.Signal(syscall.SIGTERM); err != nil {
		logrus.Warn("killing vault dev server failed: ", err)
	}

	<-v.vaultRunning
}

func (v *VaultDev) Client() *vault.Client {
	return v.client
}

func (v *VaultDev) SetPort(port int) {
	v.port = &port
}

func getUnusedPort() int {
	l, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 0,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}
