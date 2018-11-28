// Copyright Jetstack Ltd. See LICENSE for details.
package cmd

import (
	"sync"
	"testing"

	"github.com/spf13/cobra"
)

const (
	globalFlag = "--current-cluster=xxx"
)

type cmdTest struct {
	*testing.T
	cmd    *cobra.Command
	argsCh chan string
	args   []string
}

func newCmdTest(t *testing.T, cmd *cobra.Command) *cmdTest {
	return &cmdTest{
		T:   t,
		cmd: cmd,
	}
}

func (c *cmdTest) testRunFunc() func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		for _, a := range args {
			c.argsCh <- a
		}
		close(c.argsCh)
	}
}

func (c *cmdTest) flagIgnored(exp bool) {
	c.argsCh = make(chan string)
	c.cmd.Run = c.testRunFunc()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		found := false
		for a := range c.argsCh {
			if a == globalFlag {
				found = true
				if !exp {
					c.Errorf("got global flag as argument but expected not to: %s [%s]", c.args, globalFlag)
				}
			}
		}

		if !found && exp {
			c.Errorf("did not receive flag in arguments but was expecting to: %s [%s]", c.args, globalFlag)
		}
	}()

	Execute(c.args[1:])

	wg.Wait()
}

func Test_KubectlParsing(t *testing.T) {
	c := newCmdTest(t, clusterKubectlCmd)
	for _, a := range [][]string{
		{"tarmak", "kubectl", globalFlag, "arg", "arg"},
		{"tarmak", "kubectl", "arg", globalFlag, "arg"},
		{"tarmak", "cluster", "kubectl", globalFlag, "arg", "arg"},
		{"tarmak", "cluster", "kubectl", "arg", globalFlag, "arg"},
	} {
		c.args = a
		c.flagIgnored(true)
	}

	for _, a := range [][]string{
		{"tarmak", globalFlag, "kubectl", "arg", "arg"},
		{"tarmak", globalFlag, "cluster", "kubectl", "arg", "arg"},
		{"tarmak", "cluster", globalFlag, "kubectl", "arg", "arg"},
	} {
		c.args = a
		c.flagIgnored(false)
	}
}

func Test_SSHParsing(t *testing.T) {
	c := newCmdTest(t, clusterSshCmd)
	for _, a := range [][]string{
		{"tarmak", "cluster", "ssh", "arg", globalFlag, "arg"},
		{"tarmak", "cluster", "ssh", globalFlag, "arg", "arg"},
	} {
		c.args = a
		c.flagIgnored(true)
	}

	for _, a := range [][]string{
		{"tarmak", globalFlag, "cluster", "ssh", "arg", "arg"},
		{"tarmak", "cluster", globalFlag, "ssh", "arg", "arg"},
	} {
		c.args = a
		c.flagIgnored(false)
	}
}
