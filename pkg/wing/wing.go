// Copyright Jetstack Ltd. See LICENSE for details.
package wing

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/fields"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/jetstack/tarmak/pkg/apis/wing/common"
	"github.com/jetstack/tarmak/pkg/apis/wing/v1alpha1"
	client "github.com/jetstack/tarmak/pkg/wing/client/clientset/versioned"
	"github.com/jetstack/tarmak/pkg/wing/controller/machine"
)

const (
	DefaultMachineName = "$(hostname)"
)

type Wing struct {
	log       *logrus.Entry
	flags     *common.Flags
	clientset *client.Clientset

	// stop channel, signals termination to all goroutines
	stopCh chan struct{}

	convergeStopCh chan struct{}  // stop channel, signals to cancel current puppet run
	convergeWG     sync.WaitGroup // wait group for converge runs

	// controller loop
	controller *machine.Controller

	// allows overriding puppet command for testing
	puppetCommandOverride Command
}

func New(flags *common.Flags) *Wing {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel

	t := &Wing{
		log:            logger.WithField("app", "wing"),
		flags:          flags,
		stopCh:         make(chan struct{}),
		convergeStopCh: make(chan struct{}),
	}
	return t
}

func (w *Wing) Run(args []string) error {
	var errors []error

	if w.flags.MachineName == DefaultMachineName {
		machineName, err := os.Hostname()
		if err != nil {
			return err
		}

		w.flags.MachineName = machineName
	}

	if w.flags.MachineName == "" {
		errors = append(errors, fmt.Errorf("--machine-name flag cannot be empty"))
	}
	if w.flags.ManifestURL == "" {
		errors = append(errors, fmt.Errorf("--manifest-url flag cannot be empty"))
	}
	if err := utilerrors.NewAggregate(errors); err != nil {
		return err
	}

	// create connection to wing server
	restConfig := &rest.Config{
		Host: w.flags.ServerURL,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}
	clientset, err := client.NewForConfig(restConfig)
	if err != nil {
		return err
	}
	w.clientset = clientset

	// listen to signals
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	w.signalHandler(signalCh)

	// run converge loop after first start
	go w.Converge()

	// start watching for API server events that trigger applies
	w.watchForNotifications()

	// Wait for all goroutines to exit
	<-w.stopCh
	w.convergeWG.Wait()

	return err
}

func (w *Wing) Must(err error) *Wing {
	if err != nil {
		w.log.Fatal(err)
	}
	return w
}

func (w *Wing) watchForNotifications() {

	// create the machine watcher
	machineListWatcher := cache.NewListWatchFromClient(w.clientset.WingV1alpha1().RESTClient(), "machines", w.flags.ClusterName, fields.ParseSelectorOrDie(fmt.Sprintf("metadata.name=%s", w.flags.MachineName)))

	// create the workqueue
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	// Bind the workqueue to a cache with the help of an informer. This way we make sure that
	// whenever the cache is updated, the pod key is added to the workqueue.
	// Note that when we finally process the item from the workqueue, we might see a newer version
	// of the Pod than the version which was responsible for triggering the update.
	indexer, informer := cache.NewIndexerInformer(machineListWatcher, &v1alpha1.Machine{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.AddAfter(key, 1*time.Second)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.AddAfter(key, 1*time.Second)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer uses a delta queue, therefore for deletes we have to use this
			// key function.
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.AddAfter(key, 1*time.Second)
			}
		},
	}, cache.Indexers{})

	w.controller = machine.NewController(queue, indexer, informer, w)

	// Now let's start the controller
	go w.controller.Run(1, w.stopCh)

}

func (w *Wing) signalHandler(ch chan os.Signal) {
	go func() {
		select {
		case <-w.stopCh:
			break
		case sig := <-ch:
			switch sig {
			case syscall.SIGHUP:
				w.log.Infof("wing received SIGHUP")

				// if the puppet process is still running, kill and wait before re-converging
				w.log.Infof("terminating puppet if existing")
				close(w.convergeStopCh)
				w.convergeWG.Wait()

				// create new converge stop channel and run converge
				w.convergeStopCh = make(chan struct{})
				w.Converge()

			case syscall.SIGINT:
				w.log.Infof("wing received SIGINT")
				close(w.convergeStopCh)
				close(w.stopCh)

			case syscall.SIGTERM:
				w.log.Infof("wing received SIGTERM")
				close(w.convergeStopCh)
				close(w.stopCh)
			}
		}
	}()
}

func (w *Wing) Log() *logrus.Entry {
	return w.log
}

func (w *Wing) ConvergeWGWait() {
	w.convergeWG.Wait()
}

func (w *Wing) Clientset() *client.Clientset {
	return w.clientset
}

func (w *Wing) Flags() common.Flags {
	if w.flags != nil {
		return *w.flags
	}
	return common.Flags{}
}
