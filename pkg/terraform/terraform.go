// Copyright Jetstack Ltd. See LICENSE for details.
package terraform

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/terraform/command"
	"github.com/kardianos/osext"
	"github.com/sirupsen/logrus"

	tarmakDocker "github.com/jetstack/tarmak/pkg/docker"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
	"github.com/jetstack/tarmak/pkg/terraform/providers/tarmak/rpc"
)

type Terraform struct {
	*tarmakDocker.App
	log    *logrus.Entry
	tarmak interfaces.Tarmak
}

func New(tarmak interfaces.Tarmak) *Terraform {
	log := tarmak.Log().WithField("module", "terraform")

	return &Terraform{
		log:    log,
		tarmak: tarmak,
	}
}

// this method perpares the terraform plugins folder. This folder contains terraform providers and provisioners in general. We are pointing through symlinks to the tarmak binary, which contains all relevant providers
func (t *Terraform) preparePlugins(c interfaces.Cluster) error {
	binaryPath, err := osext.Executable()
	if err != nil {
		return fmt.Errorf("error finding tarmak executable: %s", err)
	}

	pluginPath := t.pluginPath(c)
	if err := utils.EnsureDirectory(pluginPath, 0755); err != nil {
		return err
	}

	for providerName, _ := range InternalProviders {
		destPath := filepath.Join(pluginPath, fmt.Sprintf("terraform-provider-%s", providerName))
		if stat, err := os.Lstat(destPath); err != nil && !os.IsNotExist(err) {
			return err
		} else if err == nil {
			if (stat.Mode() & os.ModeSymlink) == 0 {
				return fmt.Errorf("%s is not a symbolic link", destPath)
			}

			if linkPath, err := os.Readlink(destPath); err != nil {
				return err
			} else if linkPath == binaryPath {
				// link points to correct destination
				continue
			}

			err := os.Remove(destPath)
			if err != nil {
				return err
			}
		}

		err := os.Symlink(
			binaryPath,
			destPath,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// plugin path that stores terraform providers binaries
func (t *Terraform) pluginPath(c interfaces.Cluster) string {
	return filepath.Join(t.codePath(c), command.DefaultPluginVendorDir)
}

// code path to store terraform modules and files
func (t *Terraform) codePath(c interfaces.Cluster) string {
	return filepath.Join(c.ConfigPath(), "terraform")
}

// socket path for the tarmak provider socket
func tarmakSocketPath(clusterConfig string) string {
	return filepath.Join(clusterConfig, "tarmak.sock")
}
func (t *Terraform) socketPath(c interfaces.Cluster) string {
	return tarmakSocketPath(c.ConfigPath())
}

func (t *Terraform) terraformWrapper(cluster interfaces.Cluster, command string, args []string) error {

	// generate tf code
	if err := t.GenerateCode(cluster); err != nil {
		return err
	}
	terraformCodePath := t.codePath(cluster)

	// symlink tarmak plugins into folder
	if err := t.preparePlugins(cluster); err != nil {
		return err
	}

	binaryPath, err := osext.Executable()
	if err != nil {
		return fmt.Errorf("error finding tarmak executable: %s", err)
	}

	// listen to rpc
	if err := rpc.ListenUnixSocket(
		t.log,
		rpc.New(t.tarmak.Cluster()),
		t.socketPath(cluster),
	); err != nil {
		return err
	}

	envVars := []string{
		"TF_IN_AUTOMATION=1",
	}

	// get environment variables necessary for provider
	if environmentProvider, err := cluster.Environment().Provider().Environment(); err != nil {
		return fmt.Errorf("error getting environment secrets from provider: %s", err)
	} else {
		envVars = append(envVars, environmentProvider...)
	}

	// run init
	cmdInit := exec.Command(
		binaryPath,
		"terraform",
		"init",
		"-get-plugins=false",
		"-input=false",
	)
	cmdInit.Dir = terraformCodePath
	cmdInit.Stdout = os.Stdout
	cmdInit.Stderr = os.Stderr
	cmdInit.Stdin = os.Stdin
	cmdInit.Env = envVars
	if err := cmdInit.Run(); err != nil {
		return err
	}

	// plan
	cmdArgs := []string{
		"terraform",
		command,
	}
	cmdArgs = append(cmdArgs, args...)

	cmdPlan := exec.Command(
		binaryPath,
		cmdArgs...,
	)
	cmdPlan.Stdout = os.Stdout
	cmdPlan.Stderr = os.Stderr
	cmdPlan.Stdin = os.Stdin
	cmdPlan.Dir = terraformCodePath
	cmdPlan.Env = envVars

	if err := cmdPlan.Run(); err != nil {
		return err
	}

	return nil
}

func (t *Terraform) Plan(cluster interfaces.Cluster) error {
	return t.terraformWrapper(
		cluster,
		"plan",
		[]string{"-detailed-exitcode", "-input=false"},
	)
}

func (t *Terraform) Apply(cluster interfaces.Cluster) error {
	return t.terraformWrapper(
		cluster,
		"apply",
		[]string{"-input=false"},
	)
}

func (t *Terraform) Destroy(cluster interfaces.Cluster) error {
	return t.terraformWrapper(
		cluster,
		"destroy",
		[]string{"-force"},
	)
}

func (t *Terraform) Shell(cluster interfaces.Cluster) error {
	// TODO: needs to be implemented
	return fmt.Errorf("Shell unimplemented")
}

// convert interface map to terraform.tfvars format
func MapToTerraformTfvars(input map[string]interface{}) (output string, err error) {
	var buf bytes.Buffer

	for key, value := range input {
		switch v := value.(type) {
		case map[string]string:
			_, err := buf.WriteString(fmt.Sprintf("%s = {\n", key))
			if err != nil {
				return "", err
			}

			keys := make([]string, len(v))
			pos := 0
			for key, _ := range v {
				keys[pos] = key
				pos++
			}
			sort.Strings(keys)
			for _, key := range keys {
				_, err := buf.WriteString(fmt.Sprintf("  %s = \"%s\"\n", key, v[key]))
				if err != nil {
					return "", err
				}
			}

			_, err = buf.WriteString("}\n")
			if err != nil {
				return "", err
			}
		case []string:
			values := make([]string, len(v))
			for pos, _ := range v {
				values[pos] = fmt.Sprintf(`"%s"`, v[pos])
			}
			_, err := buf.WriteString(fmt.Sprintf("%s = [%s]\n", key, strings.Join(values, ", ")))
			if err != nil {
				return "", err
			}
		case string:
			_, err := buf.WriteString(fmt.Sprintf("%s = \"%s\"\n", key, v))
			if err != nil {
				return "", err
			}
		case int:
			_, err := buf.WriteString(fmt.Sprintf("%s = %d\n", key, v))
			if err != nil {
				return "", err
			}
		case *net.IPNet:
			_, err := buf.WriteString(fmt.Sprintf("%s = \"%s\"\n", key, v.String()))
			if err != nil {
				return "", err
			}
		default:
			return "", fmt.Errorf("ignoring unknown var key='%s' type='%#+v'", key, v)
		}
	}
	return buf.String(), nil
}
