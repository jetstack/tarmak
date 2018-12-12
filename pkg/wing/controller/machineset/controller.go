// Copyright Jetstack Ltd. See LICENSE for details.
package machineset

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/jetstack/tarmak/pkg/apis/wing/v1alpha1"
	clientset "github.com/jetstack/tarmak/pkg/wing/client/clientset/versioned"
)

type Controller struct {
	indexer  cache.Indexer
	queue    workqueue.RateLimitingInterface
	informer cache.Controller
	log      *logrus.Entry
	client   *clientset.Clientset
}

func NewController(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller, client *clientset.Clientset) *Controller {
	return &Controller{
		informer: informer,
		indexer:  indexer,
		queue:    queue,
		log:      logrus.NewEntry(logrus.New()).WithField("tier", "controller"),
		client:   client,
	}
}

func (c *Controller) processNextItem() bool {
	// Wait until there is a new item in the working queue
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two machinesets with the same key are never processed in
	// parallel.
	defer c.queue.Done(key)

	// Invoke the method containing the business logic
	err := c.syncToStdout(key.(string))
	// Handle the error if something went wrong during the execution of the business logic
	c.handleErr(err, key)
	return true
}

// syncToStdout is the business logic of the controller. In this controller it simply prints
// information about the machine to stdout. In case an error happened, it has to simply return the error.
// The retry logic should not be part of the business logic.
func (c *Controller) syncToStdout(key string) error {

	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		c.log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		c.log.Infof("MachineSet %s does not exist anymore\n", key)
		return nil
	}

	// Note that you also have to check the uid if you have a local controlled resource, which
	// is dependent on the actual machine, to detect that a Machine was recreated with the same name
	machineset, ok := obj.(*v1alpha1.MachineSet)
	if !ok {
		return errors.New("failed to process next item, not a machineset")
	}

	if machineset.Spec == nil {
		return fmt.Errorf("machineset spec is nil: %v", machineset.Spec)
	}

	if machineset.Spec.MaxReplicas == nil || machineset.Spec.MinReplicas == nil {
		return fmt.Errorf("expected machineset min and max replicas to not be nil: min=%v max=%v",
			machineset.Spec.MaxReplicas, machineset.Spec.MinReplicas)
	}

	var selectors []string
	for k, v := range machineset.Spec.Selector.MatchLabels {
		selectors = append(selectors, fmt.Sprintf("%s=%s", k, v))
	}

	machineAPI := c.client.WingV1alpha1().Machines(machineset.Namespace)
	machineList, err := machineAPI.List(metav1.ListOptions{
		LabelSelector: strings.Join(selectors, ","),
	})
	if err != nil {
		return err
	}

	if int32(len(machineList.Items)) > *machineset.Spec.MaxReplicas {
		c.log.Warnf("more machines exist then the maximum, max=%v curr=%v", *machineset.Spec.MaxReplicas, len(machineList.Items))
	}

	//if int32(len(machineList.Items)) < *machineset.Spec.MinReplicas {
	//	c.log.Warnf("less machines exist then the minimum, min=%v curr=%v", *machineset.Spec.MinReplicas, len(machineList.Items))
	//}

	var readyMachines int32
	var fullyLabeledMachines int32
	for _, i := range machineList.Items {
		if c.machineConverged(i) {
			readyMachines++
		}

		if c.fullyLabeledMachine(machineset, i) {
			fullyLabeledMachines++
		}
	}

	status := &v1alpha1.MachineSetStatus{
		Replicas:             int32(len(machineList.Items)),
		ObservedGeneration:   machineset.Status.ObservedGeneration + 1,
		ReadyReplicas:        readyMachines,
		FullyLabeledReplicas: fullyLabeledMachines,
		// we may want to do something with available replicas
		AvailableReplicas: readyMachines,
	}

	machineSetAPI := c.client.WingV1alpha1().MachineSets(machineset.Namespace)
	machineset.Status = status.DeepCopy()
	_, err = machineSetAPI.Update(machineset)
	if err != nil {
		return fmt.Errorf("failed to update machineset: %s", err)
	}

	return nil
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *Controller) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.queue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.queue.NumRequeues(key) < 5 {
		c.log.Errorf("Error syncing machineset %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	c.log.Errorf("Dropping machineset %q out of the queue: %v", key, err)
}

func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.queue.ShutDown()
	c.log.Info("Starting MachineSet controller")

	go c.informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	c.log.Info("Stopping MachineSet controller")
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}

func (c *Controller) machineConverged(machine v1alpha1.Machine) bool {
	if machine.Spec != nil && machine.Spec.Converge != nil && !machine.Spec.Converge.RequestTimestamp.Time.IsZero() {
		if machine.Status != nil && machine.Status.Converge != nil && !machine.Status.Converge.LastUpdateTimestamp.Time.IsZero() {
			if machine.Status.Converge.LastUpdateTimestamp.Time.After(machine.Spec.Converge.RequestTimestamp.Time) {
				return true
			}
		} else {
			return true
		}
	} else {
		return true
	}

	return false
}

func (c *Controller) fullyLabeledMachine(set *v1alpha1.MachineSet, machine v1alpha1.Machine) bool {
	for k, v := range set.Spec.Selector.MatchLabels {
		a, ok := machine.Labels[k]
		if !ok || a != v {
			return false
		}
	}

	return true
}
