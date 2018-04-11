// Copyright Jetstack Ltd. See LICENSE for details.

package terraform

import (
	"bytes"
	"fmt"

	"encoding/json"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type terraformOutputValue struct {
	Sensitive bool        `json:"sensitive,omitifempty"`
	Type      string      `json:"type,omitifempty"`
	Value     interface{} `value:"type,omitifempty"`
}

func (t *Terraform) Output(c interfaces.Cluster) (map[string]interface{}, error) {

	stdOutBuf := new(bytes.Buffer)
	stdErrBuf := new(bytes.Buffer)

	if err := t.command(
		c,
		[]string{
			"terraform",
			"output",
			"-json",
		},
		nil,
		stdOutBuf,
		stdErrBuf,
	); err != nil {
		t.log.Error(stdErrBuf.String())
		return nil, fmt.Errorf("error getting terraform output: %s")
	}

	var values map[string]terraformOutputValue
	if err := json.Unmarshal(stdOutBuf.Bytes(), &values); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %s", err)
	}

	variables := make(map[string]interface{})
	for key, value := range values {
		variables[key] = value.Value
	}
	return variables, nil

}
