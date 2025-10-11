package provider

import (
	"terraform-provider-omni/internal/models"

	"gopkg.in/yaml.v3"
)

const (
	KindCluster      = "Cluster"
	KindMachine      = "Machine"
	KindControlPlane = "ControlPlane"
	KindWorkers      = "Workers"
)

func convertClusterTalosOptionsToYAML(talosOptions models.ClusterTalos) models.ClusterTalosYAML {
	return models.ClusterTalosYAML{
		Version: talosOptions.Version.ValueString(),
	}
}

func convertClusterKubernetesOptionsToYAML(kubernetesOptions models.ClusterKubernetes) models.ClusterKubernetesYAML {
	return models.ClusterKubernetesYAML{
		Version: kubernetesOptions.Version.ValueString(),
	}
}

func convertClusterFeaturesToYAML(features models.ClusterFeatures) models.ClusterFeaturesYAML {
	return models.ClusterFeaturesYAML{
		DiskEncryption:              features.DiskEncryption.ValueBool(),
		EnableWorkloadProxy:         features.EnableWorkloadProxy.ValueBool(),
		UseEmbeddedDiscoveryService: features.UseEmbeddedDiscoveryService.ValueBool(),
		BackupConfiguration: models.ClusterBackupConfigurationYAML{
			Interval: features.BackupConfiguration.Interval.ValueString(),
		},
	}
}

func convertPatchToYAML(patches []models.Patch) []models.PatchYAML {
	var patchYAML []models.PatchYAML

	for _, patch := range patches {
		var inlineYAML map[string]any
		if patch.Inline != nil && *patch.Inline != "" {
			_ = yaml.Unmarshal([]byte(*patch.Inline), &inlineYAML)
		}
		patchYAML = append(patchYAML, models.PatchYAML{
			IDOverride:  patch.IDOverride,
			Labels:      patch.Labels,
			Annotations: patch.Annotations,
			File:        patch.File,
			Inline:      inlineYAML,
		})
	}

	return patchYAML
}
