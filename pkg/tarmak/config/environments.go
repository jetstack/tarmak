package config

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
)

func NewEnvironment(name, contact, project string) *tarmakv1alpha1.Environment {
	return &tarmakv1alpha1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			CreationTimestamp: metav1.Time{Time: time.Now()},
		},
		Contact: contact,
		Project: project,
	}
}
