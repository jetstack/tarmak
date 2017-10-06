// Copyright Jetstack Ltd. See LICENSE for details.
package instance

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

func NewStrategy(typer runtime.ObjectTyper) instanceStrategy {
	return instanceStrategy{typer, names.SimpleNameGenerator}
}

func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, bool, error) {
	apiserver, ok := obj.(*wing.Instance)
	if !ok {
		return nil, nil, false, fmt.Errorf("given object is not a Instance.")
	}
	return labels.Set(apiserver.ObjectMeta.Labels), InstanceToSelectableFields(apiserver), apiserver.Initializers != nil, nil
}

// MatchInstance is the filter used by the generic etcd backend to watch events
// from etcd to clients of the apiserver only interested in specific labels/fields.
func MatchInstance(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

// InstanceToSelectableFields returns a field set that represents the object.
func InstanceToSelectableFields(obj *wing.Instance) fields.Set {
	return generic.ObjectMetaFieldsSet(&obj.ObjectMeta, true)
}

type instanceStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

func (instanceStrategy) NamespaceScoped() bool {
	return true
}

func (instanceStrategy) PrepareForCreate(ctx genericapirequest.Context, obj runtime.Object) {
	// TODO: update all none timestamp to now()
}

func (instanceStrategy) PrepareForUpdate(ctx genericapirequest.Context, obj, old runtime.Object) {
	// TODO: update all none timestamp to now()
}

func (instanceStrategy) Validate(ctx genericapirequest.Context, obj runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (instanceStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (instanceStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (instanceStrategy) Canonicalize(obj runtime.Object) {
}

func (instanceStrategy) ValidateUpdate(ctx genericapirequest.Context, obj, old runtime.Object) field.ErrorList {
	return field.ErrorList{}
}
