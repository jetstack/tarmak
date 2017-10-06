// Copyright Jetstack Ltd. See LICENSE for details.
package instance_pool

import (
	"testing"

	"github.com/golang/mock/gomock"
	"k8s.io/apimachinery/pkg/api/resource"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/mocks"
)

type fakeVolume struct {
	*Volume
	ctrl *gomock.Controller

	conf *clusterv1alpha1.Volume

	pos int

	fakeProvider *mocks.MockProvider
}

func newFakeVolume(t *testing.T) *fakeVolume {
	v := &fakeVolume{
		conf: &clusterv1alpha1.Volume{
			Size: resource.NewQuantity(5*1024*1024*1024, resource.BinarySI),
			Type: clusterv1alpha1.VolumeTypeSSD,
		},
		ctrl: gomock.NewController(t),
	}
	v.conf.Name = "root"
	v.fakeProvider = mocks.NewMockProvider(v.ctrl)

	return v
}

func (v *fakeVolume) New() error {
	volume, err := NewVolumeFromConfig(v.pos, v.fakeProvider, v.conf)
	v.Volume = volume
	return err
}

func TestVolume_AWS_SSD(t *testing.T) {
	v := newFakeVolume(t)
	defer v.ctrl.Finish()

	v.fakeProvider.EXPECT().VolumeType("ssd").Return("gp2", nil)
	v.fakeProvider.EXPECT().Cloud().Return(clusterv1alpha1.CloudAmazon).AnyTimes()
	v.fakeProvider.EXPECT().Name().Return("aws1").AnyTimes()
	v.pos = 1

	err := v.New()
	if err != nil {
		t.Error("unexpected error: ", err)
	}

	if act, exp := v.Size(), 5; act != exp {
		t.Errorf("unexpected size, actual = %d, expected = %d", act, exp)
	}

	if act, exp := v.Name(), "root"; act != exp {
		t.Errorf("unexpected name, actual = '%s', expected = '%s'", act, exp)
	}

	if act, exp := v.Type(), "gp2"; act != exp {
		t.Errorf("unexpected type, actual = '%s', expected = '%s'", act, exp)
	}

	if act, exp := v.Device(), "/dev/sde"; act != exp {
		t.Errorf("unexpected device, actual = '%s', expected = '%s'", act, exp)
	}
}
