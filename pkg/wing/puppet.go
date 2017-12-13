// Copyright Jetstack Ltd. See LICENSE for details.
package wing

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/cenkalti/backoff"
	"github.com/docker/docker/pkg/archive"
	"github.com/jetstack/tarmak/pkg/apis/wing/v1alpha1"
	"github.com/jetstack/tarmak/pkg/wing/provider/file"
	"github.com/jetstack/tarmak/pkg/wing/provider/s3"
	"golang.org/x/net/context"
)

// This make sure puppet is converged when neccessary
func (w *Wing) runPuppet() (*v1alpha1.InstanceStatus, error) {
	// start converging mainfest
	status := &v1alpha1.InstanceStatus{
		Converge: &v1alpha1.InstanceStatusManifest{
			State: v1alpha1.InstanceManifestStateConverging,
		},
	}

	err := w.reportStatus(status)
	if err != nil {
		w.log.Warn("reporting status failed: ", err)
	}

	originalReader, err := w.getManifests(w.flags.ManifestURL)
	if err != nil {
		return status, err
	}

	// buffer file locally
	buf, err := ioutil.ReadAll(originalReader)
	if err != nil {
		return status, err
	}
	err = originalReader.Close()
	if err != nil {
		return status, err
	}
	// create reader from buffer
	reader := bytes.NewReader(buf)

	// build hash over puppet.tar.gz
	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return status, err
	}
	hashString := fmt.Sprintf("sha256:%x", hash.Sum(nil))

	// roll back reader
	reader.Seek(0, 0)

	// read tar in
	tarReader, err := gzip.NewReader(reader)
	if err != nil {
		return status, err
	}

	dir, err := ioutil.TempDir("", "wing-puppet-tar-gz")
	if err != nil {
		return status, err
	}
	defer os.RemoveAll(dir) // clean up

	err = archive.Unpack(tarReader, dir, &archive.TarOptions{})
	if err != nil {
		return status, err
	}
	tarReader.Close()

	var puppetMessages []string
	var puppetRetCodes []int

	puppetApplyCmd := func() error {
		output, retCode, err := w.puppetApply(dir)

		if err == nil && retCode != 0 {
			err = fmt.Errorf("puppet apply has not converged yet (return code %d)", retCode)
		}

		if err != nil {
			output = fmt.Sprintf("puppet apply error: %s\n%s", err, output)
		}

		puppetMessages = append(puppetMessages, output)
		puppetRetCodes = append(puppetRetCodes, retCode)

		// start converging mainfest
		status = &v1alpha1.InstanceStatus{
			Converge: &v1alpha1.InstanceStatusManifest{
				State:     v1alpha1.InstanceManifestStateConverging,
				Messages:  puppetMessages,
				ExitCodes: puppetRetCodes,
				Hash:      hashString,
			},
		}
		statusErr := w.reportStatus(status)
		if statusErr != nil {
			w.log.Warn("reporting status failed: ", statusErr)
		}

		return err
	}

	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = time.Second * 30
	expBackoff.MaxElapsedTime = time.Minute * 30

	// add context to backoff
	ctx, cancelRetries := context.WithCancel(context.Background())
	b := backoff.WithContext(expBackoff, ctx)

	quitCh := make(chan struct{})
	defer close(quitCh)

	// cancel retries when supposed to stop

	go func() {
		for {
			select {
			case <-w.convergeStopCh:
				cancelRetries()
				return
			case <-quitCh:
				return
			}
		}
	}()

	err = backoff.Retry(puppetApplyCmd, b)
	if err != nil {
		w.log.Error("error applying puppet:", err)
	}

	return status, nil
}

func (w *Wing) converge() {
	w.convergeWG.Add(1)
	defer w.convergeWG.Done()

	// run puppet
	status, err := w.runPuppet()
	if err != nil {
		status.Converge.State = v1alpha1.InstanceManifestStateError
		status.Converge.Messages = append(status.Converge.Messages, err.Error())
		w.log.Error(err)
	} else {
		status.Converge.State = v1alpha1.InstanceManifestStateConverged
	}

	// feedback puppet status to apiserver
	err = w.reportStatus(status)
	if err != nil {
		w.log.Warn("reporting status failed: ", err)
	}
}

