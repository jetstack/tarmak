// Copyright Jetstack Ltd. See LICENSE for details.
package utils

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type Context struct {
	context.Context
	cancel func()
	sig    os.Signal
}

var notifies = []os.Signal{syscall.SIGINT, syscall.SIGTERM}

var _ interfaces.Context = &Context{}

func NewContext() interfaces.Context {
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, notifies...)

	c := &Context{
		ctx,
		cancel,
		syscall.SIGCONT,
	}

	go func() {
		select {
		case sig := <-sigCh:
			c.sig = sig
			log.Infof("caught signal '%v'", sig)
			cancel()
		case <-ctx.Done():
			log.Info("context done")
		}
		signal.Stop(sigCh)
		log.Info("cancelling")
		cancel()
	}()

	return c
}

func (c *Context) Signal() os.Signal {
	return c.sig
}

func (c *Context) Err() error {
	return c.Err()
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

func (c *Context) WaitOrCancel(f func() error, ignoredExitStatuses ...int) {
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
			case <-time.After(time.Second * 2):
				log.Warn("Tarmak is shutting down.")
				log.Warn("* Tarmak will attempt to kill the current task.")
				log.Warn("* Send another SIGTERM or SIGINT (ctrl-c) to exit immediately.")
			}
		}
	}()

	err := f()
	switch err {
	case context.Canceled:
		log.Warn("Tarmak was canceled. Re-run to complete any remaining tasks.")
	case nil:
		log.Info("Tarmak performed all tasks successfully.")
	default:
		exitError, ok := err.(*exec.ExitError)
		if ok {
			status := exitError.ProcessState.Sys().(syscall.WaitStatus)
			exitStatus := status.ExitStatus()

			errorOk := false
			for _, status := range ignoredExitStatuses {
				if exitStatus == status {
					errorOk = true
					break
				}
			}
			if !errorOk {
				log.Errorf("Tarmak exited with an error: %s", err)
			}
			log.Exit(exitStatus)
		}
		log.Fatalf("Tarmak exited with an error: %s", err)
	}
}

func BasicSignalHandler(l *log.Entry) chan struct{} {
	stopCh := make(chan struct{})
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	go func(l *log.Entry) {
		<-ch
		l.Infof("Received signal interupt. shutting down...")
		close(stopCh)
		<-ch
		l.Infof("Force closed.")
		os.Exit(1)
	}(l)

	return stopCh
}
