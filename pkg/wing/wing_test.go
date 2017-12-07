// Copyright Jetstack Ltd. See LICENSE for details.
package wing

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/archive"
	gomock "github.com/golang/mock/gomock"
	"github.com/hashicorp/go-multierror"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"

	"github.com/jetstack/tarmak/pkg/wing/client"
	"github.com/jetstack/tarmak/pkg/wing/mocks"
)

var manifestURL, manifestURLgz string

type fakeWing struct {
	*Wing
	ctrl *gomock.Controller

	fakeRest       *mocks.MockInterface
	fakeHTTPClient *mocks.MockHTTPClient
	fakeCommand    *mocks.MockCommand

	signalCh chan os.Signal
}

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func newFakeWing(t *testing.T) *fakeWing {
	// create tmp gz file for wing testing
	err := createTmpFiles()
	if err != nil {
		t.Fatal(err)
	}

	logger := logrus.New()
	if testing.Verbose() {
		logger.Level = logrus.DebugLevel
	} else {
		logger.Out = ioutil.Discard
	}

	w := &fakeWing{
		ctrl: gomock.NewController(t),
		Wing: &Wing{
			clientset: &client.Clientset{},
			flags: &Flags{
				ManifestURL:  manifestURLgz,
				ServerURL:    "fakeServerURL",
				ClusterName:  "fakeClusterName",
				InstanceName: "fakeInstanceName",
			},
			log:            logrus.NewEntry(logger),
			stopCh:         make(chan struct{}),
			convergeStopCh: make(chan struct{}),
		},
	}

	w.signalCh = make(chan os.Signal, 1)

	w.fakeCommand = mocks.NewMockCommand(w.ctrl)
	w.Wing.puppetCommandOverride = w.fakeCommand
	w.fakeCommand.EXPECT().StderrPipe().AnyTimes().Return(nopCloser{bytes.NewBufferString("i am stderr")}, nil)
	w.fakeCommand.EXPECT().StdoutPipe().AnyTimes().Return(nopCloser{bytes.NewBufferString("i am stdout")}, nil)

	w.fakeRest = mocks.NewMockInterface(w.ctrl)
	w.fakeHTTPClient = mocks.NewMockHTTPClient(w.ctrl)
	w.clientset = client.New(w.fakeRest)

	w.fakeHTTPClient.EXPECT().Do(gomock.Any()).AnyTimes().Return(&http.Response{StatusCode: 0, Body: nopCloser{bytes.NewBufferString("")}}, nil)
	contentConfig := rest.ContentConfig{
		GroupVersion: &schema.GroupVersion{
			Version: "v1",
		},
	}

	request := rest.NewRequest(w.fakeHTTPClient, "verb", nil, "versionedAPIPath", contentConfig, rest.Serializers{}, nil, nil)
	w.fakeRest.EXPECT().Get().AnyTimes().Return(request)

	return w
}

