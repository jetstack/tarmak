// Copyright Jetstack Ltd. See LICENSE for details.
package environment

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/cenkalti/backoff"
)

func (e *Environment) VerifyBastionAvailable() error {

	ssh := e.Tarmak().SSH()

	if err := ssh.WriteConfig(e.Hub()); err != nil {
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
	expBackoff.MaxElapsedTime = time.Minute * 2
	b := backoff.WithContext(expBackoff, ctx)

	stderrR, stderrW := io.Pipe()
	stderrScanner := bufio.NewScanner(stderrR)
	go func() {
		for stderrScanner.Scan() {
			e.log.WithField("std", "err").Debug(stderrScanner.Text())
		}
	}()

	executeSSH := func() error {
		retCode, err := ssh.Execute(
			"bastion",
			"/bin/true",
			nil, nil, stderrW,
		)

		msg := "error while connecting to bastion host"
		if err != nil {
			err = fmt.Errorf("%s: %v", msg, err)
			e.log.Warnf(err.Error())
			return err
		}
		if retCode != 0 {
			err = fmt.Errorf("%s unexpected return code: %d", msg, retCode)
			e.log.Warn(err.Error())
			return err
		}

		return nil
	}

	err := backoff.Retry(executeSSH, b)
	close(finished)
	if err != nil {
		return fmt.Errorf("failed to connect to bastion host: %v", err)
	}

	e.log.Info("connection to bastion host successful")

	return nil
}
