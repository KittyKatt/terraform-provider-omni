package models

import (
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ClusterYAML struct {
	Kind             string                `tfsdk:"kind" yaml:"kind"`
	Name             string                `tfsdk:"name" yaml:"name"`
	Labels           map[string]string     `tfsdk:"labels" yaml:"labels,omitempty"`
	Annotations      map[string]string     `tfsdk:"annotations" yaml:"annotations,omitempty"`
	Kubernetes       ClusterKubernetesYAML `tfsdk:"kubernetes" yaml:"kubernetes"`
	Talos            ClusterTalosYAML      `tfsdk:"talos" yaml:"talos"`
	Features         ClusterFeaturesYAML   `tfsdk:"featuers" yaml:"features,omitempty"`
	Patches          []PatchYAML           `tfsdk:"patches" yaml:"patches,omitempty"`
	SystemExtensions []string              `tfsdk:"system_extensions" yaml:"systemExtensions,omitempty"`
}

type ClusterKubernetes struct {
	Version types.String `tfsdk:"version"`
}
type ClusterKubernetesYAML struct {
	Version string `tfsdk:"version" yaml:"version"`
}

type ClusterTalos struct {
	Version types.String `tfsdk:"version"`
}
type ClusterTalosYAML struct {
	Version string `tfsdk:"version" yaml:"version"`
}

type ClusterFeatures struct {
	DiskEncryption              types.Bool                 `tfsdk:"disk_encryption"`
	EnableWorkloadProxy         types.Bool                 `tfsdk:"enable_workload_proxy"`
	UseEmbeddedDiscoveryService types.Bool                 `tfsdk:"use_embedded_discovery_service"`
	BackupConfiguration         ClusterBackupConfiguration `tfsdk:"backup_configuration"`
}
type ClusterFeaturesYAML struct {
	DiskEncryption              bool                           `tfsdk:"disk_encryption" yaml:"diskEncryption,omitempty"`
	EnableWorkloadProxy         bool                           `tfsdk:"enable_workload_proxy" yaml:"enableWorkloadProxy,omitempty"`
	UseEmbeddedDiscoveryService bool                           `tfsdk:"use_embedded_discovery_service" yaml:"useEmbeddedDiscoveryService,omitempty"`
	BackupConfiguration         ClusterBackupConfigurationYAML `tfsdk:"backup_configuration" yaml:"backupConfiguration,omitempty"`
}

type ClusterBackupConfigurationYAML struct {
	Interval string `tfsdk:"interval" yaml:"interval"`
}
type ClusterBackupConfiguration struct {
	Interval timetypes.GoDuration `tfsdk:"interval"`
}
