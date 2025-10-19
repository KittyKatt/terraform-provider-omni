package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestClusterTemplateResource(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + testClusterTemplate("test", "1.33.4", "1.11.2"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.omni_cluster_template.test", tfjsonpath.New("name"), knownvalue.StringExact("test")),
					statecheck.ExpectKnownValue("data.omni_cluster_template.test", tfjsonpath.New("kubernetes").AtMapKey("version"), knownvalue.StringExact("1.33.4")),
					statecheck.ExpectKnownValue("data.omni_cluster_template.test", tfjsonpath.New("talos").AtMapKey("version"), knownvalue.StringExact("1.11.2")),
					statecheck.ExpectKnownValue("data.omni_cluster_template.test", tfjsonpath.New("labels").AtMapKey("beep"), knownvalue.StringExact("boop")),
					statecheck.ExpectKnownValue("data.omni_cluster_template.test", tfjsonpath.New("annotations").AtMapKey("example"), knownvalue.StringExact("test")),
					statecheck.ExpectKnownValue("data.omni_cluster_template.test", tfjsonpath.New("features").AtMapKey("disk_encryption"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue("data.omni_cluster_template.test", tfjsonpath.New("features").AtMapKey("enable_workload_proxy"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue("data.omni_cluster_template.test", tfjsonpath.New("features").AtMapKey("backup_configuration").AtMapKey("interval"), knownvalue.StringExact("24h")),
					statecheck.ExpectKnownValue("data.omni_cluster_template.test", tfjsonpath.New("patches").AtSliceIndex(0).AtMapKey("annotations").AtMapKey("name"), knownvalue.StringExact("test")),
					statecheck.ExpectKnownValue("data.omni_cluster_template.test", tfjsonpath.New("patches"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue("data.omni_cluster_template.test", tfjsonpath.New("system_extensions"), knownvalue.ListSizeExact(0)),
				},
			},
		},
	})
}

func testClusterTemplate(clusterName string, kubernetesVersion string, talosVersion string) string {
	return fmt.Sprintf(`
data "omni_cluster_template" "test" {
  name = "%s"
  kubernetes = {
    version = "%s"
  }
  talos = {
    version = "%s"
  }
  labels = {
    beep = "boop"
  }
  annotations = {
    boop = "beep"
    example = "test"
  }
  features = {
    disk_encryption = true
    enable_workload_proxy = false
    backup_configuration = {
      interval = "24h"
    }
  }
  patches = [
    {
      id_override = "test"
      annotations = {
        name = "test"
      }
      inline = <<EOT
      cluster:
        network:
          cni:
            name: none
      EOT
    }
  ]
  system_extensions = []
}
`, clusterName, kubernetesVersion, talosVersion)
}
