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
			log:    logrus.NewEntry(logger),
			stopCh: make(chan struct{}),
		},
	}

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

func TestWing_SIGTERM_handler(t *testing.T) {
	w := newFakeWing(t)
	defer w.ctrl.Finish()
	defer deleteTmpFiles(t)

	go func() {
		time.Sleep(time.Second)
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	w.signalHandler()
	executeSleep(w, t)
}

func TestWing_SIGHUP_handler(t *testing.T) {
	w := newFakeWing(t)
	defer w.ctrl.Finish()
	defer deleteTmpFiles(t)

	go func() {
		time.Sleep(time.Second)
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	w.signalHandler()
	executeSleep(w, t)
}

func TestWing_SIGTERM_puppet(t *testing.T) {
	w := newFakeWing(t)
	defer w.ctrl.Finish()
	defer deleteTmpFiles(t)

	go func(w *fakeWing, t *testing.T) {
		time.Sleep(time.Second)
		go executeSleep(w, t)
		time.Sleep(time.Second)
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}(w, t)

	w.signalHandler()

	_, err := w.runPuppet()
	if err != nil {
		t.Errorf("%v", err)
	}
}

func TestWing_SIGHUP_puppet(t *testing.T) {
	w := newFakeWing(t)
	defer w.ctrl.Finish()
	defer deleteTmpFiles(t)

	go func(w *fakeWing, t *testing.T) {
		time.Sleep(time.Second)
		go executeSleep(w, t)
		time.Sleep(time.Second)
		syscall.Kill(syscall.Getpid(), syscall.SIGHUP)
		time.Sleep(time.Second)
		close(w.stopCh)
	}(w, t)

	w.signalHandler()

	_, err := w.runPuppet()
	if err != nil {
		t.Errorf("%v", err)
	}
}

func executeSleep(w *fakeWing, t *testing.T) {
	w.puppetCmd = exec.Command(
		"sleep",
		"10",
	)
	if err := w.puppetCmd.Start(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	err := w.puppetCmd.Wait()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				if int(status) != 15 {
					w.log.Errorf("expected error code 15, got %d", int(status))
				}
			}
		} else {
			t.Errorf("unexpected error: %v", err)
		}
	} else {
		t.Errorf("expected error, got none")
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
