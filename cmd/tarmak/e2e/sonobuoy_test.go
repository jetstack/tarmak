// Copyright Jetstack Ltd. See LICENSE for details.
package e2e_test

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/pkg/archive"
)

var mode = "Quick"

type SonobuoyInstance struct {
	t            *testing.T
	binPath      string // bin path to tarmak binary
	mode         string // quick or conformance
	retrievePath string
	kubeconfig   string
}

func NewSonobuoyInstance(t *testing.T, environment, kubeconfig string) *SonobuoyInstance {
	si := &SonobuoyInstance{
		t:       t,
		binPath: "../../../bin/sonobuoy",
		mode:    mode,
		retrievePath: fmt.Sprintf("%s/%s-test-%s",
			os.TempDir(), environment, randStringRunes(6)),
		kubeconfig: kubeconfig,
	}

	if _, err := os.Stat(si.binPath); os.IsNotExist(err) {
		t.Fatal("sonobuoy binary not exissing: ", si.binPath)
	} else if err != nil {
		t.Fatal("error finding sonobuoy binary: ", si.binPath)
	}

	return si
}

func (s *SonobuoyInstance) run(completeStr string, args ...string) (bool, error) {
	args = append(args, "--kubeconfig", s.kubeconfig)

	cmd := exec.Command(s.binPath, args...)

	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		return false, err
	}

	if completeStr != "" {
		if !strings.Contains(string(out), completeStr) {
			return true, nil
		}
	}

	return true, nil
}

func (s *SonobuoyInstance) RunAndVerify() error {
	s.t.Log("running sonobuoy")
	if _, err := s.run("", "run", "--mode", s.mode); err != nil {
		return fmt.Errorf("error sonobuoy run: %s", err)
	}

	defer func() {
		s.t.Log("deleting sonobuoy resources in the cluster")
		if _, err := s.run("", "delete"); err != nil {
			s.t.Errorf("unexpected error deleting: %s", err)
		}
	}()

	s.t.Log("waiting for sonobuoy to complete")

	time.Sleep(time.Minute * 2)

	for {
		complete, err := s.run("Sonobuoy has completed.", "status")
		if err != nil {
			return fmt.Errorf("error sonobuoy status: %s", err)
		}

		if complete {
			break
		}

		time.Sleep(10 * time.Second)
	}

	s.t.Log("completed sonobuoy tests")

	s.t.Log("retrieving results")

	if _, err := s.run("", "retrieve", s.retrievePath); err != nil {
		return fmt.Errorf("error sonobuoy run: %s", err)
	}

	files, err := ioutil.ReadDir(s.retrievePath)
	if err != nil {
		return fmt.Errorf("failed to read retrieve path: %s", err)
	}

	if len(files) != 1 {
		return fmt.Errorf("expecting single file in %s, got=%d",
			s.retrievePath, len(files))
	}

	path := filepath.Join(s.retrievePath, files[0].Name())

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %s", path, err)
	}
	defer f.Close()

	tarReader, err := gzip.NewReader(f)
	if err != nil {
		return err
	}

	err = archive.Unpack(tarReader, s.retrievePath, &archive.TarOptions{
		NoLchown: true,
	})
	if err != nil {
		return err
	}
	tarReader.Close()

	s.t.Logf("retrieved logs written to: %s", s.retrievePath)

	r := regexp.MustCompile(`\d+ Passed \| \d+ Failed \| \d+ Pending \| \d+ Skipped`)
	path = filepath.Join(s.retrievePath, "plugins/e2e/results/e2e.log")

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	res := r.FindString(string(b))
	if res == "" {
		return fmt.Errorf("failed to get results from %s", path)
	}

	fmt.Printf("e2e: %s\n", res)

	fails := strings.Split(res, "|")[1]
	n, err := strconv.Atoi(strings.Fields(fails)[0])
	if err != nil {
		return fmt.Errorf("failed to parse int in result string: %s",
			strings.Fields(fails)[0])
	}

	if n != 0 {
		s.t.Errorf("got %d failures", n)
		return fmt.Errorf("%s", res)
	}

	s.t.Log("e2e tests PASSED")

	return nil
}
