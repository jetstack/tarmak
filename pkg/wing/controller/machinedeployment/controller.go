// Copyright Jetstack Ltd. See LICENSE for details.
package machinedeployment

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
	"github.com/jetstack/tarmak/pkg/tarmak/utils"
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
	// This allows safe parallel processing because two machinedeployments with the same key are never processed in
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
		fmt.Printf("MachineDeployment %s does not exist anymore\n", key)
		return nil
	}

	// Note that you also have to check the uid if you have a local controlled resource, which
	// is dependent on the actual machine, to detect that a Machine was recreated with the same name
	machinedeployment, ok := obj.(*v1alpha1.MachineDeployment)
	if !ok {
		return errors.New("failed to process next item, not a machinedeployment")
	}
	machinedeployment = machinedeployment.DeepCopy()

	var selectors []string
	for k, v := range machinedeployment.Spec.Selector.MatchLabels {
		selectors = append(selectors, fmt.Sprintf("%s=%s", k, v))
	}

	machinesetAPI := c.client.WingV1alpha1().MachineSets(machinedeployment.Namespace)
	machinesetList, err := machinesetAPI.List(metav1.ListOptions{
		LabelSelector: strings.Join(selectors, ","),
	})
	if err != nil {
		return err
	}
	machinesetList = machinesetList.DeepCopy()

	if len(machinesetList.Items) > 1 {
		return fmt.Errorf("there can only be a maximum of one MachineSet per MachineDeployment, got=%d", len(machinesetList.Items))
	}

	// we need to create the machine set for this deployment
	if len(machinesetList.Items) == 0 {
		fmt.Printf("Creating MachineSet for MachineDeployment %s\n", key)
		machineset := &v1alpha1.MachineSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:   machinedeployment.Name,
				Labels: utils.DuplicateMapString(machinedeployment.Labels),
			},
			Spec: &v1alpha1.MachineSetSpec{
				MaxReplicas:     machinedeployment.Spec.MaxReplicas,
				MinReadySeconds: *machinedeployment.Spec.MinReadySeconds,
				MinReplicas:     machinedeployment.Spec.MinReplicas,
				Selector:        machinedeployment.Spec.Selector,
			},
		}

		_, err := machinesetAPI.Create(machineset)
		if err != nil {
			return fmt.Errorf("error creating machine: %s", err)
		}

		return nil
	}

	machineset := machinesetList.Items[0]
	status := &v1alpha1.MachineDeploymentStatus{
		Replicas:           machineset.Status.Replicas,
		ObservedGeneration: machinedeployment.Status.ObservedGeneration + 1,
		ReadyReplicas:      machineset.Status.ReadyReplicas,
		// we may want to do something with available replicas
		AvailableReplicas:   machineset.Status.ReadyReplicas,
		UnavailableReplicas: machineset.Status.Replicas - machineset.Status.ReadyReplicas,
	}

	machineDeploymentAPI := c.client.WingV1alpha1().MachineDeployments(machinedeployment.Namespace)
	machinedeployment.Status = status.DeepCopy()
	_, err = machineDeploymentAPI.Update(machinedeployment)
	if err != nil {
		return fmt.Errorf("failed to update machinedeployment: %s", err)
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
		c.log.Errorf("Error syncing machine deployment %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	c.log.Errorf("Dropping machine deployment %q out of the queue: %v", key, err)
}

func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.queue.ShutDown()
	c.log.Info("Starting MachineDeployment controller")

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
	c.log.Info("Stopping MachineDeployment controller")
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}
