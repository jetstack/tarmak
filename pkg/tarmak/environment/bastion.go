// Copyright Jetstack Ltd. See LICENSE for details.
package environment

import (
	"bufio"
	"fmt"
	"io"
	"time"

	"github.com/cenkalti/backoff"
)

func (e *Environment) VerifyBastionAvailable() error {

	ssh := e.Tarmak().SSH()

	if err := ssh.WriteConfig(e.Hub()); err != nil {
		return err
	}

	done := make(chan struct{})
	defer close(done)
	ctx := e.tarmak.CancellationContext().TryOrCancel(done)

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
			[]string{"/bin/true"},
			nil, nil, stderrW,
		)

		msg := "error while connecting to bastion host"
		if err != nil {
			err = fmt.Errorf("%s: %v", msg, err)
			return err
		}
		if retCode != 0 {
			err = fmt.Errorf("%s unexpected return code: %d", msg, retCode)
			return err
		}

		return nil
	}

	err := backoff.Retry(executeSSH, b)
	if err != nil {
		return fmt.Errorf("failed to connect to bastion host: %v", err)
	}

	return nil
}
