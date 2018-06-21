// Copyright Jetstack Ltd. See LICENSE for details.
package amazon

import (
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/mocks"
)

type fakeAmazon struct {
	*Amazon
	ctrl *gomock.Controller

	fakeEC2         *mocks.MockEC2
	fakeEnvironment *mocks.MockEnvironment
	fakeCluster     *mocks.MockCluster
	fakeTarmak      *mocks.MockTarmak
	fakeConfig      *mocks.MockConfig
}

func newFakeAmazon(t *testing.T) *fakeAmazon {

	f := &fakeAmazon{
		ctrl: gomock.NewController(t),
		Amazon: &Amazon{
			conf: &tarmakv1alpha1.Provider{
				Amazon: &tarmakv1alpha1.ProviderAmazon{
					KeyName: "myfake_key",
				},
			},
			log: logrus.WithField("test", true),
		},
	}
	f.fakeEC2 = mocks.NewMockEC2(f.ctrl)
	f.fakeEnvironment = mocks.NewMockEnvironment(f.ctrl)
	f.fakeCluster = mocks.NewMockCluster(f.ctrl)
	f.fakeTarmak = mocks.NewMockTarmak(f.ctrl)
	f.Amazon.ec2 = f.fakeEC2
	f.Amazon.tarmak = f.fakeTarmak
	f.fakeConfig = mocks.NewMockConfig(f.ctrl)
	f.fakeTarmak.EXPECT().Cluster().AnyTimes().Return(f.fakeCluster)
	f.fakeTarmak.EXPECT().Environment().AnyTimes().Return(f.fakeEnvironment)
	f.fakeCluster.EXPECT().Environment().AnyTimes().Return(f.fakeEnvironment)
	f.fakeTarmak.EXPECT().Config().AnyTimes().Return(f.fakeConfig)

	return f
}

func TestAmazon_validateAvailabilityZonesNoneGiven(t *testing.T) {
	a := newFakeAmazon(t)
	defer a.ctrl.Finish()

	a.fakeCluster.EXPECT().Subnets().Return([]clusterv1alpha1.Subnet{}).MinTimes(1)
	a.fakeCluster.EXPECT().Region().Return("london-north-1").AnyTimes()

	a.fakeEC2.EXPECT().DescribeAvailabilityZones(gomock.Any()).Return(&ec2.DescribeAvailabilityZonesOutput{
		AvailabilityZones: []*ec2.AvailabilityZone{
			&ec2.AvailabilityZone{
				ZoneName:   aws.String("london-north-1a"),
				State:      aws.String("available"),
				RegionName: aws.String("london-north-1"),
			},
			&ec2.AvailabilityZone{
				ZoneName:   aws.String("london-north-1b"),
				State:      aws.String("available"),
				RegionName: aws.String("london-north-1"),
			},
			&ec2.AvailabilityZone{
				ZoneName:   aws.String("london-north-1c"),
				State:      aws.String("available"),
				RegionName: aws.String("london-north-1"),
			},
		},
	}, nil)

	err := a.validateAvailabilityZones()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if act, exp := a.AvailabilityZones(), []string{"london-north-1a"}; !reflect.DeepEqual(act, exp) {
		t.Errorf("unexpected availability zones: act=%+v exp=%+v", act, exp)
	}
}

func TestAmazon_validateAvailabilityZonesCorrectGiven(t *testing.T) {
	a := newFakeAmazon(t)
	defer a.ctrl.Finish()

	a.fakeCluster.EXPECT().Subnets().Return([]clusterv1alpha1.Subnet{
		clusterv1alpha1.Subnet{
			Zone: "london-north-1b",
		},
		clusterv1alpha1.Subnet{
			Zone: "london-north-1c",
		},
	}).MinTimes(1)
	a.fakeCluster.EXPECT().Region().Return("london-north-1").AnyTimes()

	a.fakeEC2.EXPECT().DescribeAvailabilityZones(gomock.Any()).Return(&ec2.DescribeAvailabilityZonesOutput{
		AvailabilityZones: []*ec2.AvailabilityZone{
			&ec2.AvailabilityZone{
				ZoneName:   aws.String("london-north-1a"),
				State:      aws.String("available"),
				RegionName: aws.String("london-north-1"),
			},
			&ec2.AvailabilityZone{
				ZoneName:   aws.String("london-north-1b"),
				State:      aws.String("available"),
				RegionName: aws.String("london-north-1"),
			},
			&ec2.AvailabilityZone{
				ZoneName:   aws.String("london-north-1c"),
				State:      aws.String("available"),
				RegionName: aws.String("london-north-1"),
			},
		},
	}, nil)

	err := a.validateAvailabilityZones()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if act, exp := a.AvailabilityZones(), []string{"london-north-1b", "london-north-1c"}; !reflect.DeepEqual(act, exp) {
		t.Errorf("unexpected availability zones: act=%+v exp=%+v", act, exp)
	}
}

