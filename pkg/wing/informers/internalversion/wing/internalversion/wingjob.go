// Copyright Jetstack Ltd. See LICENSE for details.

// This file was automatically generated by informer-gen

package internalversion

import (
	wing "github.com/jetstack/tarmak/pkg/apis/wing"
	internalclientset "github.com/jetstack/tarmak/pkg/wing/clients/internalclientset"
	internalinterfaces "github.com/jetstack/tarmak/pkg/wing/informers/internalversion/internalinterfaces"
	internalversion "github.com/jetstack/tarmak/pkg/wing/listers/wing/internalversion"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
	time "time"
)

// WingJobInformer provides access to a shared informer and lister for
// WingJobs.
type WingJobInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() internalversion.WingJobLister
}

type wingJobInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewWingJobInformer constructs a new informer for WingJob type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewWingJobInformer(client internalclientset.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredWingJobInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredWingJobInformer constructs a new informer for WingJob type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredWingJobInformer(client internalclientset.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.Wing().WingJobs(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.Wing().WingJobs(namespace).Watch(options)
			},
		},
		&wing.WingJob{},
		resyncPeriod,
		indexers,
	)
}

func (f *wingJobInformer) defaultInformer(client internalclientset.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredWingJobInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *wingJobInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&wing.WingJob{}, f.defaultInformer)
}

func (f *wingJobInformer) Lister() internalversion.WingJobLister {
	return internalversion.NewWingJobLister(f.Informer().GetIndexer())
}
