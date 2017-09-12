package aws

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/mock/gomock"

	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
	tarmakv1alpha1 "github.com/jetstack/tarmak/pkg/apis/tarmak/v1alpha1"
	"github.com/jetstack/tarmak/pkg/tarmak/mocks"
)

type fakeAWS struct {
	*AWS
	ctrl *gomock.Controller

	fakeEC2         *mocks.MockEC2
	fakeEnvironment *mocks.MockEnvironment
	fakeContext     *mocks.MockContext
	fakeTarmak      *mocks.MockTarmak
}

func newFakeAWS(t *testing.T) *fakeAWS {

	f := &fakeAWS{
		ctrl: gomock.NewController(t),
		AWS: &AWS{
			conf: &tarmakv1alpha1.Provider{
				AWS: &tarmakv1alpha1.ProviderAWS{
					KeyName: "myfake_key",
				},
			},
			log: logrus.WithField("test", true),
		},
	}
	f.fakeEC2 = mocks.NewMockEC2(f.ctrl)
	f.fakeEnvironment = mocks.NewMockEnvironment(f.ctrl)
	f.fakeContext = mocks.NewMockContext(f.ctrl)
	f.fakeTarmak = mocks.NewMockTarmak(f.ctrl)
	f.AWS.ec2 = f.fakeEC2
	f.AWS.tarmak = f.fakeTarmak
	f.fakeTarmak.EXPECT().Context().AnyTimes().Return(f.fakeContext)
	f.fakeTarmak.EXPECT().Environment().AnyTimes().Return(f.fakeEnvironment)
	f.fakeContext.EXPECT().Environment().AnyTimes().Return(f.fakeEnvironment)

	return f
}

func TestAWS_validateAvailabilityZonesNoneGiven(t *testing.T) {
	a := newFakeAWS(t)
	defer a.ctrl.Finish()

	a.fakeContext.EXPECT().Subnets().Return([]clusterv1alpha1.Subnet{}).MinTimes(1)
	a.fakeContext.EXPECT().Region().Return("london-north-1").AnyTimes()

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

func TestAWS_validateAvailabilityZonesCorrectGiven(t *testing.T) {
	a := newFakeAWS(t)
	defer a.ctrl.Finish()

	a.fakeContext.EXPECT().Subnets().Return([]clusterv1alpha1.Subnet{
		clusterv1alpha1.Subnet{
			Zone: "london-north-1b",
		},
		clusterv1alpha1.Subnet{
			Zone: "london-north-1c",
		},
	}).MinTimes(1)
	a.fakeContext.EXPECT().Region().Return("london-north-1").AnyTimes()

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

func TestAWS_validateAvailabilityZonesFalseGiven(t *testing.T) {
	a := newFakeAWS(t)
	defer a.ctrl.Finish()

	a.fakeContext.EXPECT().Subnets().Return([]clusterv1alpha1.Subnet{
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
	a.fakeContext.EXPECT().Region().Return("london-north-1").AnyTimes()
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