func TestAmazon_validateAvailabilityZonesFalseGiven(t *testing.T) {
	a := newFakeAmazon(t)
	defer a.ctrl.Finish()

	a.fakeCluster.EXPECT().Subnets().Return([]clusterv1alpha1.Subnet{
		clusterv1alpha1.Subnet{
			Zone: "london-north-1a",
		},
		clusterv1alpha1.Subnet{
			Zone: "london-north-1d",
		},
		clusterv1alpha1.Subnet{
			Zone: "london-north-1e",
		},
	}).MinTimes(1)
	a.fakeCluster.EXPECT().Region().Return("london-north-1").AnyTimes()
	a.fakeEnvironment.EXPECT().Location().Return("london-north-1").AnyTimes()

	a.fakeEC2.EXPECT().DescribeAvailabilityZones(gomock.Any()).Return(&ec2.DescribeAvailabilityZonesOutput{
		AvailabilityZones: []*ec2.AvailabilityZone{
			&ec2.AvailabilityZone{
				ZoneName:   aws.String("london-north-1a"),
				State:      aws.String("available"),
				RegionName: aws.String("london-north-1"),
			},
			&ec2.AvailabilityZone{
				ZoneName:   aws.String("london-north-1b"),
				State:      aws.String("available"),
				RegionName: aws.String("london-north-1"),
			},
			&ec2.AvailabilityZone{
				ZoneName:   aws.String("london-north-1c"),
				State:      aws.String("available"),
				RegionName: aws.String("london-north-1"),
			},
		},
	}, nil)

	err := a.validateAvailabilityZones()
	if err == nil {
		t.Error("expected an error")
	} else if !strings.Contains(err.Error(), "specified invalid availability zone") {
		t.Errorf("unexpected error messge: %s", err)
	}
}

func TestAmazon_verifyInstanceTypeNoneGiven(t *testing.T) {
	a := newFakeAmazon(t)
	defer a.ctrl.Finish()

	svc, err := a.EC2()
	if err != nil {
		t.Errorf("unexpected err:%v", err)
	}

	responce := &ec2.DescribeReservedInstancesOfferingsOutput{
		ReservedInstancesOfferings: []*ec2.ReservedInstancesOffering{
			&ec2.ReservedInstancesOffering{
				AvailabilityZone: aws.String("test-east-1a"),
			},
			&ec2.ReservedInstancesOffering{
				AvailabilityZone: aws.String("test-east-1b"),
			},
			&ec2.ReservedInstancesOffering{
				AvailabilityZone: aws.String("test-east-1c"),
			},
		},
	}

	a.fakeEC2.EXPECT().DescribeReservedInstancesOfferings(gomock.Any()).Return(responce, nil)

	err = a.verifyInstanceType("atype", []string{}, svc)
	if err != nil {
		t.Errorf("unexpected err:%v", err)
	}
}

func TestAmazon_verifyInstanceTypeZonesGiven(t *testing.T) {
	a := newFakeAmazon(t)
	defer a.ctrl.Finish()

	svc, err := a.EC2()
	if err != nil {
		t.Errorf("unexpected err:%v", err)
	}

	responce := &ec2.DescribeReservedInstancesOfferingsOutput{
		ReservedInstancesOfferings: []*ec2.ReservedInstancesOffering{
			&ec2.ReservedInstancesOffering{
				AvailabilityZone: aws.String("test-east-1a"),
			},
			&ec2.ReservedInstancesOffering{
				AvailabilityZone: aws.String("test-east-1b"),
			},
			&ec2.ReservedInstancesOffering{
				AvailabilityZone: aws.String("test-east-1c"),
			},
		},
	}

	a.fakeEC2.EXPECT().DescribeReservedInstancesOfferings(gomock.Any()).Return(responce, nil)

	err = a.verifyInstanceType("atype", []string{"test-east-1a", "test-east-1b", "test-east-1c"}, svc)
	if err != nil {
		t.Errorf("unexpected err:%v", err)
	}
}

