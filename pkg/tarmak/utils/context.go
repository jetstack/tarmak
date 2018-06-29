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
)

type Context struct {
	context.Context
	cancel func()
}

func GetContext() *Context {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case sig := <-ch:
			log.Infof("caught signal '%v'", sig)
			cancel()
		case <-ctx.Done():
			log.Info("context done")
		}
		signal.Stop(ch)
		log.Info("cancelling")
		cancel()
	}()

	return &Context{
		ctx,
		cancel,
	}
}

func WaitOrCancel(f func(context.Context) error, c context.Context, ignoredExitStatuses ...int) {
	ctx, ok := c.(*Context)
	if !ok {
		log.Error("Failed to assert context")
		return
	}
	defer ctx.cancel()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	defer wg.Wait()
	finished := make(chan struct{})
	defer close(finished)

	go func() {
		select {
		case <-ctx.Done():
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
	err := f(c)
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
