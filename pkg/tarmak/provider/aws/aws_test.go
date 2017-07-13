package aws

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/mock/gomock"

	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/mocks"
)

type fakeAWS struct {
	*AWS
	ctrl *gomock.Controller

	fakeEC2         *mocks.MockEC2
	fakeEnvironment *mocks.MockEnvironment
}

func newFakeAWS(t *testing.T) *fakeAWS {

	f := &fakeAWS{
		ctrl: gomock.NewController(t),
		AWS: &AWS{
			conf: &config.AWSConfig{
				KeyName: "myfake_key",
			},
			log: logrus.WithField("test", true),
		},
	}
	f.fakeEC2 = mocks.NewMockEC2(f.ctrl)
	f.fakeEnvironment = mocks.NewMockEnvironment(f.ctrl)
	f.AWS.ec2 = f.fakeEC2
	f.AWS.environment = f.fakeEnvironment

	return f
}

func TestAWS_validateAvailabilityZonesNoneGiven(t *testing.T) {
	a := newFakeAWS(t)
	defer a.ctrl.Finish()

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

	if exp, act := a.AvailabilityZones(), []string{"london-north-1a"}; !reflect.DeepEqual(act, exp) {
		t.Errorf("unexpected availability zones: act=%+v exp=%+v", act, exp)
	}
}

func TestAWS_validateAvailabilityZonesCorrectGiven(t *testing.T) {
	a := newFakeAWS(t)
	defer a.ctrl.Finish()

	a.AWS.conf.AvailabiltyZones = []string{"london-north-1b", "london-north-1c"}
	a.AWS.conf.Region = "london-north-1"

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

	if exp, act := a.AvailabilityZones(), []string{"london-north-1b", "london-north-1c"}; !reflect.DeepEqual(act, exp) {
		t.Errorf("unexpected availability zones: act=%+v exp=%+v", act, exp)
	}
}

func TestAWS_validateAvailabilityZonesFalseGiven(t *testing.T) {
	a := newFakeAWS(t)
	defer a.ctrl.Finish()

	a.AWS.conf.AvailabiltyZones = []string{"london-north-1-a", "london-north-1-d", "london-north-1-e"}
	a.AWS.conf.Region = "london-north-1"

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
		t.Errorf("expected error: %s", err)
	} else if !strings.Contains(err.Error(), "specified invalid availability zone") {
		t.Errorf("unexpected error messge: %s", err)
	}
}
