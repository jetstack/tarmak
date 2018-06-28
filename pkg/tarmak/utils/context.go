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

func GetContext() (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case sig := <-ch:
			log.Infof("caught signal '%v'", sig)
		case <-ctx.Done():
			log.Info("context done")
		}
		signal.Stop(ch)
		log.Info("cancelling")
		cancel()
	}()
	return ctx, cancel
}

func WaitOrCancel(f func(context.Context) error, stopCh chan struct{}, ignoredExitStatuses ...int) {
	ctx, cancel := GetContext()
	defer cancel()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	defer wg.Wait()
	finished := make(chan struct{})
	defer close(finished)

	closedStopCh := false

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
			case <-time.After(time.Second):
				log.Warn("Tarmak is shutting down.")
				log.Warn("* Tarmak will exit after the current task finishes.")
				log.Warn("* Send another SIGTERM or SIGINT (ctrl-c) to exit immediately.")
				if !closedStopCh {
					close(stopCh)
					closedStopCh = true
				}
			}
		}
	}()
	err := f(ctx)
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
				log.WithError(err).Errorf("Tarmak exited with an error: %s", err)
			}
			log.Exit(exitStatus)
		}
		log.WithError(err).Fatalf("Tarmak exited with an error: %s", err)
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
