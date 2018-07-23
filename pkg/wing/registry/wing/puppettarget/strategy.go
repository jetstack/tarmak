// Copyright Jetstack Ltd. See LICENSE for details.
package puppettarget

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

func NewStrategy(typer runtime.ObjectTyper) puppetTargetStrategy {
	return puppetTargetStrategy{typer, names.SimpleNameGenerator}
}

func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, bool, error) {
	apiserver, ok := obj.(*wing.PuppetTarget)
	if !ok {
		return nil, nil, false, fmt.Errorf("given object is not a PuppetTarget.")
	}
	return labels.Set(apiserver.ObjectMeta.Labels), PuppetTargetToSelectableFields(apiserver), apiserver.Initializers != nil, nil
}

// MatchPuppetTarget is the filter used by the generic etcd backend to watch events
// from etcd to clients of the apiserver only interested in specific labels/fields.
func MatchPuppetTarget(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

// PuppetTargetToSelectableFields returns a field set that represents the object.
func PuppetTargetToSelectableFields(obj *wing.PuppetTarget) fields.Set {
	return generic.ObjectMetaFieldsSet(&obj.ObjectMeta, true)
}

type puppetTargetStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

func (puppetTargetStrategy) NamespaceScoped() bool {
	return true
}

func (puppetTargetStrategy) PrepareForCreate(ctx genericapirequest.Context, obj runtime.Object) {
	// TODO: update all none timestamp to now()
}

func (puppetTargetStrategy) PrepareForUpdate(ctx genericapirequest.Context, obj, old runtime.Object) {
	// TODO: update all none timestamp to now()
}

func (puppetTargetStrategy) Validate(ctx genericapirequest.Context, obj runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

func (puppetTargetStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (puppetTargetStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (puppetTargetStrategy) Canonicalize(obj runtime.Object) {
}

func (puppetTargetStrategy) ValidateUpdate(ctx genericapirequest.Context, obj, old runtime.Object) field.ErrorList {
	return field.ErrorList{}
}
