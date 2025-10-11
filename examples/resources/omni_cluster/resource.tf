# Copyright (c) HashiCorp, Inc.

data "omni_cluster_template" "my_cluster" {
  name = var.cluster_name
  kubernetes {
    version = var.kubernetes_version
  }
  talos {
    version = var.talos_version
  }
}

resource "omni_cluster_machine_set_template" "controlplane" {
  kind     = "controlplane"
  machines = var.control_plane.machines
}

resource "omni_cluster_machine_set_template" "worker" {
  name     = var.workers.name
  kind     = "worker"
  machines = var.workers.machines
}

resource "omni_cluster_machine_template" "controlplanes" {
  for_each = var.control_plane.machines
  name     = each.key
  role     = "controlplane"
}

resource "omni_cluster_machine_template" "workers" {
  for_each = var.workers.machines
  name     = each.key
  role     = "worker"
}

resource "omni_cluster" "my_cluster" {
  cluster_template       = data.omni_cluster_template.my_cluster.yaml
  control_plane_template = omni_cluster_machine_set_template.controlplane.yaml
  workers_template = [
    omni_cluster_machine_set_template.workers.yaml
  ]
  machines_template = values(
    merge(
      { for machine in omni_cluster_machine_template.controlplane : machine.name => machine.yaml },
      { for machine in omni_cluster_machine_template.workers : machine.name => machine.yaml }
    )
  )
  delete_machine_links = true
}