func (w *Wing) puppetCommand(dir string) Command {
	if w.puppetCommandOverride != nil {
		return w.puppetCommandOverride
	}

	return &execCommand{
		Cmd: exec.Command(
			"puppet",
			"apply",
			"--detailed-exitcodes",
			"--color",
			"no",
			"--environment",
			"production",
			"--hiera_config",
			filepath.Join(dir, "hiera.yaml"),
			"--modulepath",
			filepath.Join(dir, "modules"),
			filepath.Join(dir, "manifests/site.pp"),
		),
	}
}

// apply puppet code in a specific directory
func (w *Wing) puppetApply(dir string) (output string, retCode int, err error) {
	puppetCmd := w.puppetCommand(dir)

	var mu sync.Mutex
	var wg sync.WaitGroup

	outputBuffer := new(bytes.Buffer)
	stdoutPipe, err := puppetCmd.StdoutPipe()
	if err != nil {
		return "", 0, err
	}

	stderrPipe, err := puppetCmd.StderrPipe()
	if err != nil {
		return "", 0, err
	}

	// forward stdout
	stdoutScanner := bufio.NewScanner(stdoutPipe)
	go func() {
		for stdoutScanner.Scan() {
			//critical region to avoid race condition
			mu.Lock()
			w.log.WithField("cmd", "puppet").Debug(stdoutScanner.Text())
			outputBuffer.WriteString(stdoutScanner.Text())
			outputBuffer.WriteString("\n")
			mu.Unlock()
		}
	}()

	// forward stderr
	stderrScanner := bufio.NewScanner(stderrPipe)
	go func() {
		for stderrScanner.Scan() {
			//critical region to avoid race condition
			mu.Lock()
			w.log.WithField("cmd", "puppet").Debug(stderrScanner.Text())
			outputBuffer.WriteString(stderrScanner.Text())
			outputBuffer.WriteString("\n")
			mu.Unlock()
		}
	}()

	// handle exit signal
	wg.Add(1)
	quitCh := make(chan struct{})
	go func() {
		for {
			select {
			case <-w.convergeStopCh:
				if puppetCmd != nil && puppetCmd.Process() != nil {
					w.log.Debugf("terminating puppet pid=%d process early", puppetCmd.Process().Pid)
					err := puppetCmd.Process().Signal(syscall.SIGTERM)
					if err != nil {
						w.log.Warn("error terminating puppet process early:", err)
					}
				}
				wg.Done()
				return
			case <-quitCh:
				wg.Done()
				return
			}
		}
	}()

	err = puppetCmd.Start()
	if err != nil {
		return "", 0, err
	}

	w.log.Printf("Waiting for command to finish...")
	err = puppetCmd.Wait()
	close(quitCh)
	//ensure go routine has closed
	wg.Wait()

	//critical region to avoid race condition
	mu.Lock()
	output = outputBuffer.String()
	mu.Unlock()
	if err != nil {
		perr, ok := err.(*exec.ExitError)
		if ok {
			if status, ok := perr.Sys().(syscall.WaitStatus); ok {
				return output, status.ExitStatus(), nil
			}
		}
		return output, 0, err
	}

	return output, 0, nil
}

// report status to the API server
func (w *Wing) reportStatus(status *v1alpha1.InstanceStatus) error {
	instanceAPI := w.clientset.WingV1alpha1().Instances(w.flags.ClusterName)
	instance, err := instanceAPI.Get(
		w.flags.InstanceName,
		metav1.GetOptions{},
	)
	if err != nil {
		if kerr, ok := err.(*apierrors.StatusError); ok && kerr.ErrStatus.Reason == metav1.StatusReasonNotFound {
			instance = &v1alpha1.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name: w.flags.InstanceName,
				},
				Status: status.DeepCopy(),
			}
			_, err := instanceAPI.Create(instance)
			if err != nil {
				return fmt.Errorf("error creating instance: %s", err)
			}
			return nil
		}
		return fmt.Errorf("error get existing instance: %s", err)
	}

	instance.Status = status.DeepCopy()
	_, err = instanceAPI.Update(instance)
	if err != nil {
		return fmt.Errorf("error updating existing instance: %s", err)
		// TODO: handle race for update
	}

	return nil

}

func (w *Wing) getManifests(manifestURL string) (io.ReadCloser, error) {
	if strings.HasPrefix(manifestURL, "s3://") {
		return s3.New(w.log).GetManifest(manifestURL)
	}
	return file.New(w.log).GetManifest(manifestURL)
}
