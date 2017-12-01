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
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/cenk/backoff"
	"github.com/docker/docker/pkg/archive"
	"github.com/jetstack/tarmak/pkg/apis/wing/v1alpha1"
	"github.com/jetstack/tarmak/pkg/wing/provider/file"
	"github.com/jetstack/tarmak/pkg/wing/provider/s3"
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

	go func(puppetApplyCmd backoff.Operation) {
		b := backoff.NewExponentialBackOff()
		b.InitialInterval = time.Second * 30
		b.MaxElapsedTime = time.Minute * 30

		err := backoff.Retry(puppetApplyCmd, b)
		if err != nil {
			log.Fatal(err)
			return
		}

		// successfull backoff
		close(w.convergedCh)

	}(puppetApplyCmd)

	// block on successful backoff or close
	select {
	case <-w.stopCh:
		break
	case <-w.convergedCh:
		break
	}

	return status, nil
}

func (w *Wing) converge() {
	w.convergedCh = make(chan struct{})

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

// apply puppet code in a specific directory
func (w *Wing) puppetApply(dir string) (output string, retCode int, err error) {
	w.puppetCmd = exec.Command(
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
	)

	outputBuffer := new(bytes.Buffer)
	stdoutPipe, err := w.puppetCmd.StdoutPipe()
	if err != nil {
		return "", 0, err
	}

	stderrPipe, err := w.puppetCmd.StderrPipe()
	if err != nil {
		return "", 0, err
	}

	stdoutScanner := bufio.NewScanner(stdoutPipe)
	go func() {
		for stdoutScanner.Scan() {
			w.log.WithField("cmd", "puppet").Debug(stdoutScanner.Text())
			outputBuffer.WriteString(stdoutScanner.Text())
			outputBuffer.WriteString("\n")
		}
	}()

	stderrScanner := bufio.NewScanner(stderrPipe)
	go func() {
		for stderrScanner.Scan() {
			w.log.WithField("cmd", "puppet").Debug(stderrScanner.Text())
			outputBuffer.WriteString(stderrScanner.Text())
			outputBuffer.WriteString("\n")
		}
	}()

	err = w.puppetCmd.Start()
	if err != nil {
		return "", 0, err
	}

	w.log.Printf("Waiting for command to finish...")
	err = w.puppetCmd.Wait()
	output = outputBuffer.String()
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
