// Copyright Jetstack Ltd. See LICENSE for details.
package tarmak

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	clusterv1alpha1 "github.com/jetstack/tarmak/pkg/apis/cluster/v1alpha1"
)

func TestAccDataSourceTarmakBastionInstance(t *testing.T) {
	s := newRPCServer(t)

	s.fakeEnvironment.EXPECT().VerifyBastionAvailable().MinTimes(1).Return(nil)
	s.fakeCluster.EXPECT().GetState().MinTimes(1).Return("")

	s.Start()
	defer s.Finish()

	resource.Test(t, resource.TestCase{
		Providers:  testAccProviders,
		IsUnitTest: true,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccDataSourceTarmakBastionStatusBase, s.socketPath),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceTarmakBastionInstance("data.tarmak_bastion_instance.bastion"),
				),
			},
		},
	})
}

func TestAccDataSourceTarmakBastionInstanceDuringDestroy(t *testing.T) {
	s := newRPCServer(t)

	s.fakeCluster.EXPECT().GetState().AnyTimes().Return(clusterv1alpha1.StateDestroy)

	s.Start()
	defer s.Finish()

	resource.Test(t, resource.TestCase{
		Providers:  testAccProviders,
		IsUnitTest: true,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccDataSourceTarmakBastionStatusBase, s.socketPath),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceTarmakBastionInstance("data.tarmak_bastion_instance.bastion"),
				),
			},
		},
	})
}

func testAccDataSourceTarmakBastionInstance(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		attr := rs.Primary.Attributes

		if attr["hostname"] != "1.2.3.4" {
			return fmt.Errorf("bad hostname %s", attr["hostname"])
		}
		if attr["username"] != "centos" {
			return fmt.Errorf("bad username %s", attr["username"])
		}
		return nil
	}
}

const testAccDataSourceTarmakBastionStatusBase = `
provider "tarmak" {
  socket_path = "%s"
}

data "tarmak_bastion_instance" "bastion" {
  hostname = "1.2.3.4"
  username = "centos"
}
`
