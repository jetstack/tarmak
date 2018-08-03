// Copyright Jetstack Ltd. See LICENSE for details.
package wingjob

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"

	"github.com/jetstack/tarmak/pkg/apis/wing"
	"github.com/jetstack/tarmak/pkg/wing/registry"
)

// NewREST returns a RESTStorage object that will work against API services.
func NewREST(scheme *runtime.Scheme, optsGetter generic.RESTOptionsGetter) (*registry.REST, error) {
	strategy := NewStrategy(scheme)

	store := &genericregistry.Store{
		NewFunc:                  func() runtime.Object { return &wing.WingJob{} },
		NewListFunc:              func() runtime.Object { return &wing.WingJobList{} },
		PredicateFunc:            MatchWingJob,
		DefaultQualifiedResource: wing.Resource("wingjobs"),

		CreateStrategy: strategy,
		UpdateStrategy: strategy,
		DeleteStrategy: strategy,
	}
	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		return nil, err
	}
	return &registry.REST{
		Store: store,
	}, nil
}
