// Copyright Jetstack Ltd. See LICENSE for details.
package machinedeployment

import (
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
	queue       workqueue.RateLimitingInterface
	depInformer cache.Controller
	setInformer cache.Controller
	depIndexer  cache.Indexer
	setIndexer  cache.Indexer
	log         *logrus.Entry
	client      *clientset.Clientset
}

func NewController(queue workqueue.RateLimitingInterface, depIndexer cache.Indexer,
	setIndexer cache.Indexer, depInformer cache.Controller, setInformer cache.Controller, client *clientset.Clientset) *Controller {
	return &Controller{
		depInformer: depInformer,
		setInformer: setInformer,
		depIndexer:  depIndexer,
		setIndexer:  setIndexer,
		queue:       queue,
		log:         logrus.NewEntry(logrus.New()).WithField("tier", "MachineDeployment-controller"),
		client:      client,
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
	err := c.sync(key.(string))
	// Handle the error if something went wrong during the execution of the business logic
	c.handleErr(err, key)
	return true
}

func (c *Controller) sync(key string) error {

	// see whether we have a deployment or set
	obj, exist, err := c.depIndexer.GetByKey(key)
	if err != nil || !exist {
		// we should have a set
		obj, exists, err := c.setIndexer.GetByKey(key)
		if err != nil {
			c.log.Errorf("Fetching object with key %s from store failed with %v", key, err)
			return err
		}

		if !exists {
			return nil
		}

		ms, ok := obj.(*v1alpha1.MachineSet)
		if !ok {
			return errors.New("failed to process next item, not a machinedeployment or machineset")
		}

		// get our deployment controlling this set
		md, err := c.getMachineDeployment(ms)
		if err != nil {
			return err
		}

		// sync deployment from the set
		return c.syncFromMachineSet(md, ms)
	}

	md, ok := obj.(*v1alpha1.MachineDeployment)
	if !ok {
		return errors.New("failed to process next item, not a machinedeployment")
	}

	var selectors []string
	for k, v := range md.Spec.Selector.MatchLabels {
		selectors = append(selectors, fmt.Sprintf("%s=%s", k, v))
	}

	// get controlled set from our deployment
	msAPI := c.client.WingV1alpha1().MachineSets(md.Namespace)
	ms, err := msAPI.Get(md.Name, metav1.GetOptions{})
	if err != nil {
		// the set doesn't exist so we need to create it
		if kerr, ok := err.(*apierrors.StatusError); ok && kerr.ErrStatus.Reason == metav1.StatusReasonNotFound {
			return c.createControlledMachineSet(md)
		} else {
			return fmt.Errorf("error get existing machineset: %s", err)
		}
	}

	// sync deployment from the set
	err = c.syncFromMachineSet(md, ms)
	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) createControlledMachineSet(md *v1alpha1.MachineDeployment) error {
	var minReplicas int32 = 1
	var maxReplicas int32 = 1
	if md.Spec.MinReplicas != nil {
		minReplicas = *md.Spec.MinReplicas
	}
	if md.Spec.MaxReplicas != nil {
		maxReplicas = *md.Spec.MaxReplicas
	}

	c.log.Infof("Creating MachineSet for MachineDeployment %s\n", md.Name)
	ms := &v1alpha1.MachineSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:   md.Name,
			Labels: utils.DuplicateMapString(md.Labels),
		},
		Spec: &v1alpha1.MachineSetSpec{
			MaxReplicas: &minReplicas,
			MinReplicas: &maxReplicas,
			Selector:    md.Spec.Selector,
		},
		Status: &v1alpha1.MachineSetStatus{},
	}

	msAPI := c.client.WingV1alpha1().MachineSets(md.Namespace)
	_, err := msAPI.Create(ms)
	if err != nil {
		return fmt.Errorf("error creating machineset: %s", err)
	}

	return nil
}

func (c *Controller) getMachineDeployment(ms *v1alpha1.MachineSet) (*v1alpha1.MachineDeployment, error) {
	mdAPI := c.client.WingV1alpha1().MachineDeployments(ms.Namespace)
	md, err := mdAPI.Get(ms.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return md, nil
}

func (c *Controller) syncFromMachineSet(md *v1alpha1.MachineDeployment, ms *v1alpha1.MachineSet) error {
	status := &v1alpha1.MachineDeploymentStatus{}
	if ms.Status != nil {
		status.Replicas = ms.Status.Replicas
		status.ObservedGeneration = ms.Status.ObservedGeneration + 1
		status.ReadyReplicas = ms.Status.ReadyReplicas
		// TODO: we may want to do something with available replicas
		status.AvailableReplicas = ms.Status.ReadyReplicas
		status.UnavailableReplicas = ms.Status.Replicas - ms.Status.ReadyReplicas
	}

	mdAPI := c.client.WingV1alpha1().MachineDeployments(md.Namespace)
	md.Status = status
	_, err := mdAPI.Update(md)
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

	go c.depInformer.Run(stopCh)
	go c.setInformer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.depInformer.HasSynced) || !cache.WaitForCacheSync(stopCh, c.setInformer.HasSynced) {
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
