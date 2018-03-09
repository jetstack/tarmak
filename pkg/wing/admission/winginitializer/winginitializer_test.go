// Copyright Jetstack Ltd. See LICENSE for details.
package winginitializer_test

import (
	"testing"
	"time"

	"github.com/jetstack/tarmak/pkg/wing/admission/winginitializer"
	"github.com/jetstack/tarmak/pkg/wing/clients/internalclientset/fake"
	informers "github.com/jetstack/tarmak/pkg/wing/informers/internalversion"
	"k8s.io/apiserver/pkg/admission"
)

// TestWantsInternalWingInformerFactory ensures that the informer factory is injected
// when the WantsInternalWingInformerFactory interface is implemented by a plugin.
func TestWantsInternalWingInformerFactory(t *testing.T) {
	cs := &fake.Clientset{}
	sf := informers.NewSharedInformerFactory(cs, time.Duration(1)*time.Second)
	target, err := winginitializer.New(sf)
	if err != nil {
		t.Fatalf("expected to create an instance of initializer but got an error = %s", err.Error())
	}
	wantWingInformerFactory := &wantInternalWingInformerFactory{}
	target.Initialize(wantWingInformerFactory)
	if wantWingInformerFactory.sf != sf {
		t.Errorf("expected informer factory to be initialized")
	}
}

// wantInternalWingInformerFactory is a test stub that fulfills the WantsInternalWingInformerFactory interface
type wantInternalWingInformerFactory struct {
	sf informers.SharedInformerFactory
}

func (self *wantInternalWingInformerFactory) SetInternalWingInformerFactory(sf informers.SharedInformerFactory) {
	self.sf = sf
}
func (self *wantInternalWingInformerFactory) Admit(a admission.Attributes) error { return nil }
func (self *wantInternalWingInformerFactory) Handles(o admission.Operation) bool { return false }
func (self *wantInternalWingInformerFactory) ValidateInitialization() error      { return nil }

var _ admission.Interface = &wantInternalWingInformerFactory{}
var _ winginitializer.WantsInternalWingInformerFactory = &wantInternalWingInformerFactory{}
