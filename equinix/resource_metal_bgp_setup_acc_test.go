package equinix

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccMetalBGPSetup_basic(t *testing.T) {
	rs := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccMetalBGPSetupCheckDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccMetalBGPSetupConfig_basic(rs),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"equinix_metal_device.test", "id",
						"equinix_metal_bgp_session.test4", "device_id"),
					resource.TestCheckResourceAttrPair(
						"equinix_metal_device.test", "id",
						"equinix_metal_bgp_session.test6", "device_id"),
					resource.TestCheckResourceAttr(
						"equinix_metal_bgp_session.test4", "default_route", "true"),
					resource.TestCheckResourceAttr(
						"equinix_metal_bgp_session.test6", "default_route", "true"),
					resource.TestCheckResourceAttr(
						"equinix_metal_bgp_session.test4", "address_family", "ipv4"),
					resource.TestCheckResourceAttr(
						"equinix_metal_bgp_session.test6", "address_family", "ipv6"),
					// there will be 2 BGP neighbors, for IPv4 and IPv6
					resource.TestCheckResourceAttr(
						"data.equinix_metal_device_bgp_neighbors.test", "bgp_neighbors.#", "2"),
				),
			},
			{
				ResourceName:      "equinix_metal_bgp_session.test4",
				ImportState:       true,
				ImportStateVerify: true,
				// TODO(ocobleseqx) status returns "unknown" first and "down" after refresh. Should we add WaitForStateContext for "down"/"up"?
				ImportStateVerifyIgnore: []string{"status"},
			},
		},
	})
}

func testAccMetalBGPSetupCheckDestroyed(s *terraform.State) error {
	client := testAccProvider.Meta().(*Config).metal

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "equinix_metal_bgp_session" {
			continue
		}
		if _, _, err := client.BGPSessions.Get(rs.Primary.ID, nil); err == nil {
			return fmt.Errorf("Metal BGPSession still exists")
		}
	}

	return nil
}

func testAccMetalBGPSetupConfig_basic(name string) string {
	return fmt.Sprintf(`
%s

resource "equinix_metal_project" "test" {
    name = "tfacc-bgp_session-%s"
	bgp_config {
		deployment_type = "local"
		md5 = "C179c28c41a85b"
		asn = 65000
	}
}

resource "equinix_metal_device" "test" {
    hostname         = "tfacc-test-bgp-sesh"
    plan             = local.plan
    facilities       = local.facilities
    operating_system = local.os
    billing_cycle    = "hourly"
    project_id       = "${equinix_metal_project.test.id}"
    termination_time = "%s"

	lifecycle {
      ignore_changes = [
        plan,
        facilities,
      ]
    }
}

resource "equinix_metal_bgp_session" "test4" {
	device_id = "${equinix_metal_device.test.id}"
	address_family = "ipv4"
	default_route = true
}

resource "equinix_metal_bgp_session" "test6" {
	device_id = "${equinix_metal_device.test.id}"
	address_family = "ipv6"
	default_route = true
}

data "equinix_metal_device_bgp_neighbors" "test" {
	device_id  = equinix_metal_bgp_session.test4.device_id
}
`, confAccMetalDevice_base(preferable_plans, preferable_metros, preferable_os), name, testDeviceTerminationTime())
}
