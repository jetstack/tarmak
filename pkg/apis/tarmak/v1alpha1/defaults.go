// Copyright Jetstack Ltd. See LICENSE for details.
package v1alpha1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var zeroTime metav1.Time

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

func SetDefaults_Environment(obj *Environment) {
	// set creation time, if unset
	if obj.CreationTimestamp == zeroTime {
		obj.CreationTimestamp.Time = time.Now()
	}

	// set private zone if not existing
	if obj.PrivateZone == "" {
		obj.PrivateZone = "tarmak.local"
	}
}

func SetDefaults_Config(obj *Config) {
	// set creation time, if unset
	if obj.CreationTimestamp == zeroTime {
		obj.CreationTimestamp.Time = time.Now()
	}
}

func SetDefaults_Provider(obj *Provider) {
	// set creation time, if unset
	if obj.CreationTimestamp == zeroTime {
		obj.CreationTimestamp.Time = time.Now()
	}
}
