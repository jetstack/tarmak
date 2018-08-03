// Copyright Jetstack Ltd. See LICENSE for details.
package instaceinittime

import (
	"errors"
	"io"
	"time"

	"k8s.io/apiserver/pkg/admission"

	"github.com/jetstack/tarmak/pkg/apis/wing"
)

const PluginName = "InstanceInitTime"

// Register registers a plugin
func Register(plugins *admission.Plugins) {
	plugins.Register(PluginName, func(config io.Reader) (admission.Interface, error) {
		return New()
	})
}

type instanceInitTime struct {
	*admission.Handler
}

// Admit ensures that the object in-flight is of kind Flunder.
// In addition checks that the Name is not on the banned list.
// The list is stored in Fischers API objects.
func (d *instanceInitTime) Admit(a admission.Attributes) error {
	// we are only interested in WingJobs
	if a.GetKind().GroupKind() != wing.Kind("WingJob") {
		return nil
	}

	job, ok := a.GetObject().(*wing.WingJob)
	if !ok {
		return errors.New("unexpected object time")
	}

	if job.Spec != nil {
		if job.Spec.RequestTimestamp.IsZero() {
			job.Spec.RequestTimestamp.Time = time.Now()
		}
	}

	return nil
}

// Validate checks whether the plugin was correctly initialized.
func (d *instanceInitTime) Validate() error {
	return nil
}

// New creates a new instances init time admission plugin
func New() (admission.Interface, error) {
	return &instanceInitTime{
		Handler: admission.NewHandler(admission.Create, admission.Update),
	}, nil
}
