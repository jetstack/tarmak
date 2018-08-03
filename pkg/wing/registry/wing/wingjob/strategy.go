// Copyright Jetstack Ltd. See LICENSE for details.
package wingjob

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

func NewStrategy(typer runtime.ObjectTyper) wingJobStrategy {
	return wingJobStrategy{typer, names.SimpleNameGenerator}
}

func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, bool, error) {
	apiserver, ok := obj.(*wing.WingJob)
	if !ok {
		return nil, nil, false, fmt.Errorf("given object is not a WingJob.")
	}
	return labels.Set(apiserver.ObjectMeta.Labels), WingJobToSelectableFields(apiserver), apiserver.Initializers != nil, nil
}

// MatchWingJob is the filter used by the generic etcd backend to watch events
// from etcd to clients of the apiserver only interested in specific labels/fields.
func MatchWingJob(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

// WingJobToSelectableFields returns a field set that represents the object.
func WingJobToSelectableFields(obj *wing.WingJob) fields.Set {
	return generic.MergeFieldsSets(
		generic.ObjectMetaFieldsSet(&obj.ObjectMeta, true),
		fields.Set{
			"spec.instanceName": obj.Spec.InstanceName,
		},
	)
}

type wingJobStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

func (wingJobStrategy) NamespaceScoped() bool {
	return true
}

func (wingJobStrategy) PrepareForCreate(ctx genericapirequest.Context, obj runtime.Object) {
	// TODO: update all none timestamp to now()
}

func (wingJobStrategy) PrepareForUpdate(ctx genericapirequest.Context, obj, old runtime.Object) {
	// TODO: update all none timestamp to now()
}

func (wingJobStrategy) Validate(ctx genericapirequest.Context, obj runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (wingJobStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (wingJobStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (wingJobStrategy) Canonicalize(obj runtime.Object) {
}

func (wingJobStrategy) ValidateUpdate(ctx genericapirequest.Context, obj, old runtime.Object) field.ErrorList {
	return field.ErrorList{}
}
