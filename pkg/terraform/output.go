// Copyright Jetstack Ltd. See LICENSE for details.

package terraform

import (
	"bytes"
	"fmt"
	"os"

	"encoding/json"

	"github.com/hashicorp/terraform/command"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type terraformOutputValue struct {
	Sensitive bool        `json:"sensitive,omitifempty"`
	Type      string      `json:"type,omitifempty"`
	Value     interface{} `value:"type,omitifempty"`
}

func (t *Terraform) Output(c interfaces.Cluster) (map[string]interface{}, error) {

	//return nil, fmt.Errorf("unimplemented")

	stdOut, stdErr, returnCode, err := terraformOutput([]string{"-json"})
	if err != nil {
		return nil, err
	}
	if exp, act := 0, returnCode; exp != act {
		return nil, fmt.Errorf("unexpected return code: exp=%d, act=%d: %s", exp, act, stdErr)
	}

	var values map[string]terraformOutputValue
	if err := json.Unmarshal([]byte(stdOut), &values); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %s", err)
	}

	variables := make(map[string]interface{})
	for key, value := range values {
		variables[key] = value.Value
	}
	return variables, nil

}

func terraformOutput(args []string) (stdOut string, stdErr string, returnCode int, err error) {

	var stdOutBuf bytes.Buffer
	var stdErrBuf bytes.Buffer
	//stdOutWriter := bufio.NewWriter(&stdOutBuf)
	//stdErrWriter := bufio.NewWriter(&stdErrBuf)
	//os.Stdout
	c := &command.OutputCommand{
		Meta: newMeta(newErrUI(os.Stdout, os.Stderr)),
		//Meta: newMeta(newErrUI(os.Stdout, os.Stdout)),
	}

	err = os.Stdout.Sync()
	if err != nil {
		return "", "", -1, fmt.Errorf("error flushing stdout: %s", err)
	}
	err = os.Stderr.Sync()
	if err != nil {
		return "", "", -1, fmt.Errorf("error flushing stdout: %s", err)
	}

	errCode := c.Run(args)
	//if errCode != 0 {
	return "", "", -1, fmt.Errorf("error running command: %d %s %s %s %s", errCode, "test", stdOutBuf.String(), stdErrBuf.String(), "test")
	//}

	/*command := []string{cmd}
	command = append(command, args...)
	ac.log.WithField("command", command).Debug()
	exec, err := ac.app.dockerClient.CreateExec(docker.CreateExecOptions{
		Cmd:          command,
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Container:    ac.dockerContainer.ID,
	})
	if err != nil {
		return "", "", -1, err
	}*/

	//var stdOutBuf bytes.Buffer
	//var stdErrBuf bytes.Buffer
	//stdOutWriter := bufio.NewWriter(&stdOutBuf)
	//stdErrWriter := bufio.NewWriter(&stdErrBuf)

	/*cmd.Stdout = &stdOutBuf
	cmd.Stderr = &stdErrBuf
	cmd.Args = args

	err = cmd.Run()
	if err != nil {
		return "", "", -1, fmt.Errorf("error running command: %s %s %s", err, stdOutBuf.String(), stdErrBuf.String())
	}*/

	/*err = ac.app.dockerClient.StartExec(exec.ID, docker.StartExecOptions{
		ErrorStream:  stdErrWriter,
		OutputStream: stdOutWriter,
	})
	if err != nil {
		return "", "", -1, fmt.Errorf("error starting exec: %s", err)
	}

	err = stdOutWriter.Flush()
	if err != nil {
		return "", "", -1, fmt.Errorf("error flushing stdout: %s", err)
	}
	err = stdErrWriter.Flush()
	if err != nil {
		return "", "", -1, fmt.Errorf("error flushing stdout: %s", err)
	}

	execInspect, err := ac.app.dockerClient.InspectExec(exec.ID)
	if err != nil {
		return "", "", -1, fmt.Errorf("error inspecting exec: %s", err)
	}*/

	return stdOutBuf.String(), stdErrBuf.String(), errCode, nil //cmd.ProcessState., nil
}
