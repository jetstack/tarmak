// Copyright Jetstack Ltd. See LICENSE for details.
package server

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/jetstack/tarmak/pkg/apis/wing/v1alpha1"
	"github.com/jetstack/tarmak/pkg/wing/controller/machinedeployment"
	"github.com/jetstack/tarmak/pkg/wing/controller/machineset"
)

func (o WingServerOptions) StartMachineControllers() error {
	machinesetListWatcher := cache.NewListWatchFromClient(o.client.WingV1alpha1().RESTClient(), "machinesets", metav1.NamespaceAll, fields.Everything())
	machinedeploymentListWatcher := cache.NewListWatchFromClient(o.client.WingV1alpha1().RESTClient(), "machinedeployments", metav1.NamespaceAll, fields.Everything())
	queueSet := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	queueDeployment := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	indexerSet, informerSet := cache.NewIndexerInformer(machinesetListWatcher, &v1alpha1.MachineSet{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queueSet.AddAfter(key, 2*time.Second)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queueSet.AddAfter(key, 2*time.Second)
				queueDeployment.AddAfter(key, 2*time.Second)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queueSet.AddAfter(key, 2*time.Second)
			}
		},
	}, cache.Indexers{})
	machinesetController := machineset.NewController(queueSet, indexerSet, informerSet, o.client)

	indexerDeployment, informerDeployment := cache.NewIndexerInformer(machinedeploymentListWatcher, &v1alpha1.MachineDeployment{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queueDeployment.AddAfter(key, 2*time.Second)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queueDeployment.AddAfter(key, 2*time.Second)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queueDeployment.AddAfter(key, 2*time.Second)
			}
		},
	}, cache.Indexers{})
	machinedeploymentController := machinedeployment.NewController(queueDeployment, indexerDeployment, informerDeployment, o.client)

	//// Now let's start the controllers
	go machinesetController.Run(1, o.stopCh)
	go machinedeploymentController.Run(1, o.stopCh)
	return nil
}
