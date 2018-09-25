// Copyright Jetstack Ltd. See LICENSE for details.
package e2e_test

import (
	"fmt"
	"os"
	"testing"
)

func TestAWSSingleCluster(t *testing.T) {
	t.Parallel()
	skipE2ETests(t)

	ti := NewTarmakInstance(t)
	ti.singleCluster = true
	ti.singleZone = true

	t.Log("initialise config for single cluster")
	if err := ti.Init(); err != nil {
		t.Errorf("unexpected error: %+v", err)
	}

	t.Log("build tarmak image")
	c := ti.Command("cluster", "image", "build")
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		t.Errorf("unexpected error: %+v", err)
	}

	defer func() {
		t.Log("run cluster destroy command")
		c = ti.Command("cluster", "destroy")
		// write error out to my stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			t.Errorf("unexpected error: %+v", err)
		}
	}()
	t.Log("run cluster apply command")
	c = ti.Command("cluster", "apply")
	// write error out to my stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	t.Log("get component status")
	c = ti.Command("cluster", "kubectl", "get", "cs", "-o", "yaml")
	// write error out to my stdout
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	if err := c.Run(); err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}
}

func TestAWSMultiCluster(t *testing.T) {
	t.Parallel()
	skipE2ETests(t)

	ti := NewTarmakInstance(t)
	ti.singleCluster = false
	ti.singleZone = false

	t.Log("initialise config for single cluster")
	if err := ti.Init(); err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	t.Log("build tarmak image")
	c := ti.Command("cluster", "image", "build")
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	defer func() {
		t.Log("run hub destroy command")
		c = ti.Command("--current-cluster", fmt.Sprintf("%s-hub", ti.environmentName), "cluster", "destroy")
		// write error out to my stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			t.Errorf("unexpected error: %+v", err)
		}
	}()
	t.Log("run hub apply command")
	c = ti.Command("--current-cluster", fmt.Sprintf("%s-hub", ti.environmentName), "cluster", "apply")
	// write error out to my stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	defer func() {
		t.Log("run cluster destroy command")
		c = ti.Command("cluster", "destroy")
		// write error out to my stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			t.Errorf("unexpected error: %+v", err)
		}
	}()
	t.Log("run cluster apply command")
	c = ti.Command("cluster", "apply")
	// write error out to my stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	t.Log("get component status")
	c = ti.Command("cluster", "kubectl", "get", "cs", "-o", "yaml")
	// write error out to my stdout
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	if err := c.Run(); err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

}
