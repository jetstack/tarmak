// Copyright Jetstack Ltd. See LICENSE for details.
package stack

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/cenkalti/backoff"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type ToolsStack struct {
	*Stack
}

var _ interfaces.Stack = &ToolsStack{}

func newToolsStack(s *Stack) (*ToolsStack, error) {
	t := &ToolsStack{
		Stack: s,
	}

	s.roles = make(map[string]bool)
	s.roles[clusterv1alpha1.InstancePoolTypeJenkins] = true
	s.roles[clusterv1alpha1.InstancePoolTypeBastion] = true

	s.name = tarmakv1alpha1.StackNameTools
	s.verifyPreDeploy = append(s.verifyPostDeploy, s.verifyImageIDs)
	s.verifyPostDeploy = append(s.verifyPostDeploy, t.VerifyBastionAvailable)
	return t, nil
}

func (s *ToolsStack) Variables() map[string]interface{} {
	return s.Stack.Variables()
}

func (s *ToolsStack) VerifyBastionAvailable() error {

	ssh := s.Cluster().Environment().Tarmak().SSH()

	if err := ssh.WriteConfig(); err != nil {
		return err
	}

	//Ensure go routine exits before returning
	finished := make(chan struct{})
	var wg sync.WaitGroup
	defer wg.Wait()

	//Capture signals to cancel ssh retry
	ctx, cancelRetries := context.WithCancel(context.Background())
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	wg.Add(1)
	go func() {
		select {
		case <-ch:
			cancelRetries()
		case <-finished:
		}
		signal.Stop(ch)
		wg.Done()
		return
	}()

	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = time.Second
	expBackoff.MaxElapsedTime = time.Minute * 1
	b := backoff.WithContext(expBackoff, ctx)

	executeSSH := func() error {
		retCode, err := ssh.Execute(
			"bastion",
			"/bin/true",
			[]string{},
		)

		msg := "error while connecting to bastion host"
		if err != nil {
			err = fmt.Errorf("%s: %v", msg, err)
			s.log.Warnf(err.Error())
			return err
		}
		if retCode != 0 {
			err = fmt.Errorf("%s unexpected return code: %d", msg, retCode)
			s.log.Warn(err.Error())
			return err
		}

		return nil
	}

	err := backoff.Retry(executeSSH, b)
	close(finished)
	if err != nil {
		return fmt.Errorf("failed to connect to bastion host: %v", err)
	}

	return nil
}
