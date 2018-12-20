// Copyright Jetstack Ltd. See LICENSE for details.
package e2e_test

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/google/goexpect"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")
var e2eEnable bool

const tarmakInitPrompt = "> "
const tarmakInitYesNo = " \\[Y\\/n\\] "

func init() {
	// add custom e2e flag
	flag.BoolVar(&e2eEnable, "e2e", false, "Enable E2E tests")
	flag.Parse()

	// seed random
	rand.Seed(time.Now().UnixNano())
}

func skipE2ETests(t *testing.T) {
	goTestCmd := "go test -v -timeout 1h github.com/jetstack/tarmak/cmd/tarmak/e2e -e2e"

	if !e2eEnable {
		t.Skipf("E2E tests are disabled. Run tests with '%s'", goTestCmd)
	}
	if timeoutFlag := flag.Lookup("test.timeout"); timeoutFlag != nil {
		t.Logf("flag=%+v", timeoutFlag.Value.String())
		if timeout, err := time.ParseDuration(timeoutFlag.Value.String()); err != nil {
			t.Fatal("Unparseable timeout: ", err)
		} else {
			if timeout < time.Hour {
				t.Skipf("E2E tests are disabled, as timeout is set to short. Run tests with '%s'", goTestCmd)
			}
		}
	}
}

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

type TarmakInstance struct {
	t                     *testing.T
	configPath            string // config path
	binPath               string // bin path to tarmak binary
	region                string // region of the cluster
	singleCluster         bool   // single or multi cluster
	singleZone            bool   // single or multi zone
	environmentName       string // name of the tarmak environment
	awsBucketDynamoPrefix string // aws bucket and dynamodb prefix
	awsSTSVaultPath       string // vault auth path
}

func NewTarmakInstance(t *testing.T) *TarmakInstance {
	ti := &TarmakInstance{
		t:                     t,
		binPath:               fmt.Sprintf("../../../tarmak_%s_%s", runtime.GOOS, runtime.GOARCH),
		region:                "eu-central-1",
		singleCluster:         true,
		singleZone:            true,
		environmentName:       fmt.Sprintf("e2e%s", randStringRunes(6)),
		awsBucketDynamoPrefix: fmt.Sprintf("jetstack-e2e-%s", randStringRunes(6)),
		awsSTSVaultPath:       "jetstack/aws/jetstack-dev/sts/admin",
	}

	if _, err := os.Stat(ti.binPath); os.IsNotExist(err) {
		t.Fatal("tarmak binary not existing: ", ti.binPath)
	} else if err != nil {
		t.Fatal("error finding tarmak binary: ", ti.binPath)
	}

	if dir, err := ioutil.TempDir("", "tarmak-config"); err != nil {
		t.Fatal("error creating temporary config directory: ", err)
	} else {
		t.Logf("created temp config directory in %s", dir)
		ti.configPath = dir
	}

	return ti
}

func (ti *TarmakInstance) Command(args ...string) *exec.Cmd {
	c := exec.CommandContext(
		context.Background(),
		ti.binPath,
		args...,
	)

	c.Env = append(os.Environ(), fmt.Sprintf("TARMAK_CONFIG=%s", ti.configPath))

	return c
}

func (ti *TarmakInstance) Init() error {
	e, wait, err := ti.Expect("init")
	if err != nil {
		return fmt.Errorf("error init expect: %s", err)
	}
	defer e.Close()

	if err := ti.initProvider(e); err != nil {
		return fmt.Errorf("error initialsing provider config: %s", err)
	}
	if err := ti.initEnvironment(e); err != nil {
		return fmt.Errorf("error initialsing environment config: %s", err)
	}
	if err := ti.initCluster(e); err != nil {
		return fmt.Errorf("error initialsing cluster config: %s", err)
	}

	if err := wait(); err != nil {
		return fmt.Errorf("error waiting for tarmak: %s", err)
	}

	return nil

}

// This returns a GExpect. It needs to be closed if it's no longer used
func (ti *TarmakInstance) Expect(args ...string) (*expect.GExpect, func() error, error) {
	c := ti.Command(args...)

	// write error out to my stdout
	c.Stderr = os.Stderr

	stdIn, err := c.StdinPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("error creating pipe: %s", err)
	}

	stdOut, err := c.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("error creating pipe: %s", err)
	}

	if err := c.Start(); err != nil {
		return nil, nil, fmt.Errorf("Unexcepted error starting tarmak: %+v", err)
	}

	waitCh := make(chan error, 1)

	e, _, err := expect.SpawnGeneric(
		&expect.GenOptions{
			In:  stdIn,
			Out: stdOut,
			Wait: func() error {
				err := c.Wait()
				waitCh <- err
				return err
			},
			Close: c.Process.Kill,
			Check: func() bool {
				if c.Process == nil {
					return false
				}
				// Sending Signal 0 to a process returns nil if process can take a signal , something else if not.
				return c.Process.Signal(syscall.Signal(0)) == nil
			},
		},
		time.Second,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating expect: %s", err)
	}

	wait := func() error {
		err := <-waitCh
		return err
	}

	return e, wait, nil
}

