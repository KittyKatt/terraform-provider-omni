package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var machineTestControlPlaneMachines = []string{
	"000000000-0000-0000-0000-000000000000",
	"111111111-1111-1111-1111-111111111111",
	"222222222-2222-2222-2222-222222222222",
}

func TestAccMachineTemplateResource(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + testMachineTemplate(machineTestControlPlaneMachines),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("omni_cluster_machine_template.test1", "name", machineTestControlPlaneMachines[0]),
					resource.TestCheckResourceAttr("omni_cluster_machine_template.test2", "name", machineTestControlPlaneMachines[1]),
					resource.TestCheckResourceAttr("omni_cluster_machine_template.test3", "name", machineTestControlPlaneMachines[2]),
					resource.TestCheckResourceAttr("omni_cluster_machine_template.test1", "role", "controlplane"),
					resource.TestCheckResourceAttr("omni_cluster_machine_template.test1", "labels.beep", "boop"),
					resource.TestCheckResourceAttr("omni_cluster_machine_template.test2", "locked", "true"),
					resource.TestCheckResourceAttr("omni_cluster_machine_template.test2", "patches.#", "1"),
					resource.TestCheckResourceAttr("omni_cluster_machine_template.test3", "role", "worker"),
					// TODO: find out why things break with machine templates now
					// resource.TestCheckResourceAttr("omni_cluster_machine_template.test2", "patches[0].annotations.beep", "boop"),
					// resource.TestCheckResourceAttr("omni_cluster_machine_template.test3", "system_extensions[0]", "siderolabs/iscsi-tools"),
					resource.TestCheckResourceAttr("omni_cluster_machine_template.test3", "system_extensions.#", "1"),
					resource.TestCheckResourceAttrSet("omni_cluster_machine_template.test1", "id"),
					resource.TestCheckResourceAttrSet("omni_cluster_machine_template.test2", "id"),
					resource.TestCheckResourceAttrSet("omni_cluster_machine_template.test3", "id"),
					resource.TestCheckResourceAttrSet("omni_cluster_machine_template.test1", "last_updated"),
					resource.TestCheckResourceAttrSet("omni_cluster_machine_template.test2", "last_updated"),
					resource.TestCheckResourceAttrSet("omni_cluster_machine_template.test3", "last_updated"),
					resource.TestCheckResourceAttrSet("omni_cluster_machine_template.test1", "created_at"),
					resource.TestCheckResourceAttrSet("omni_cluster_machine_template.test2", "created_at"),
					resource.TestCheckResourceAttrSet("omni_cluster_machine_template.test3", "created_at"),
				),
			},
		},
	})
}

func testMachineTemplate(machines []string) string {
	return fmt.Sprintf(`
resource "omni_cluster_machine_template" "test1" {
  name = "%s"
  role = "controlplane"
  labels = {
    beep = "boop"
  }
}

resource "omni_cluster_machine_template" "test2" {
  name = "%s"
  role = "controlplane"
  locked = true
  patches = [
    {
      id_override = "test"
      annotations = {
        beep = "boop"
      }
      inline = <<EOT
      machine : {
        network : {
          hostname : lower("test")
        }
      EOT
    }
  ]
}

resource "omni_cluster_machine_template" "test3" {
  name = "%s"
  role = "worker"
  system_extensions = [
    "siderolabs/iscsi-tools"
  ]
}


`, machines[0], machines[1], machines[2])
}
