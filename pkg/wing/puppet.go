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
	"sync"
	"syscall"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/docker/docker/pkg/archive"
	"golang.org/x/net/context"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jetstack/tarmak/pkg/apis/wing/v1alpha1"
	"github.com/jetstack/tarmak/pkg/wing/provider/file"
	"github.com/jetstack/tarmak/pkg/wing/provider/s3"
)

// This make sure puppet is converged when neccessary
func (w *Wing) runPuppet(job *v1alpha1.WingJob) error {
	targetAPI := w.clientset.WingV1alpha1().PuppetTargets(w.flags.ClusterName)
	target, err := targetAPI.Get(
		job.Spec.PuppetTargetRef,
		metav1.GetOptions{},
	)

	noop := true
	if job.Spec.Operation == v1alpha1.ApplyOperation {
		noop = false
	}

	originalReader, err := w.getManifests(target)
	if err != nil {
		return err
	}

	// buffer file locally
	buf, err := ioutil.ReadAll(originalReader)
	if err != nil {
		return err
	}
	err = originalReader.Close()
	if err != nil {
		return err
	}
	// create reader from buffer
	reader := bytes.NewReader(buf)

	// build hash over puppet.tar.gz
	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return err
	}
	//hashString := fmt.Sprintf("sha256:%x", hash.Sum(nil))

	// roll back reader
	reader.Seek(0, 0)

	// read tar in
	tarReader, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}

	dir, err := ioutil.TempDir("", "wing-puppet-tar-gz")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir) // clean up

	err = archive.Unpack(tarReader, dir, &archive.TarOptions{})
	if err != nil {
		return err
	}
	tarReader.Close()

	puppetApplyCmd := func() error {
		output, retCode, err := w.puppetApply(dir, noop)

		w.log.Infof("puppet output: %s", output)

		if err == nil && retCode != 0 {
			err = fmt.Errorf("puppet apply has not converged yet (return code %d)", retCode)
		}

		if err != nil {
			output = fmt.Sprintf("puppet apply error: %s\n%s", err, output)
		}

		// start converging mainfest
		job.Status = &v1alpha1.WingJobStatus{
			Messages:            output,
			ExitCode:            retCode,
			LastUpdateTimestamp: metav1.Now(),
			Completed:           true,
			//Hash:      hashString,
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

	return nil
}

func (w *Wing) convergeMachine() error {
	machineAPI := w.clientset.WingV1alpha1().Machines(w.flags.ClusterName)
	machine, err := machineAPI.Get(
		w.flags.InstanceName,
		metav1.GetOptions{},
	)
	if err != nil {
		if kerr, ok := err.(*apierrors.StatusError); ok && kerr.ErrStatus.Reason == metav1.StatusReasonNotFound {
			machine = &v1alpha1.Machine{
				ObjectMeta: metav1.ObjectMeta{
					Name: w.flags.InstanceName,
				},
				Status: v1alpha1.MachineStatus{
					Converged: false,
				},
			}
			_, err := machineAPI.Create(machine)
			if err != nil {
				return fmt.Errorf("error creating machine: %s", err)
			}
			return nil
		}
		return fmt.Errorf("error get existing machine: %s", err)
	}

	if machine.Status.Converged {
		w.log.Infof("Machine already converged: ", machine.Name)
		return nil
	}

	puppetTarget := machine.Spec.PuppetTargetRef
	if puppetTarget == "" {
		w.log.Warn("no puppet target for machine: ", machine.Name)
		return nil
	}

	// FIXME: this shouldn't be done on the wing agent
	jobName := fmt.Sprintf("%s-%s", w.flags.InstanceName, puppetTarget)
	jobsAPI := w.clientset.WingV1alpha1().WingJobs(w.flags.ClusterName)
	job, err := jobsAPI.Get(
		jobName,
		metav1.GetOptions{},
	)
	if err != nil {
		if kerr, ok := err.(*apierrors.StatusError); ok && kerr.ErrStatus.Reason == metav1.StatusReasonNotFound {
			job = &v1alpha1.WingJob{
				ObjectMeta: metav1.ObjectMeta{
					Name: jobName,
				},
				Spec: &v1alpha1.WingJobSpec{
					InstanceName:     machine.Name,
					PuppetTargetRef:  puppetTarget,
					Operation:        "apply",
					RequestTimestamp: metav1.Now(),
				},
				Status: &v1alpha1.WingJobStatus{},
			}
			_, err := jobsAPI.Create(job)
			if err != nil {
				return fmt.Errorf("error creating WingJob: %s", err)
			}
			return nil
		}
		return fmt.Errorf("error get existing WingJob: %s", err)
	}

	machineCopy := machine.DeepCopy()
	machineCopy.Status.Converged = true
	_, err = machineAPI.Update(machineCopy)
	if err != nil {
		return err
	}

	return nil
}

func (w *Wing) converge(job *v1alpha1.WingJob) {
	w.convergeWG.Add(1)
	defer w.convergeWG.Done()

	jobCopy := job.DeepCopy()

	// run puppet
	err := w.runPuppet(jobCopy)
	if err != nil {
		w.log.Warn("running puppet failed: ", err)
	}

	// feedback puppet status to apiserver
	err = w.reportStatus(jobCopy)
	if err != nil {
		w.log.Warn("reporting status failed: ", err)
	}
}

func (w *Wing) puppetCommand(dir string, noop bool) Command {
	if w.puppetCommandOverride != nil {
		return w.puppetCommandOverride
	}

	args := []string{"apply"}

	if noop {
		args = append(args, "--noop")
	}

	args = append(args, []string{
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
	}...,
	)

	return &execCommand{
		Cmd: exec.Command("git", "status"),
	}

	return &execCommand{
		Cmd: exec.Command("puppet", args...),
	}
}

// apply puppet code in a specific directory
func (w *Wing) puppetApply(dir string, noop bool) (output string, retCode int, err error) {
	w.log.Infof("Running puppet in %s", dir)

	puppetCmd := w.puppetCommand(dir, noop)

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

func (w *Wing) reportStatus(job *v1alpha1.WingJob) error {
	jobAPI := w.clientset.WingV1alpha1().WingJobs(w.flags.ClusterName)

	_, err := jobAPI.Update(job)

	if err != nil {
		return fmt.Errorf("error updating job status: %s", err)
	}

	return nil
}

func (w *Wing) getManifests(target *v1alpha1.PuppetTarget) (io.ReadCloser, error) {
	if target.Source.S3 != nil {

		manifestStr := fmt.Sprintf("s3://%s/%s", target.Source.S3.BucketName, target.Source.S3.Path)
		w.log.Infof("Getting manifests from S3: %s", manifestStr)
		return s3.New(w.log).GetManifest(
			manifestStr,
			target.Source.S3.Region,
		)
	}

	if target.Source.File != nil {
		return file.New(w.log).GetManifest(target.Source.File.Path)
	}

	return nil, fmt.Errorf("unknown source type")
}
