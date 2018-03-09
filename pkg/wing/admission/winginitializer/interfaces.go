// Copyright Jetstack Ltd. See LICENSE for details.

package winginitializer

import (
	informers "github.com/jetstack/tarmak/pkg/wing/informers/internalversion"
	"k8s.io/apiserver/pkg/admission"
)

// WantsInternalWingInformerFactory defines a function which sets InformerFactory for admission plugins that need it
type WantsInternalWingInformerFactory interface {
	SetInternalWingInformerFactory(informers.SharedInformerFactory)
	admission.InitializationValidator
}
