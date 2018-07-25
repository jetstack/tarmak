// Copyright Jetstack Ltd. See LICENSE for details.
package terraform

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/hashicorp/terraform/command"
	"github.com/kardianos/osext"
	"github.com/sirupsen/logrus"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
	"github.com/jetstack/tarmak/pkg/terraform/providers/tarmak/rpc"
)

const debugShell = "debug-shell"

// wingHash is set by a linker flag to the hash of the lastest wing binary
var wingHash = ""

type Terraform struct {
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

// this method perpares the terraform plugins folder. This folder contains
// terraform providers and provisioners in general. We are pointing through
// symlinks to the tarmak binary, which contains all relevant providers
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

func (t *Terraform) Prepare(cluster interfaces.Cluster) error {

	// generate tf code
	if err := t.GenerateCode(cluster); err != nil {
		return fmt.Errorf("failed to generate code: %s", err)
	}

	// symlink tarmak plugins into folder
	if err := t.preparePlugins(cluster); err != nil {
		return fmt.Errorf("failed to prepare plugins: %s", err)
	}

	// run init
	if err := t.command(
		cluster,
		[]string{
			"terraform",
			"init",
			"-get-plugins=false",
			"-input=false",
		},
		nil,
		nil,
		nil,
	); err != nil {
		return fmt.Errorf("failed to run terraform init: %s", err)
	}

	return nil
}

func (t *Terraform) terraformWrapper(cluster interfaces.Cluster, command string, args []string) error {

	if err := t.Prepare(cluster); err != nil {
		return fmt.Errorf("failed to prepare terraform: %s", err)
	}

	// listen to rpc
	stopCh := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := rpc.ListenUnixSocket(
			rpc.New(t.tarmak.Cluster()),
			t.socketPath(cluster),
			stopCh,
		); err != nil {
			t.log.Fatalf("error listening to unix socket: %s", err)
		}
	}()

	// command
	if command == debugShell {
		dir := t.codePath(cluster)
		envVars, err := t.envVars(cluster)
		if err != nil {
			return err
		}

		// use $SHELL if available, fall back to /bin/sh
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
			envVars = append(envVars, fmt.Sprintf("PS1=[%s]$ ", dir))
		}

		cmd := exec.Command(shell)
		cmd.Dir = dir
		// envVars variables will override any shell envs with equal key
		cmd.Env = append(os.Environ(), envVars...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		cmd.Run()
	} else if command != "" {
		cmdArgs := []string{
			"terraform",
			command,
		}
		cmdArgs = append(cmdArgs, args...)

		if err := t.command(
			cluster,
			cmdArgs,
			nil,
			nil,
			nil,
		); err != nil {
			return err
		}
	}

	close(stopCh)
	wg.Wait()

	return nil
}

func (t *Terraform) envVars(cluster interfaces.Cluster) ([]string, error) {
	envVars := []string{
		"TF_IN_AUTOMATION=1",
	}

	// get environment variables necessary for provider
	if environmentProvider, err := cluster.Environment().Provider().Environment(); err != nil {
		return []string{}, fmt.Errorf("error getting environment secrets from provider: %s", err)
	} else {
		envVars = append(envVars, environmentProvider...)
	}

	envVars = append(envVars, fmt.Sprintf("TF_LOG=%s", os.Getenv("TF_LOG")))

	return envVars, nil
}

func (t *Terraform) command(cluster interfaces.Cluster, args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	envVars, err := t.envVars(cluster)
	if err != nil {
		return err
	}

	binaryPath, err := osext.Executable()
	if err != nil {
		return fmt.Errorf("error finding tarmak executable: %s", err)
	}

	cmd := exec.Command(
		binaryPath,
		args...,
	)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	// forward stdout
	if stdout == nil {
		stdoutScanner := bufio.NewScanner(stdoutPipe)
		go func() {
			for stdoutScanner.Scan() {
				t.log.WithField("std", "out").Debug(stdoutScanner.Text())
			}
		}()
	} else {
		cmd.Stdout = stdout
	}

	// forward stderr
	if stderr == nil {
		stderrScanner := bufio.NewScanner(stderrPipe)
		go func() {
			for stderrScanner.Scan() {
				t.log.WithField("std", "err").Debug(stderrScanner.Text())
			}
		}()
	} else {
		cmd.Stderr = stderr
	}

	cmd.Stdin = stdin
	cmd.Dir = t.codePath(cluster)
	cmd.Env = envVars

	if err := cmd.Run(); err != nil {
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
		[]string{"-input=false", "-auto-approve=true"},
	)
}

func (t *Terraform) Destroy(cluster interfaces.Cluster) error {
	return t.terraformWrapper(
		cluster,
		"destroy",
		[]string{"-force", "-refresh=false"},
	)
}

func (t *Terraform) Shell(cluster interfaces.Cluster) error {

	if err := t.terraformWrapper(cluster, debugShell, nil); err != nil {
		return err
	}
	return nil
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
