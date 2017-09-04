package v1alpha1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

func SetDefaults_Config(obj *Config) {
	obj.CreationTimestamp = metav1.Time{Time: time.Now()}
}

func SetDefaults_Provider(obj *Provider) {
	obj.CreationTimestamp = metav1.Time{Time: time.Now()}
}
