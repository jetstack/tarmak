// Copyright Jetstack Ltd. See LICENSE for details.
package machineset

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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

func (c *Controller) syncToStdout(key string) error {
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		c.log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		c.log.Infof("Machine %s does not exist anymore\n", key)
		return nil
	}

	m, ok := obj.(*v1alpha1.Machine)
	if !ok {
		return errors.New("failed to process next item, not a machine")
	}

	c.log.Debugf("machineset controller got a mahcine: %s", m.Name)

	ms, found, err := c.getMachineSet(m)
	if err != nil {
		return err
	}

	// machineset doesn't exist for this machine
	if !found {
		c.log.Warnf("did not find machineset for machine %s", m.Name)
		return nil
	}

	c.log.Debugf("machineset controller got matching machineset: %s", ms.Name)

	if ms.Spec == nil {
		return fmt.Errorf("machineset spec is nil: %v", ms.Spec)
	}

	if ms.Spec.MaxReplicas == nil || ms.Spec.MinReplicas == nil {
		return fmt.Errorf("expected machineset min and max replicas to not be nil: min=%v max=%v",
			ms.Spec.MaxReplicas, ms.Spec.MinReplicas)
	}

	var selectors []string
	for k, v := range ms.Spec.Selector.MatchLabels {
		selectors = append(selectors, fmt.Sprintf("%s=%s", k, v))
	}

	// find all the machines coverged by this machien set and update status accordingly
	mAPI := c.client.WingV1alpha1().Machines(ms.Namespace)
	mList, err := mAPI.List(metav1.ListOptions{
		LabelSelector: strings.Join(selectors, ","),
	})
	if err != nil {
		return err
	}

	if int32(len(mList.Items)) > *ms.Spec.MaxReplicas {
		c.log.Warnf("more machines exist then the maximum, max=%v curr=%v", *ms.Spec.MaxReplicas, len(mList.Items))
	}

	//if int32(len(mList.Items)) < *ms.Spec.MinReplicas {
	//	c.log.Warnf("less machines exist then the minimum, min=%v curr=%v", *ms.Spec.MinReplicas, len(m.Items))
	//}

	var readyMachines int32
	var fullyLabeledMachines int32
	for _, i := range mList.Items {
		if c.machineConverged(i) {
			readyMachines++
		}

		if c.fullyLabeledMachine(ms, i) {
			fullyLabeledMachines++
		}
	}

	var observedGeneration int64 = 0
	if ms.Status != nil {
		observedGeneration = ms.Status.ObservedGeneration
	}

	status := &v1alpha1.MachineSetStatus{
		Replicas:             int32(len(mList.Items)),
		ObservedGeneration:   observedGeneration + 1,
		ReadyReplicas:        readyMachines,
		FullyLabeledReplicas: fullyLabeledMachines,
		// we may want to do something with available replicas
		AvailableReplicas: readyMachines,
	}

	c.log.Debugf("updating machine set status: %+v", status)

	msAPI := c.client.WingV1alpha1().MachineSets(ms.Namespace)
	ms.Status = status.DeepCopy()
	_, err = msAPI.Update(ms)
	if err != nil {
		return fmt.Errorf("failed to update machineset: %s", err)
	}

	return nil
}

func (c *Controller) getMachineSet(m *v1alpha1.Machine) (*v1alpha1.MachineSet, bool, error) {
	pool, ok := m.Labels["pool"]
	if !ok {
		return nil, false, nil
	}

	cluster, ok := m.Labels["cluster"]
	if !ok {
		return nil, false, nil
	}

	msAPI := c.client.WingV1alpha1().MachineSets(cluster)
	ms, err := msAPI.Get(pool, metav1.GetOptions{})
	if err != nil {
		if kerr, ok := err.(*apierrors.StatusError); ok && kerr.ErrStatus.Reason == metav1.StatusReasonNotFound {
			return nil, false, nil
		} else {
			return nil, false, err
		}
	}

	return ms, true, nil
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
	if machine.Status != nil && machine.Status.Converge != nil && machine.Status.Converge.State == v1alpha1.MachineManifestStateConverged {
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
