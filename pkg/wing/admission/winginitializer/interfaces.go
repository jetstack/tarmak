// Copyright Jetstack Ltd. See LICENSE for details.

package winginitializer

import (
	"k8s.io/apiserver/pkg/admission"

	informers "github.com/jetstack/tarmak/pkg/wing/client/informers/externalversions"
)

// WantsInternalWingInformerFactory defines a function which sets InformerFactory for admission plugins that need it
type WantsInternalWingInformerFactory interface {
	SetInternalWingInformerFactory(informers.SharedInformerFactory)
	admission.InitializationValidator
}
