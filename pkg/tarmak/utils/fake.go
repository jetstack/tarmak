// Copyright Jetstack Ltd. See LICENSE for details.
package utils

import (
	"fmt"
)

type Closer struct {
	stopCh chan struct{}
}

func NewCloser() (*Closer, chan struct{}) {
	c := &Closer{
		stopCh: make(chan struct{}),
	}
	return c, c.stopCh

}

func (c *Closer) Close() error {
	if c.stopCh == nil {
		return fmt.Errorf("no close channel available, already closed")
	}
	close(c.stopCh)
	c.stopCh = nil
	return nil
}
