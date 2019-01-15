// Copyright Jetstack Ltd. See LICENSE for details.
package instance

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"

	"github.com/jetstack/tarmak/pkg/apis/wing"
)

// NewStrategy creates and returns a instanceStrategy instance
func NewStrategy(typer runtime.ObjectTyper) instanceStrategy {
	return instanceStrategy{typer, names.SimpleNameGenerator}
}

// GetAttrs returns labels.Set, fields.Set, the presence of Initializers if any
// and error in case the given runtime.Object is not a Instance
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, bool, error) {
	apiserver, ok := obj.(*wing.Instance)
	if !ok {
		return nil, nil, false, fmt.Errorf("given object is not a Instance")
	}
	return labels.Set(apiserver.ObjectMeta.Labels), SelectableFields(apiserver), apiserver.Initializers != nil, nil
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

// SelectableFields returns a field set that represents the object.
func SelectableFields(obj *wing.Instance) fields.Set {
	return generic.ObjectMetaFieldsSet(&obj.ObjectMeta, true)
}

type instanceStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

func (instanceStrategy) NamespaceScoped() bool {
	return true
}

func (instanceStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

func (instanceStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

func (instanceStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return nil
}

func (instanceStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (instanceStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (instanceStrategy) Canonicalize(obj runtime.Object) {
}

func (instanceStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return field.ErrorList{}
}
