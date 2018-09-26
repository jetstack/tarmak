// Copyright Jetstack Ltd. See LICENSE for details.
package utils

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type CancellationContext struct {
	context.Context
	cancel func()
	sig    os.Signal
}

var notifies = []os.Signal{syscall.SIGINT, syscall.SIGTERM}

var _ interfaces.CancellationContext = &CancellationContext{}

func NewCancellationContext() interfaces.CancellationContext {
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, notifies...)

	c := &CancellationContext{
		ctx,
		cancel,
		syscall.SIGCONT,
	}

	go func() {
		select {
		case sig := <-sigCh:
			c.sig = sig
			log.Infof("Caught signal %s.", sig)
			log.Info("Attempting to shutdown gracefully.")
			cancel()
		case <-ctx.Done():
		}

		signal.Stop(sigCh)
		return
	}()

	return c
}

func (c *CancellationContext) Signal() os.Signal {
	return c.sig
}

func (c *CancellationContext) Err() error {
	if c.Context.Err() == context.Canceled {
		return fmt.Errorf("signal %s", c.Signal())
	}
	return c.Context.Err()
}

func MakeShutdownCh() <-chan struct{} {
	resultCh := make(chan struct{})

	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, notifies...)
	go func() {
		for {
			<-signalCh
			resultCh <- struct{}{}
		}
	}()

	return resultCh
}

func (c *CancellationContext) WaitOrCancel(f func() error) {
	c.WaitOrCancelReturnCode(
		func() (int, error) {
			return 0, f()
		},
	)
}

func (c *CancellationContext) WaitOrCancelReturnCode(f func() (int, error)) {
	defer c.cancel()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	defer wg.Wait()

	finished := make(chan struct{})
	defer close(finished)

	go func() {
		select {
		case <-c.Done():
		case <-finished:
		}
		for {
			select {
			case <-finished:
				wg.Done()
				return
			case <-time.After(time.Second * 3):
				log.Warn("Tarmak is shutting down.")
				log.Warn("* Tarmak will attempt to kill the current task.")
				log.Warn("* Send another SIGTERM or SIGINT (ctrl-c) to exit immediately.")
			}
		}
	}()

	retCode, err := f()
	switch err {
	case nil:
		log.Info("Tarmak performed all tasks successfully.")
		log.Exit(retCode)
	case context.Canceled:
		log.Errorf("Tarmak was canceled (%s). Re-run to complete any remaining tasks.", c.sig)
		log.Exit(1)
	default:
		log.Errorf("Tarmak exited with an error: %s", err)
		log.Exit(1)
	}
}

func BasicSignalHandler(l *log.Entry) chan struct{} {
	stopCh := make(chan struct{})
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	go func(l *log.Entry) {
		sig := <-ch
		l.Infof("Received signal %s. Shutting down...", sig)
		close(stopCh)
		sig = <-ch
		l.Infof("Received signal %s. Force closing.", sig)
		os.Exit(1)
	}(l)

	return stopCh
}
