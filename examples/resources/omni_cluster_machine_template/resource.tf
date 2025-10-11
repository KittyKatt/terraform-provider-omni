# Copyright (c) HashiCorp, Inc.

resource "omni_cluster_machine_template" "example" {
  name = "my-machine"
  role = "controlplane"
  labels = {
    "beep" = "boop"
  }
  annotations = {
    "example.io/example" = "beep-boop"
  }
  install = {
    disk = "/dev/sda"
  }
  locked = false
  patches = [
    {
      id_override = "my-machine-configpatch.yaml"
      annotations = {
        name = "my-machine-configpatch"
      }
      inline = <<EOT
        machine:
          network:
            hostname: "my-machine"
      EOT
    }
  ]
  system_extensions = [
    "siderolabs/util-linux-tools"
  ]
}
