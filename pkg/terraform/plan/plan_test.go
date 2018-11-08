// Copyright Jetstack Ltd. See LICENSE for details.
package plan

import (
	"reflect"
	"testing"
)

func expResources() map[string]bool {

	return map[string]bool{
		"module.etcd.aws_ebs_volume.volume.0": false,
		"module.etcd.aws_ebs_volume.volume.1": false,
		"module.etcd.aws_ebs_volume.volume.2": false,
	}

}

func NewTest(t *testing.T, testCase string) *Plan {
	plan, err := New(testCase)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	return plan
}

func TestIsDestroyedCreate(t *testing.T) {
	isDestroyed, resourceNames := NewTest(t, "test_data/create.plan").IsDestroyingEBSVolume()

	if exp, act := false, isDestroyed; exp != act {
		t.Errorf("unexpected value exp=%+v\n act=%+v\n", exp, act)
	}

	if len(resourceNames) != 0 {
		t.Errorf("unexpected resourceNames returned %+v", resourceNames)
	}

}

func TestIsDestroyedTainted(t *testing.T) {
	isDestroyed, resourceNames := NewTest(t, "test_data/tainted.plan").IsDestroyingEBSVolume()

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
	isDestroyed, resourceNames := NewTest(t, "test_data/modify.plan").IsDestroyingEBSVolume()

	if exp, act := false, isDestroyed; exp != act {
		t.Errorf("unexpected value exp=%+v\n act=%+v\n", exp, act)
	}

	if len(resourceNames) != 0 {
		t.Errorf("unexpected resourceNames returned %+v", resourceNames)
	}

}

func TestIsDestroyedDestroy(t *testing.T) {
	isDestroyed, resourceNames := NewTest(t, "test_data/destroy.plan").IsDestroyingEBSVolume()

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
	isDestroyed, resourceNames := NewTest(t, "test_data/recreate.plan").IsDestroyingEBSVolume()

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
	isDestroyed, resourceNames := NewTest(t, "test_data/nochanges.plan").IsDestroyingEBSVolume()

	if exp, act := false, isDestroyed; exp != act {
		t.Errorf("unexpected value exp=%+v\n act=%+v\n", exp, act)
	}

	if len(resourceNames) != 0 {
		t.Errorf("expected resourceNames returned %+v", resourceNames)
	}
}

func TestIsDestroyedNonEbs(t *testing.T) {
	isDestroyed, resourceNames := NewTest(t, "test_data/destroy_non_ebs.plan").IsDestroyingEBSVolume()

	if exp, act := false, isDestroyed; exp != act {
		t.Errorf("unexpected value exp=%+v\n act=%+v\n", exp, act)
	}

	if len(resourceNames) != 0 {
		t.Errorf("expected resourceNames returned %+v", resourceNames)
	}

}

func TestIsDestroyedDestroyNonModuleEbs(t *testing.T) {
	isDestroyed, resourceNames := NewTest(t, "test_data/destroy_non_module_ebs.plan").IsDestroyingEBSVolume()

	if exp, act := true, isDestroyed; exp != act {
		t.Errorf("unexpected value exp=%+v\n act=%+v\n", exp, act)
	}
	if exp, act := []string{"aws_ebs_volume.extra"}, resourceNames; !reflect.DeepEqual(exp, act) {
		t.Errorf("unexpected slice exp=%+v act=%+v", exp, act)
	}
}
