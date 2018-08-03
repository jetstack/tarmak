// Copyright Jetstack Ltd. See LICENSE for details.
package machine

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

func NewStrategy(typer runtime.ObjectTyper) machineStrategy {
	return machineStrategy{typer, names.SimpleNameGenerator}
}

func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, bool, error) {
	apiserver, ok := obj.(*wing.Machine)
	if !ok {
		return nil, nil, false, fmt.Errorf("given object is not a Machine.")
	}
	return labels.Set(apiserver.ObjectMeta.Labels), MachineToSelectableFields(apiserver), apiserver.Initializers != nil, nil
}

// MatchMachine is the filter used by the generic etcd backend to watch events
// from etcd to clients of the apiserver only interested in specific labels/fields.
func MatchMachine(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
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

type machineStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

func (machineStrategy) NamespaceScoped() bool {
	return true
}

func (machineStrategy) PrepareForCreate(ctx genericapirequest.Context, obj runtime.Object) {
	// TODO: update all none timestamp to now()
}

func (machineStrategy) PrepareForUpdate(ctx genericapirequest.Context, obj, old runtime.Object) {
	// TODO: update all none timestamp to now()
}

func (machineStrategy) Validate(ctx genericapirequest.Context, obj runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (machineStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (machineStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (machineStrategy) Canonicalize(obj runtime.Object) {
}

func (machineStrategy) ValidateUpdate(ctx genericapirequest.Context, obj, old runtime.Object) field.ErrorList {
	return field.ErrorList{}
}
