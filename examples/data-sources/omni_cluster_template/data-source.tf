# Copyright (c) HashiCorp, Inc.

data "omni_cluster_template" "example" {
  name = "my-cluster"
  kubernetes = {
    version = "1.34.3"
  }
  talos = {
    version = "1.11.1"
  }
  labels = {
    "beep" = "boop"
  }
  annotations = {
    "example.io/example" = "beep-boop"
  }
  features = {
    disk_encryption                = false
    enable_workload_proxy          = false
    use_embedded_discovery_service = false
    backup_configuration = {
      interval = "5m"
    }
  }
  patches = [
    {
      id_override = "my-cluster-configpatch.yaml"
      annotations = {
        name = "my-cluster-configpatch"
      }
      inline = <<EOT
cluster:
  network:
    cni:
      name: none
  proxy:
    disabled: true
  allowSchedulingOnControlPlanes: true
machine:
  install:
    diskSelector:
      size: '>15GB'
      EOT
    }
  ]
  system_extensions = [
    "siderolabs/qemu-guest-agent",
  ]
}
