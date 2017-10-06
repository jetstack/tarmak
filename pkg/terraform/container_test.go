// Copyright Jetstack Ltd. See LICENSE for details.
package terraform

import (
	"reflect"
	"testing"
)

func Test_MapToTerraformCLIArgs(t *testing.T) {
	var err error
	var args string

	args, err = MapToTerraformTfvars(map[string]interface{}{"test": "value"})
	if err != nil {
		t.Error("unexpected error: ", err)
	}
	if exp, act := "test = \"value\"\n", args; exp != act {
		t.Errorf("unexpected output exp: '%s', act: '%s'", exp, act)
	}

	args, err = MapToTerraformTfvars(map[string]interface{}{"test": map[string]string{"key1": "value1"}})
	if err != nil {
		t.Error("unexpected error: ", err)
	}
	if exp, act := "test = {\n  key1 = \"value1\"\n}\n", args; !reflect.DeepEqual(exp, act) {
		t.Errorf("unexpected output exp: '%s', act: '%s'", exp, act)
	}

	args, err = MapToTerraformTfvars(map[string]interface{}{"test": map[string]string{"key1": "value1", "key2": "value2"}})
	if err != nil {
		t.Error("unexpected error: ", err)
	}
	if exp, act := "test = {\n  key1 = \"value1\"\n  key2 = \"value2\"\n}\n", args; !reflect.DeepEqual(exp, act) {
		t.Errorf("unexpected output exp: '%s', act: '%s'", exp, act)
	}

	vars := map[string]interface{}{"test": []string{"valueone", "valuetwo"}}
	args, err = MapToTerraformTfvars(vars)
	if err != nil {
		t.Error("unexpected error: ", err)
	}
	if exp, act := "test = [\"valueone\", \"valuetwo\"]\n", args; !reflect.DeepEqual(exp, act) {
		t.Errorf("unexpected output exp: '%s', act: '%s'", exp, act)
	}

	// test twice if there's no double double quotes
	args, err = MapToTerraformTfvars(vars)
	if err != nil {
		t.Error("unexpected error: ", err)
	}
	if exp, act := "test = [\"valueone\", \"valuetwo\"]\n", args; !reflect.DeepEqual(exp, act) {
		t.Errorf("unexpected output exp: '%s', act: '%s'", exp, act)
	}
}
