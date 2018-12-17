// Copyright Jetstack Ltd. See LICENSE for details.
package machine

import (
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/jetstack/tarmak/pkg/apis/wing/common"
	"github.com/jetstack/tarmak/pkg/apis/wing/v1alpha1"
	"github.com/jetstack/tarmak/pkg/wing/interfaces"
)

type Controller struct {
	indexer  cache.Indexer
	queue    workqueue.RateLimitingInterface
	informer cache.Controller
	log      *logrus.Entry
	wing     interfaces.Wing
}

func NewController(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller, wing interfaces.Wing) *Controller {
	return &Controller{
		informer: informer,
		indexer:  indexer,
		queue:    queue,
		log:      wing.Log().WithField("tier", "Machine-controller"),
		wing:     wing,
	}
}

func (c *Controller) processNextItem() bool {
	// Wait until there is a new item in the working queue
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two machines with the same key are never processed in
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

	// ensure only one converge at a time
	c.wing.ConvergeWGWait()

	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		c.log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		// Below we will warm up our cache with a Machine, so that we will see a delete for one machine
		c.log.Infof("Machine %s does not exist anymore\n", key)
		machineAPI := c.wing.Clientset().WingV1alpha1().Machines(c.wing.Flags().ClusterName)
		machine := &v1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Name: c.wing.Flags().MachineName,
				Labels: map[string]string{
					"pool":    c.wing.Flags().Pool,
					"cluster": c.wing.Flags().ClusterName,
				},
			},
			Status: &v1alpha1.MachineStatus{
				Converge: &v1alpha1.MachineStatusManifest{
					State: common.MachineManifestStateConverging,
				},
			},
		}
		_, err := machineAPI.Create(machine)
		if err != nil {
			return fmt.Errorf("error creating machine: %s", err)
		}

		c.log.Infof("Machine created %v", key)
		return nil
	}

	// Note that you also have to check the uid if you have a local controlled resource, which
	// is dependent on the actual machine, to detect that a Machine was recreated with the same name
	machine, ok := obj.(*v1alpha1.Machine)
	if !ok {
		return errors.New("failed to process next item, not a machine")
	}

	// trigger converge if status time is older or not existing
	if machine.Spec != nil && machine.Spec.Converge != nil && !machine.Spec.Converge.RequestTimestamp.Time.IsZero() {
		if machine.Status != nil && machine.Status.Converge != nil && !machine.Status.Converge.LastUpdateTimestamp.Time.IsZero() {
			if machine.Status.Converge.LastUpdateTimestamp.Time.After(machine.Spec.Converge.RequestTimestamp.Time) {
				c.log.Debug("no converge neccessary, last update was after request")
				return nil
			}
		} else {
			c.log.Debug("no converge neccessary, no status section found or update timestamp zero")
			return nil
		}
	}

	if machine.Status != nil && machine.Status.Converge != nil &&
		machine.Status.Converge.State != common.MachineManifestStateConverging &&
		machine.Status.Converge.State != common.MachineManifestStateConverged {
		c.log.Infof("running converge")
		c.wing.Converge()
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
		c.log.Errorf("Error syncing machine %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	c.log.Errorf("Dropping machine %q out of the queue: %v", key, err)
}

func (c *Controller) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.queue.ShutDown()
	c.log.Info("Starting Machine controller")

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
	c.log.Info("Stopping Machine controller")
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}
