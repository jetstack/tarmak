// Copyright Jetstack Ltd. See LICENSE for details.

package terraform

import (
	"fmt"

	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

type terraformOutputValue struct {
	Sensitive bool        `json:"sensitive,omitifempty"`
	Type      string      `json:"type,omitifempty"`
	Value     interface{} `value:"type,omitifempty"`
}

func (t *Terraform) Output(c interfaces.Cluster) (map[string]interface{}, error) {

	return nil, fmt.Errorf("unimplemented")
	/*
		stdOut, stdErr, returnCode, err := tc.Capture("terraform", []string{"output", "-json"})
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
	*/
}