func TestAmazon_verifyInstanceTypeZonesGivenWrong(t *testing.T) {
	a := newFakeAmazon(t)
	defer a.ctrl.Finish()

	svc, err := a.EC2()
	if err != nil {
		t.Errorf("unexpected err:%v", err)
	}

	responce := &ec2.DescribeReservedInstancesOfferingsOutput{
		ReservedInstancesOfferings: []*ec2.ReservedInstancesOffering{
			&ec2.ReservedInstancesOffering{
				AvailabilityZone: aws.String("test-wrong-1a"),
			},
			&ec2.ReservedInstancesOffering{
				AvailabilityZone: aws.String("test-east-1b"),
			},
			&ec2.ReservedInstancesOffering{
				AvailabilityZone: aws.String("test-east-1c"),
			},
		},
	}

	a.fakeEC2.EXPECT().DescribeReservedInstancesOfferings(gomock.Any()).Return(responce, nil)

	err = a.verifyInstanceType("atype", []string{"test-east-1a", "test-east-1b", "test-east-1c"}, svc)
	if err == nil {
		t.Errorf("expected an error but got none.")
	}
}

func TestAmazon_verifyInstanceTypeZonesOneOne(t *testing.T) {
	a := newFakeAmazon(t)
	defer a.ctrl.Finish()

	svc, err := a.EC2()
	if err != nil {
		t.Errorf("unexpected err:%v", err)
	}

	responce := &ec2.DescribeReservedInstancesOfferingsOutput{
		ReservedInstancesOfferings: []*ec2.ReservedInstancesOffering{
			&ec2.ReservedInstancesOffering{
				AvailabilityZone: aws.String("test-east-1a"),
			},
		},
	}

	a.fakeEC2.EXPECT().DescribeReservedInstancesOfferings(gomock.Any()).Return(responce, nil)

	err = a.verifyInstanceType("atype", []string{"test-east-1a"}, svc)
	if err != nil {
		t.Errorf("unexpected err:%v", err)
	}
}

func TestAmazon_verifyInstanceTypeZonesOneMany(t *testing.T) {
	a := newFakeAmazon(t)
	defer a.ctrl.Finish()

	svc, err := a.EC2()
	if err != nil {
		t.Errorf("unexpected err:%v", err)
	}

	responce := &ec2.DescribeReservedInstancesOfferingsOutput{
		ReservedInstancesOfferings: []*ec2.ReservedInstancesOffering{
			&ec2.ReservedInstancesOffering{
				AvailabilityZone: aws.String("test-east-1a"),
			},
			&ec2.ReservedInstancesOffering{
				AvailabilityZone: aws.String("test-east-1b"),
			},
		},
	}

	a.fakeEC2.EXPECT().DescribeReservedInstancesOfferings(gomock.Any()).Return(responce, nil)

	err = a.verifyInstanceType("atype", []string{"test-east-1a"}, svc)
	if err != nil {
		t.Errorf("unexpected err:%v", err)
	}
}

func TestAmazon_verifyInstanceTypeZonesNotAllAvailable(t *testing.T) {
	a := newFakeAmazon(t)
	defer a.ctrl.Finish()

	svc, err := a.EC2()
	if err != nil {
		t.Errorf("unexpected err:%v", err)
	}

	responce := &ec2.DescribeReservedInstancesOfferingsOutput{
		ReservedInstancesOfferings: []*ec2.ReservedInstancesOffering{
			&ec2.ReservedInstancesOffering{
				AvailabilityZone: aws.String("test-east-1a"),
			},
		},
	}

	a.fakeEC2.EXPECT().DescribeReservedInstancesOfferings(gomock.Any()).Return(responce, nil)

	err = a.verifyInstanceType("atype", []string{"test-east-1a", "test-east-1b"}, svc)
	if err == nil {
		t.Errorf("expected error but got none")
	}
}
