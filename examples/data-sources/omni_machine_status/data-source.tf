# Copyright (c) HashiCorp, Inc.

# Example shown filters to machiens that currently don't belong to a cluster
data "omni_machine_status" "example_1" {
  filters = {
    cluster = ""
  }
}

# Example show filters to machiens that have eight cpu cores and belong
# to no cluster
data "omni_machine_status" "example_2" {
  filters = {
    cluster = ""
    labels = {
      "omni.sidero.dev/cores" = "8"
    }
  }
}

# Example shown filters to machines that belong to the cluster "my-cluster",
# are connected currently, and are a control plane node
data "omni_machine_status" "example_3" {
  filters = {
    cluster   = "my-cluster"
    connected = true
    role      = "controlplane"
  }
}

# Example shown filters to machines that belong to the cluster "my-cluster",
# are connected currently, and are a worker node
data "omni_machine_status" "example_4" {
  filters = {
    cluster   = "my-cluster"
    connected = true
    role      = "worker"
  }
}

# Example shown filters to machines that are currently connected and have the
# image labels "homelab:my-homelab" and "beep:boop"
data "omni_machine_status" "example_5" {
  filters = {
    connected = true
    image_labels = [
      "homelab:my-homelab",
      "beep:boop"
    ]
  }
}

# Output below shows outputting the machine-id of all machines identified in
# example_5
output "omni_machines_example_5" {
  value = data.omni_machine_status.example_5.machines[*].id
}

