// Copyright Jetstack Ltd. See LICENSE for details.
package e2e_test

import (
	"fmt"
	"os"
	"runtime"
	"testing"
)

func TestAWSSingleCluster(t *testing.T) {
	t.Parallel()
	skipE2ETests(t)

	ti := NewTarmakInstance(t)
	ti.singleCluster = true
	ti.singleZone = true

	if err := ti.GenerateAndBuild(); err != nil {
		t.Fatal(err)
	}

	defer func() {
		t.Log("run environment destroy command")
		c := ti.Command("environment", "destroy", ti.environmentName, "--auto-approve")
		// write error out to my stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			t.Errorf("unexpected error: %+v", err)
		}
	}()

	if err := ti.RunAndVerify(); err != nil {
		t.Fatal(err)
	}
}

func TestAWSMultiCluster(t *testing.T) {
	t.Parallel()
	skipE2ETests(t)

	ti := NewTarmakInstance(t)
	ti.singleCluster = false
	ti.singleZone = false

	if err := ti.GenerateAndBuild(); err != nil {
		t.Fatal(err)
	}

	defer func() {
		t.Log("run environment destroy command")
		c := ti.Command("environment", "destroy", ti.environmentName, "--auto-approve")
		// write error out to my stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			t.Errorf("unexpected error: %+v", err)
		}
	}()
	t.Log("run hub apply command")
	c := ti.Command("--current-cluster", fmt.Sprintf("%s-hub", ti.environmentName), "cluster", "apply")
	// write error out to my stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}

	if err := ti.RunAndVerify(); err != nil {
		t.Fatal(err)
	}
}

func TestAWSUpgradeTarmak(t *testing.T) {
	t.Parallel()
	skipE2ETests(t)

	ti := NewTarmakInstance(t)
	ti.singleCluster = true
	ti.singleZone = true

	ti.binPath = fmt.Sprintf("../../../tarmak_0.5.2_%s_%s", runtime.GOOS, runtime.GOARCH)

	if err := ti.GenerateAndBuild(); err != nil {
		t.Fatal(err)
	}

	defer func() {
		t.Log("run environment destroy command")
		c := ti.Command("environment", "destroy", ti.environmentName, "--auto-approve")
		// write error out to my stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			t.Errorf("unexpected error: %+v", err)
		}
	}()

	if err := ti.RunAndVerify(); err != nil {
		t.Fatal(err)
	}

	ti.binPath = fmt.Sprintf("../../../tarmak_%s_%s", runtime.GOOS, runtime.GOARCH)

	if err := ti.RunAndVerify(); err != nil {
		t.Fatal(err)
	}
}

func TestAWSUpgradeKubernetes(t *testing.T) {
	t.Parallel()
	skipE2ETests(t)

	ti := NewTarmakInstance(t)
	ti.singleCluster = true
	ti.singleZone = true

	if err := ti.GenerateAndBuild(); err != nil {
		t.Fatal(err)
	}

	defer func() {
		t.Log("run environment destroy command")
		c := ti.Command("environment", "destroy", ti.environmentName, "--auto-approve")
		// write error out to my stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			t.Errorf("unexpected error: %+v", err)
		}
	}()

	if err := ti.RunAndVerify(); err != nil {
		t.Fatal(err)
	}

	ti.UpdateKubernetesVersion()

	if err := ti.RunAndVerify(); err != nil {
		t.Fatal(err)
	}
}
