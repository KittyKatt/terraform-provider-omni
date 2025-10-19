package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var machineSetTestcontrolPlaneMachines = []string{
	"000000000-0000-0000-0000-000000000000",
	"111111111-1111-1111-1111-111111111111",
	"222222222-2222-2222-2222-222222222222",
}

func TestAccMachineSetTemplateResource(t *testing.T) {
	rName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + testMachineSetTemplate(rName, "controlplane", strings.Join(machineSetTestcontrolPlaneMachines, ", ")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("omni_cluster_machine_set_template.test", "machines.#", "3"),
					resource.TestCheckResourceAttr("omni_cluster_machine_set_template.test", "patches.#", "1"),
					resource.TestCheckResourceAttr("omni_cluster_machine_set_template.test", "system_extensions.#", "1"),
					resource.TestCheckResourceAttrSet("omni_cluster_machine_set_template.test", "id"),
					resource.TestCheckResourceAttrSet("omni_cluster_machine_set_template.test", "last_updated"),
					resource.TestCheckResourceAttrSet("omni_cluster_machine_set_template.test", "created_at"),
				),
			},
		},
	})
}

func testMachineSetTemplate(name string, kind string, machines string) string {
	return fmt.Sprintf(`
resource "omni_cluster_machine_set_template" "test" {
  name = "%s"
  kind = "%s"
  machines = [%s]
  patches = [
    {
      id_override = "test"
      labels = {
        "beep" = "boop"
      }
      annotations = {
        "beep" = "boop"
      }
      inline = <<EOT
      machine:
      features:
        kubernetesTalosAPIAccess:
        allowedKubernetesNamespaces:
          - default
        allowedRoles:
          - os:etcd:backup
        enabled: true
      EOT
    }
  ]
  system_extensions = [
    "siderolabs/qemu-guest-agent"
  ]
}`, name, kind, machines)
}
