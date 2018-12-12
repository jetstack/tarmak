// Copyright Jetstack Ltd. See LICENSE for details.
package machinedeployment

import (
	"fmt"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"

	"github.com/jetstack/tarmak/pkg/apis/wing"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
)

func NewStrategy(typer runtime.ObjectTyper) machinedeploymentStrategy {
	return machinedeploymentStrategy{typer, names.SimpleNameGenerator}
}

func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, bool, error) {
	apiserver, ok := obj.(*wing.MachineDeployment)
	if !ok {
		return nil, nil, false, fmt.Errorf("given object is not a MachineDeployment.")
	}
	return labels.Set(apiserver.ObjectMeta.Labels), MachineToSelectableFields(apiserver), apiserver.Initializers != nil, nil
}

// MatchMachineDeployment is the filter used by the generic etcd backend to watch events
// from etcd to clients of the apiserver only interested in specific labels/fields.
func MatchMachineDeployment(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

// MachineToSelectableFields returns a field set that represents the object.
func MachineToSelectableFields(obj *wing.MachineDeployment) fields.Set {
	return generic.ObjectMetaFieldsSet(&obj.ObjectMeta, true)
}

type machinedeploymentStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

func (machinedeploymentStrategy) NamespaceScoped() bool {
	return true
}

func (machinedeploymentStrategy) PrepareForCreate(ctx genericapirequest.Context, obj runtime.Object) {
	// TODO: update all none timestamp to now()
}

func (machinedeploymentStrategy) PrepareForUpdate(ctx genericapirequest.Context, obj, old runtime.Object) {
	// TODO: update all none timestamp to now()
}

func (machinedeploymentStrategy) Validate(ctx genericapirequest.Context, obj runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (machinedeploymentStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (machinedeploymentStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (machinedeploymentStrategy) Canonicalize(obj runtime.Object) {
}

func (machinedeploymentStrategy) ValidateUpdate(ctx genericapirequest.Context, obj, old runtime.Object) field.ErrorList {
	return field.ErrorList{}
}
