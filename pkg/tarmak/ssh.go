package tarmak

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jetstack/tarmak/pkg/tarmak/utils"
)

func (t *Tarmak) SSH(argsAdditional []string) {

	hosts, err := t.Context().Environment().Provider().ListHosts()
	if err != nil {
		t.log.Fatal(err)
	}

	var sshConfig bytes.Buffer
	sshConfig.WriteString(fmt.Sprintf("# ssh config for tarmak context %s\n", t.Context().Name()))

	for _, host := range hosts {
		_, err = sshConfig.WriteString(host.SSHConfig())
		if err != nil {
			t.log.Fatal(err)
		}
	}

	err = utils.EnsureDirectory(filepath.Dir(t.Context().SSHConfigPath()), 0700)
	if err != nil {
		t.log.Fatal(err)
	}

	err = ioutil.WriteFile(t.Context().SSHConfigPath(), sshConfig.Bytes(), 0600)
	if err != nil {
		t.log.Fatal(err)
	}

	args := []string{
		"ssh",
		"-F",
		t.Context().SSHConfigPath(),
	}
	args = append(args, argsAdditional...)

	cmd := exec.Command(args[0], args[1:len(args)]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	err = cmd.Start()
	if err != nil {
		t.log.Fatal(err)
	}

	err = cmd.Wait()
	if err != nil {
		t.log.Fatal(err)
	}

}
