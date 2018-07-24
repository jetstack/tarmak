// Copyright Jetstack Ltd. See LICENSE for details.
package wing

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/jetstack/tarmak/pkg/apis/wing/v1alpha1"
)

type InstanceController struct {
	indexer  cache.Indexer
	queue    workqueue.RateLimitingInterface
	informer cache.Controller
	log      *logrus.Entry
	wing     *Wing
}

func NewInstanceController(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller, wing *Wing) *InstanceController {
	return &InstanceController{
		informer: informer,
		indexer:  indexer,
		queue:    queue,
		log:      wing.log.WithField("tier", "InstanceController"),
		wing:     wing,
	}
}

func (c *InstanceController) processNextItem() bool {
	// Wait until there is a new item in the working queue
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two instances with the same key are never processed in
	// parallel.
	defer c.queue.Done(key)

	// Invoke the method containing the business logic
	err := c.syncInstance(key.(string))
	// Handle the error if something went wrong during the execution of the business logic
	c.handleErr(err, key)
	return true
}

func (c *InstanceController) syncInstance(key string) error {

	// ensure only one converge at a time
	c.wing.convergeWG.Wait()

	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		c.log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		// Below we will warm up our cache with a Instance, so that we will see a delete for one instance
		fmt.Printf("Instance %s does not exist anymore\n", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a Instance was recreated with the same name
		_, ok := obj.(*v1alpha1.Instance)
		if !ok {
			c.log.Error("couldn't cast to Instance")
			return nil
		}

		c.log.Infof("running converge")
		c.wing.convergeInstance()
	}
	return nil
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *InstanceController) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.queue.Forget(key)
		return
	}

	// This controller retries 5 times if something goes wrong. After that, it stops trying.
	if c.queue.NumRequeues(key) < 5 {
		c.log.Infof("Error syncing instance %v: %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	// Report to an external entity that, even after several retries, we could not successfully process this key
	runtime.HandleError(err)
	c.log.Infof("Dropping instance %q out of the queue: %v", key, err)
}

func (c *InstanceController) Run(threadiness int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// Let the workers stop when we are done
	defer c.queue.ShutDown()
	c.log.Info("Starting Instance controller")

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
	c.log.Info("Stopping Instance controller")
}

func (c *InstanceController) runWorker() {
	for c.processNextItem() {
	}
}
