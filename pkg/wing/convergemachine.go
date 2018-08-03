package wing

import (
	"fmt"

	"github.com/jetstack/tarmak/pkg/apis/wing/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (w *Wing) convergeMachine() error {
	machineAPI := w.clientset.WingV1alpha1().Machines(w.flags.ClusterName)
	machine, err := machineAPI.Get(
		w.flags.InstanceName,
		metav1.GetOptions{},
	)
	if err != nil {
		if kerr, ok := err.(*apierrors.StatusError); ok && kerr.ErrStatus.Reason == metav1.StatusReasonNotFound {
			machine = &v1alpha1.Machine{
				ObjectMeta: metav1.ObjectMeta{
					Name: w.flags.InstanceName,
				},
				Status: v1alpha1.MachineStatus{
					Converged: false,
				},
			}
			_, err := machineAPI.Create(machine)
			if err != nil {
				return fmt.Errorf("error creating machine: %s", err)
			}
			return nil
		}
		return fmt.Errorf("error get existing machine: %s", err)
	}

	if machine.Status.Converged {
		w.log.Infof("Machine already converged: %s", machine.Name)
		return nil
	}

	puppetTarget := machine.Spec.PuppetTargetRef
	if puppetTarget == "" {
		w.log.Warn("no puppet target for machine: ", machine.Name)
		return nil
	}

	// FIXME: this shouldn't be done on the wing agent
	jobName := fmt.Sprintf("%s-%s", w.flags.InstanceName, puppetTarget)
	jobsAPI := w.clientset.WingV1alpha1().WingJobs(w.flags.ClusterName)
	job, err := jobsAPI.Get(
		jobName,
		metav1.GetOptions{},
	)
	if err != nil {
		if kerr, ok := err.(*apierrors.StatusError); ok && kerr.ErrStatus.Reason == metav1.StatusReasonNotFound {
			job = &v1alpha1.WingJob{
				ObjectMeta: metav1.ObjectMeta{
					Name: jobName,
				},
				Spec: &v1alpha1.WingJobSpec{
					InstanceName:     machine.Name,
					PuppetTargetRef:  puppetTarget,
					Operation:        "apply",
					RequestTimestamp: metav1.Now(),
				},
				Status: &v1alpha1.WingJobStatus{},
			}
			_, err := jobsAPI.Create(job)
			if err != nil {
				return fmt.Errorf("error creating WingJob: %s", err)
			}
			return nil
		}
		return fmt.Errorf("error get existing WingJob: %s", err)
	}

	machineCopy := machine.DeepCopy()
	machineCopy.Status.Converged = true
	_, err = machineAPI.Update(machineCopy)
	if err != nil {
		return err
	}

	return nil
}
