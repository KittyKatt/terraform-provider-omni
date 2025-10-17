resource "omni_cluster_machine_set_template" "example" {
  name = "my-cluster-workers"
  kind = "worker"
  labels = {
    "beep" = "boop"
  }
  annotations = {
    "example.io/example" = "beep-boop"
  }
  machines = [
    "00000000-0000-0000-0000-000000000000",
    "11111111-1111-1111-1111-111111111111",
  ]
  patches = [
    {
      id_override = "my-cluster-workers-configpatch.yaml"
      annotations = {
        "name" = "my-cluster-workers"
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
    "siderolabs/util-linux-tools"
  ]
}