// this tests when the SIGTERM hits wing when it's currently running puppet
func TestWing_SIGTERM_handler_first_execute(t *testing.T) {
	w := newFakeWing(t)
	defer w.ctrl.Finish()
	defer deleteTmpFiles(t)

	// enable signal handling
	go func() {
		w.signalHandler(w.signalCh)
	}()

	// start replacement process
	process := exec.Command(
		"sleep",
		"100",
	)
	if err := process.Start(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// channel that gets closed after sleep has been stopped
	exitCh := make(chan struct{})
	go func() {
		process.Wait()
		close(exitCh)
	}()

	// once started send sigterm to wing
	w.fakeCommand.EXPECT().Start().Do(func() {
		// after start send sigterm to wing
		w.signalCh <- syscall.SIGTERM
	})

	// make wing wait for us to signal the exit
	w.fakeCommand.EXPECT().Wait().Do(func() {
		<-exitCh
	})

	// return sleep process instead
	w.fakeCommand.EXPECT().Process().AnyTimes().Return(process.Process)

	// run a converge
	w.converge()

	if _, ok := (<-w.convergeStopCh); ok {
		t.Error("expected convergeStopCh to be closed")
	}
	if _, ok := (<-w.stopCh); ok {
		t.Error("expected stopCh to be closed")
	}
}

// this tests when the SIGTERM hits wing when it's currently waiting in the exp backoff
func TestWing_SIGTERM_handler_backoff(t *testing.T) {
	w := newFakeWing(t)
	defer w.ctrl.Finish()
	defer deleteTmpFiles(t)

	// TODO: find a better way to get an error with non zero return code
	process := exec.Command(
		"false",
	)
	if err := process.Start(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	processErr := process.Wait()

	// enable signal handling
	go func() {
		w.signalHandler(w.signalCh)
	}()

	w.fakeCommand.EXPECT().Start().Do(func() {
		w.log.Debugf("fake process called start")
	})

	// make wing wait for us to signal the exit
	puppetFinished := make(chan struct{})
	w.fakeCommand.EXPECT().Wait().Return(processErr).Do(func() {
		w.log.Debugf("fake process called wait")
		close(puppetFinished)
	})

	// send signal after puppet exited
	go func() {
		<-puppetFinished
		time.Sleep(time.Millisecond)
		w.signalCh <- syscall.SIGTERM
	}()

	// run a converge
	w.converge()
	if _, ok := (<-w.convergeStopCh); ok {
		t.Error("expected convergeStopCh to be closed")
	}
	if _, ok := (<-w.stopCh); ok {
		t.Error("expected stopCh to be closed")
	}
}

// this tests when the SIGHUP hits wing when it's currently running puppet
func TestWing_SIGHUP_handler_first_excute(t *testing.T) {
	w := newFakeWing(t)
	defer w.ctrl.Finish()
	defer deleteTmpFiles(t)

	// enable signal handling
	go func() {
		w.signalHandler(w.signalCh)
	}()

	// start replacement process
	process := exec.Command(
		"sleep",
		"100",
	)
	if err := process.Start(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// channel that gets closed after sleep has been stopped
	exitCh := make(chan struct{})
	go func() {
		process.Wait()
		close(exitCh)

		//After SIGHUP expect another convergence to run and then close
		w.fakeCommand.EXPECT().Start().Do(func() {
			w.log.Debugf("fake process called start again")
		}).Times(1)
		w.fakeCommand.EXPECT().Wait().Do(func() {
			w.log.Debugf("fake process called wait again")
			close(w.stopCh)
		}).Times(1)
	}()

	// once started send sigterm to wing
	w.fakeCommand.EXPECT().Start().Do(func() {
		// after start send SIGHUP to wing
		w.signalCh <- syscall.SIGHUP
		<-w.convergeStopCh
	}).Times(1)

	// make wing wait for us to signal the exit
	w.fakeCommand.EXPECT().Wait().Do(func() {
		<-exitCh
	}).Times(1)

	// return sleep process instead
	w.fakeCommand.EXPECT().Process().AnyTimes().Return(process.Process)

	// run a converge
	w.converge()

	if _, ok := (<-w.convergeStopCh); ok {
		t.Error("expected convergeStopCh to be closed")
	}
	if _, ok := (<-w.stopCh); ok {
		t.Error("expected stopCh to be closed")
	}
}

// this tests when the SIGHUP hits wing when it's currently waiting in the exp backoff
func TestWing_SIGHUP_handler_backoff(t *testing.T) {
	w := newFakeWing(t)
	defer w.ctrl.Finish()
	defer deleteTmpFiles(t)

	// TODO: find a better way to get an error with non zero return code
	process := exec.Command(
		"false",
	)
	if err := process.Start(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	processErr := process.Wait()

	// enable signal handling
	go func() {
		w.signalHandler(w.signalCh)
	}()

	w.fakeCommand.EXPECT().Start().Do(func() {
		w.log.Debugf("fake process called start")
	})

	// make wing wait for us to signal the exit
	puppetFinished := make(chan struct{})
	w.fakeCommand.EXPECT().Wait().Return(processErr).Do(func() {
		w.log.Debugf("fake process called wait")
		close(puppetFinished)
	})

	// send signal after puppet exited
	go func() {
		<-puppetFinished

		//Ensure re-converge after sending SIGHUP during backoff
		w.fakeCommand.EXPECT().Start().Do(func() {
			w.log.Debugf("fake process called start again")
		})
		w.fakeCommand.EXPECT().Wait().Return(processErr).Do(func() {
			w.log.Debugf("fake process called wait again")
			close(w.stopCh)
		})

		time.Sleep(time.Millisecond)
		w.signalCh <- syscall.SIGHUP
	}()

	// run a converge
	w.converge()
	if _, ok := (<-w.convergeStopCh); ok {
		t.Error("expected convergeStopCh to be closed")
	}
	if _, ok := (<-w.stopCh); ok {
		t.Error("expected stopCh to be closed")
	}
}

//Test nothing happens when sending a SIGTERM after puppet has converged
func TestWing_SIGTERM_puppet_converged(t *testing.T) {
	w := newFakeWing(t)
	defer w.ctrl.Finish()
	defer deleteTmpFiles(t)

	// TODO: find a better way to get an error with non zero return code
	process := exec.Command(
		"true",
	)
	if err := process.Start(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	processErr := process.Wait()

	// enable signal handling
	go func() {
		w.signalHandler(w.signalCh)
	}()

	w.fakeCommand.EXPECT().Start().Do(func() {
		w.log.Debugf("fake process called start")
	})

	// make wing wait for us to signal the exit
	puppetFinished := make(chan struct{})
	w.fakeCommand.EXPECT().Wait().Return(processErr).Do(func() {
		w.log.Debugf("fake process called wait")
		close(puppetFinished)
	})

	// run a converge
	w.converge()

	w.convergeWG.Wait()
	if _, ok := (<-puppetFinished); ok {
		t.Error("expected puppetFinished to be closed")
	}
	time.Sleep(time.Microsecond)

	//Send SIGTERM after convergence
	w.signalCh <- syscall.SIGTERM

	//Ensure close from signal handler
	if _, ok := (<-w.convergeStopCh); ok {
		t.Error("expected convergeStopCh to be closed")
	}
	if _, ok := (<-w.stopCh); ok {
		t.Error("expected stopCh to be closed")
	}
}

//Test re-convergence happens when sending a SIGHUP after puppet has converged
func TestWing_SIGHUP_puppet_converged(t *testing.T) {
	w := newFakeWing(t)
	defer w.ctrl.Finish()
	defer deleteTmpFiles(t)

	// TODO: find a better way to get an error with non zero return code
	process := exec.Command(
		"true",
	)
	if err := process.Start(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	processErr := process.Wait()

	// enable signal handling
	go func() {
		w.signalHandler(w.signalCh)
	}()

	w.fakeCommand.EXPECT().Start().Do(func() {
		w.log.Debugf("fake process called start")
	})

	// make wing wait for us to signal the exit
	puppetFinished := make(chan struct{})
	w.fakeCommand.EXPECT().Wait().Return(processErr).Do(func() {
		w.log.Debugf("fake process called wait")
		close(puppetFinished)
	})

	// run a converge
	w.converge()

	w.convergeWG.Wait()
	if _, ok := (<-puppetFinished); ok {
		t.Error("expected puppetFinished to be closed")
	}

	//Expect re-convergence
	w.fakeCommand.EXPECT().Start().Do(func() {
		w.log.Debugf("fake process called start again")
	})
	w.fakeCommand.EXPECT().Wait().Return(processErr).Do(func() {
		w.log.Debugf("fake process called wait again")
		close(w.stopCh)
	})

	time.Sleep(time.Microsecond)

	//Send SIGHUP after convergence
	w.signalCh <- syscall.SIGHUP

	//Ensure close from wait
	if _, ok := (<-w.stopCh); ok {
		t.Error("expected stopCh to be closed")
	}
}

func createTmpFiles() error {
	file, err := ioutil.TempFile(os.TempDir(), "manifestURL")
	if err != nil {
		return err
	}
	filegz, err := ioutil.TempFile(os.TempDir(), "manifestURL.gz.tar")
	if err != nil {
		return err
	}
	manifestURL = file.Name()
	manifestURLgz = filegz.Name()

	if err := ioutil.WriteFile(manifestURL, []byte("0000"), 0644); err != nil {
		return err
	}

	tarOpts := &archive.TarOptions{
		Compression: archive.Gzip,
		NoLchown:    true,
	}

	reader, err := archive.TarWithOptions(manifestURL, tarOpts)
	if err != nil {
		return fmt.Errorf("error creating tar from path '%s': %v", manifestURL, err)
	}

	if _, err := io.Copy(filegz, reader); err != nil {
		return fmt.Errorf("error writing temp tar file: %v", err)
	}

	return nil
}

func deleteTmpFiles(t *testing.T) {
	var result *multierror.Error

	if err := os.Remove(manifestURL); err != nil {
		result = multierror.Append(result, err)
	}

	if err := os.Remove(manifestURLgz); err != nil {
		result = multierror.Append(result, err)
	}

	if result != nil {
		t.Errorf("failed to delete temp files: %v", result)
	}
}
