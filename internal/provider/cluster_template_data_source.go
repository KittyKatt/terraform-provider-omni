package provider

import (
	"context"
	"terraform-provider-omni/internal/models"
	"terraform-provider-omni/internal/validators"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gopkg.in/yaml.v3"
)

var _ datasource.DataSource = &omniClusterTemplateDataSource{}

type omniClusterTemplateDataSource struct{}

type OmniClusterTemplateModelV0 struct {
	ID               types.String             `tfsdk:"id"`
	Kind             types.String             `tfsdk:"kind"`
	Name             types.String             `tfsdk:"name"`
	Labels           models.Labels            `tfsdk:"labels"`
	Annotations      models.Annotations       `tfsdk:"annotations"`
	Kubernetes       models.ClusterKubernetes `tfsdk:"kubernetes"`
	Talos            models.ClusterTalos      `tfsdk:"talos"`
	Features         models.ClusterFeatures   `tfsdk:"features"`
	Patches          []models.Patch           `tfsdk:"patches"`
	SystemExtensions models.SystemExtensions  `tfsdk:"system_extensions"`
	YAML             types.String             `tfsdk:"yaml"`
}

func NewOmniClusterTemplateDataSource() datasource.DataSource {
	return &omniClusterTemplateDataSource{}
}

func (d *omniClusterTemplateDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_template"
}

func (d *omniClusterTemplateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
}

func (d *omniClusterTemplateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates a machine template to use within a cluster template.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the resource.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Name (ID) of the machine. Only required on Worker machine sets.",
				Validators: []validator.String{
					validators.RequiredWhenValueIs(path.MatchRoot("kind"), types.StringValue("worker")),
				},
			},
			"kind": schema.StringAttribute{
				Computed:    true,
				Description: "Kind of Omni machine set.",
				Validators: []validator.String{
					stringvalidator.OneOf("controlplane", "worker"),
				},
			},
			"labels": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Labels to add to the machine.",
			},
			"annotations": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Labels to add to the machine.",
			},
			"kubernetes": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"version": schema.StringAttribute{
						Description: "Kubernetes verions of cluster.",
						Required:    true,
					},
				},
				Description: "Kubernetes options for cluster.",
				Required:    true,
			},
			"talos": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"version": schema.StringAttribute{
						Description: "Talos verions of cluster.",
						Required:    true,
					},
				},
				Description: "Talos options for cluster.",
				Required:    true,
			},
			"features": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"disk_encryption": schema.BoolAttribute{
						Description: "Setting to enable or disable disk encryption.",
						Optional:    true,
					},
					"enable_workload_proxy": schema.BoolAttribute{
						Description: "Setting to enable or disable workload proxy functionality.",
						Optional:    true,
					},
					"use_embedded_discovery_service": schema.BoolAttribute{
						Description: "Setting to enable or disable using the embedded discovery service.",
						Optional:    true,
					},
					"backup_configuration": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"interval": schema.StringAttribute{
								Optional:   true,
								CustomType: timetypes.GoDurationType{},
							},
						},
						Optional: true,
					},
				},
				Optional:    true,
				Description: "Setings to enable or disable different cluster features.",
			},
			"patches": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id_override": schema.StringAttribute{
							Optional:    true,
							Description: "The name (ID) of the patch.",
						},
						"labels": schema.MapAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Description: "The labels of the patch.",
						},
						"annotations": schema.MapAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Description: "The annotations of the patch.",
						},
						"file": schema.StringAttribute{
							Optional:    true,
							Description: "The file path to use as input for the patch.",
						},
						"inline": schema.StringAttribute{
							Optional:    true,
							Description: "The inline patch as YAML.",
						},
					},
				},
				Optional:    true,
				Description: "Machine-specific patches.",
			},
			"system_extensions": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of system extensions installed to machine.",
				Optional:    true,
			},
			"yaml": schema.StringAttribute{
				Description: "Rendered YAML cluster machine template.",
				Computed:    true,
			},
		},
	}
}

func (d *omniClusterTemplateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config OmniClusterTemplateModelV0

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	yamlOutput, err := yaml.Marshal(models.ClusterYAML{
		Kind:             string(KindCluster),
		Name:             config.Name.ValueString(),
		Labels:           config.Labels,
		Annotations:      config.Annotations,
		Kubernetes:       convertClusterKubernetesOptionsToYAML(config.Kubernetes),
		Talos:            convertClusterTalosOptionsToYAML(config.Talos),
		Features:         convertClusterFeaturesToYAML(config.Features),
		Patches:          convertPatchToYAML(config.Patches),
		SystemExtensions: config.SystemExtensions,
	})
	if err != nil {
		resp.Diagnostics.AddError("Could not template YAML", "Error encountered generating YAML from inputs.")
		return
	}

	config.ID = config.Name
	config.YAML = types.StringValue(string(yamlOutput))

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