func (ti *TarmakInstance) initProvider(e *expect.GExpect) error {
	ti.t.Log("setting up provider")
	res, err := e.ExpectBatch([]expect.Batcher{
		&expect.BExp{R: tarmakInitPrompt},
		&expect.BSnd{S: "jetstack-dev\n"},                              // enter provider name
		&expect.BExp{R: tarmakInitPrompt},                              //
		&expect.BSnd{S: "1\n"},                                         // select AWS
		&expect.BExp{R: tarmakInitPrompt},                              //
		&expect.BSnd{S: "2\n"},                                         // vault auth
		&expect.BExp{R: tarmakInitPrompt},                              //
		&expect.BSnd{S: fmt.Sprintf("%s\n", ti.awsSTSVaultPath)},       // path for vault endpoint
		&expect.BExp{R: tarmakInitPrompt},                              //
		&expect.BSnd{S: fmt.Sprintf("%s\n", ti.awsBucketDynamoPrefix)}, // bucket / dynamodb backend
		&expect.BExp{R: tarmakInitPrompt},                              //
		&expect.BSnd{S: "develop.tarmak.org\n"},                        // public route 53
		&expect.BExp{R: tarmakInitYesNo},                               //
		&expect.BSnd{S: 	"Y\n"},                                         // save provider
	}, 30*time.Second)
	if err != nil {
		return fmt.Errorf("unexpected expect flow for init provider res=%+v error: %+v", res, err)
	}
	return nil
}

func (ti *TarmakInstance) initEnvironment(e *expect.GExpect) error {
	ti.t.Logf("setting up environment %s", ti.environmentName)
	res, err := e.ExpectBatch([]expect.Batcher{
		&expect.BExp{R: tarmakInitPrompt},                        //
		&expect.BSnd{S: fmt.Sprintf("%s\n", ti.environmentName)}, // environment
		&expect.BExp{R: tarmakInitPrompt},                        //
		&expect.BSnd{S: fmt.Sprintf("%s\n", ti.environmentName)}, // enter project name
		&expect.BExp{R: tarmakInitPrompt},                        //
		&expect.BSnd{S: "tech+e2e@jetstack.io\n"},                // enter contact mail
		&expect.BExp{R: tarmakInitPrompt},                        //
		&expect.BSnd{S: "7\n"},                                   // select eu-central-1 // TODO: this will break with AWS adding more regions
		&expect.BExp{R: tarmakInitYesNo},                         //
		&expect.BSnd{S: "Y\n"},                                   // save environment
	}, 30*time.Second)
	if err != nil {
		return fmt.Errorf("unexpected expect flow for init environment res=%+v error: %+v", res, err)
	}
	return nil
}

func (ti *TarmakInstance) initCluster(e *expect.GExpect) error {

	batches := []expect.Batcher{
		&expect.BExp{R: tarmakInitPrompt}, //
	}

	// single vs. multi cluster
	if ti.singleCluster {
		batches = append(batches, &expect.BSnd{S: "1\n"})
	} else {
		batches = append(batches, &expect.BSnd{S: "2\n"})
	}

	// single vs. multi zone
	if !ti.singleZone {
		batches = append(
			batches,
			&expect.BExp{R: tarmakInitPrompt},
			&expect.BSnd{S: "2\n"}, // enable second zone
			&expect.BExp{R: tarmakInitPrompt},
			&expect.BSnd{S: "3\n"}, // enable third zone
		)
	}

	// continue zone selection
	batches = append(
		batches,
		&expect.BExp{R: tarmakInitPrompt},
		&expect.BSnd{S: "4\n"}, // continue
	)

	// set cluster name for multi cluster
	if !ti.singleCluster {
		batches = append(
			batches,
			&expect.BExp{R: tarmakInitPrompt},
			&expect.BSnd{S: "greeen\n"},
		)
	}

	// save cluster
	batches = append(
		batches,
		&expect.BExp{R: tarmakInitYesNo}, //
		&expect.BSnd{S: "Y\n"},           // save cluster
	)

	ti.t.Logf("setting up cluster in environment %s", ti.environmentName)
	res, err := e.ExpectBatch(
		batches,
		30*time.Second,
	)
	if err != nil {
		return fmt.Errorf("unexpected expect flow for init cluster res=%+v error: %+v", res, err)
	}
	return nil
}

func (ti *TarmakInstance) UpdateKubernetesVersion() error {

	config, err := ioutil.ReadFile(fmt.Sprintf("%v/tarmak.yaml", ti.configPath))
	if err != nil {
		fmt.Errorf("Error reading config file: %+v", err)
	}
	output := strings.Replace(string(config), "version: 1.11.5", "version: 1.12.4", 1)

	d1 := []byte(output)
	err = ioutil.WriteFile(fmt.Sprintf("%v/tarmak.yaml", ti.configPath), d1, 0644)
	if err != nil {
		fmt.Errorf("Error writing config file: %+v", err)
	}
	return nil
}

func (ti *TarmakInstance) RunAndVerify() error {
	ti.t.Log("run cluster apply command")
	c := ti.Command("cluster", "apply")
	// write error out to my stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("unexpected error: %+v", err)
	}

	ti.t.Log("get component status")
	c = ti.Command("cluster", "kubectl", "get", "cs", "-o", "yaml")
	// write error out to my stdout
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	if err := c.Run(); err != nil {
		return fmt.Errorf("unexpected error: %+v", err)
	}
	return nil
}