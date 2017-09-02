package v1alpha1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

func SetDefaults_Cluster(obj *Cluster) {
	obj.CreationTimestamp = metav1.Time{time.Now()}
}

func SetDefaults_ServerPool(obj *ServerPool) {
	obj.CreationTimestamp = metav1.Time{time.Now()}
	if obj.Name == "" {
		obj.Name = obj.Type
	}
}
