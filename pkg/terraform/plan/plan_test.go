// Copyright Jetstack Ltd. See LICENSE for details.
package plan

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform/terraform"
)

func expResources() map[string]bool {

	return map[string]bool{
		"module.etcd.aws_ebs_volume.volume.0": false,
		"module.etcd.aws_ebs_volume.volume.1": false,
		"module.etcd.aws_ebs_volume.volume.2": false,
	}

}

func openReadPlan(t *testing.T, testCase string) *terraform.Plan {
	plan, err := Open(testCase)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	return plan
}

func TestIsDestroyedCreate(t *testing.T) {
	isDestroyed, resourceNames := IsDestroyingEBSVolume(openReadPlan(t, "test_data/create.plan"))

	if exp, act := false, isDestroyed; exp != act {
		t.Errorf("unexpected value exp=%+v\n act=%+v\n", exp, act)
	}

	if len(resourceNames) != 0 {
		t.Errorf("unexpected resourceNames returned %+v", resourceNames)
	}

}

func TestIsDestroyedTainted(t *testing.T) {
	isDestroyed, resourceNames := IsDestroyingEBSVolume(openReadPlan(t, "test_data/tainted.plan"))

	if exp, act := true, isDestroyed; exp != act {
		t.Errorf("unexpected value exp=%+v\n act=%+v\n", exp, act)
	}

	if len(resourceNames) != 1 {
		t.Errorf("unexpected resourceNames returned %+v", resourceNames)
	}

	if exp, act := []string{"module.etcd.aws_ebs_volume.volume.0"}, resourceNames; !reflect.DeepEqual(exp, act) {
		t.Errorf("unexpected slice exp=%+v act=%+v", exp, act)
	}

}

func TestIsDestroyedModify(t *testing.T) {
	isDestroyed, resourceNames := IsDestroyingEBSVolume(openReadPlan(t, "test_data/modify.plan"))

	if exp, act := false, isDestroyed; exp != act {
		t.Errorf("unexpected value exp=%+v\n act=%+v\n", exp, act)
	}

	if len(resourceNames) != 0 {
		t.Errorf("unexpected resourceNames returned %+v", resourceNames)
	}

}

func TestIsDestroyedDestroy(t *testing.T) {
	isDestroyed, resourceNames := IsDestroyingEBSVolume(openReadPlan(t, "test_data/destroy.plan"))

	if exp, act := true, isDestroyed; exp != act {
		t.Errorf("unexpected value exp=%+v\n act=%+v\n", exp, act)
	}

	if len(resourceNames) != 3 {
		t.Errorf("unexpected resourceNames returned %+v", resourceNames)

	}

	for _, resource := range resourceNames {
		if found, ok := expResources()[resource]; ok && !found {
			expResources()[resource] = true
		} else {
			t.Errorf("unexpected slice act=%+v", resource)
		}
	}
}

func TestIsDestroyedRecreate(t *testing.T) {
	isDestroyed, resourceNames := IsDestroyingEBSVolume(openReadPlan(t, "test_data/recreate.plan"))

	if exp, act := true, isDestroyed; exp != act {
		t.Errorf("unexpected value exp=%+v\n act=%+v\n", exp, act)
	}

	if len(resourceNames) != 3 {
		t.Errorf("expected resourceNames returned %+v", resourceNames)
	}

	for _, resource := range resourceNames {
		if found, ok := expResources()[resource]; ok && !found {
			expResources()[resource] = true
		} else {
			t.Errorf("unexpected slice  act=%+v", resource)
		}
	}

}

func TestIsDestroyedNochanges(t *testing.T) {
	isDestroyed, resourceNames := IsDestroyingEBSVolume(openReadPlan(t, "test_data/nochanges.plan"))

	if exp, act := false, isDestroyed; exp != act {
		t.Errorf("unexpected value exp=%+v\n act=%+v\n", exp, act)
	}

	if len(resourceNames) != 0 {
		t.Errorf("expected resourceNames returned %+v", resourceNames)
	}
}

func TestIsDestroyedNonEbs(t *testing.T) {
	isDestroyed, resourceNames := IsDestroyingEBSVolume(openReadPlan(t, "test_data/destroy_non_ebs.plan"))

	if exp, act := false, isDestroyed; exp != act {
		t.Errorf("unexpected value exp=%+v\n act=%+v\n", exp, act)
	}

	if len(resourceNames) != 0 {
		t.Errorf("expected resourceNames returned %+v", resourceNames)
	}

}

func TestIsDestroyedDestroyNonModuleEbs(t *testing.T) {
	isDestroyed, resourceNames := IsDestroyingEBSVolume(openReadPlan(t, "test_data/destroy_non_module_ebs.plan"))

	if exp, act := true, isDestroyed; exp != act {
		t.Errorf("unexpected value exp=%+v\n act=%+v\n", exp, act)
	}
	if exp, act := []string{"aws_ebs_volume.extra"}, resourceNames; !reflect.DeepEqual(exp, act) {
		t.Errorf("unexpected slice exp=%+v act=%+v", exp, act)
	}
}
