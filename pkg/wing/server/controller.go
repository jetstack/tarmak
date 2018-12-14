// Copyright Jetstack Ltd. See LICENSE for details.
package server

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/jetstack/tarmak/pkg/apis/wing/v1alpha1"
	"github.com/jetstack/tarmak/pkg/wing/controller/machinedeployment"
	"github.com/jetstack/tarmak/pkg/wing/controller/machineset"
)

func (o WingServerOptions) StartMachineControllers() error {

	// machineset controller needs to watch for changes to machines
	machineListWatcher := cache.NewListWatchFromClient(o.client.WingV1alpha1().RESTClient(), "machines", metav1.NamespaceAll, fields.Everything())
	queueSet := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	indexerMachine, informerMachine := o.newIndexInfomer(machineListWatcher, &v1alpha1.Machine{}, queueSet)
	machinesetController := machineset.NewController(queueSet, indexerMachine, informerMachine, o.client)

	// machindeployment controller needs to watch for changes to machinedeployments and machinesets
	machinesetListWatcher := cache.NewListWatchFromClient(o.client.WingV1alpha1().RESTClient(), "machinesets", metav1.NamespaceAll, fields.Everything())
	machinedeploymentListWatcher := cache.NewListWatchFromClient(o.client.WingV1alpha1().RESTClient(), "machinedeployments", metav1.NamespaceAll, fields.Everything())
	queueDeployment := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	indexerSet, informerSet := o.newIndexInfomer(machinesetListWatcher, &v1alpha1.MachineSet{}, queueDeployment)
	indexerDeployment, informerDeployment := o.newIndexInfomer(machinedeploymentListWatcher, &v1alpha1.MachineDeployment{}, queueDeployment)
	machinedeploymentController := machinedeployment.NewController(queueDeployment, indexerDeployment, indexerSet, informerDeployment, informerSet, o.client)

	//// Now let's start the controllers
	go machinesetController.Run(1, o.stopCh)
	go machinedeploymentController.Run(1, o.stopCh)
	return nil
}

func (o WingServerOptions) newIndexInfomer(listWatcher cache.ListerWatcher, objType runtime.Object, queue workqueue.RateLimitingInterface) (cache.Indexer, cache.Controller) {
	return cache.NewIndexerInformer(listWatcher, objType, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.AddAfter(key, 2*time.Second)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.AddAfter(key, 2*time.Second)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.AddAfter(key, 2*time.Second)
			}
		},
	}, cache.Indexers{})
}
