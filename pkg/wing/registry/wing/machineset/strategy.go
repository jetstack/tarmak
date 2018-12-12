// Copyright Jetstack Ltd. See LICENSE for details.
package machineset

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

func NewStrategy(typer runtime.ObjectTyper) machinesetStrategy {
	return machinesetStrategy{typer, names.SimpleNameGenerator}
}

func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, bool, error) {
	apiserver, ok := obj.(*wing.Machine)
	if !ok {
		return nil, nil, false, fmt.Errorf("given object is not a Machine.")
	}
	return labels.Set(apiserver.ObjectMeta.Labels), MachineToSelectableFields(apiserver), apiserver.Initializers != nil, nil
}

// MatchMachineSet is the filter used by the generic etcd backend to watch events
// from etcd to clients of the apiserver only interested in specific labels/fields.
func MatchMachineSet(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

// MachineToSelectableFields returns a field set that represents the object.
func MachineToSelectableFields(obj *wing.Machine) fields.Set {
	return generic.ObjectMetaFieldsSet(&obj.ObjectMeta, true)
}

type machinesetStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

func (machinesetStrategy) NamespaceScoped() bool {
	return true
}

func (machinesetStrategy) PrepareForCreate(ctx genericapirequest.Context, obj runtime.Object) {
	// TODO: update all none timestamp to now()
}

func (machinesetStrategy) PrepareForUpdate(ctx genericapirequest.Context, obj, old runtime.Object) {
	// TODO: update all none timestamp to now()
}

func (machinesetStrategy) Validate(ctx genericapirequest.Context, obj runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (machinesetStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (machinesetStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (machinesetStrategy) Canonicalize(obj runtime.Object) {
}

func (machinesetStrategy) ValidateUpdate(ctx genericapirequest.Context, obj, old runtime.Object) field.ErrorList {
	return field.ErrorList{}
}
