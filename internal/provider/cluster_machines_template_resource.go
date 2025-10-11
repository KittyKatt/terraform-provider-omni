package provider

import (
	"bytes"
	"context"
	"fmt"
	"terraform-provider-omni/internal/models"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/siderolabs/omni/client/pkg/client"
	"gopkg.in/yaml.v3"
)

var (
	_ resource.Resource              = &omniClusterMachinesTemplate{}
	_ resource.ResourceWithConfigure = &omniClusterMachinesTemplate{}
)

type omniClusterMachinesTemplate struct {
	omniClient *client.Client
	context    context.Context
}

type ClusterMachinesTemplate struct {
	Kind             string                  `yaml:"kind"`
	SystemExtensions models.SystemExtensions `yaml:"systemExtensions,omitempty"`
	Name             string                  `yaml:"name"`
	Labels           map[string]string       `yaml:"labels,omitempty"`
	Annotations      map[string]string       `yaml:"annotations,omitempty"`
	Locked           bool                    `yaml:"locked,omitempty"`
	Install          *models.MachineInstall  `yaml:"install,omitempty"`
	Patches          []models.PatchYAML      `yaml:"patches,omitempty"`
}

type OmniClusterMachinesTemplateModelV0 struct {
	ID               types.String            `tfsdk:"id"`
	CreatedAt        types.String            `tfsdk:"created_at"`
	LastUpdated      types.String            `tfsdk:"last_updated"`
	Kind             types.String            `tfsdk:"kind"`
	SystemExtensions models.SystemExtensions `tfsdk:"system_extensions"`
	Name             types.String            `tfsdk:"name"`
	Role             types.String            `tfsdk:"role"`
	Labels           models.Labels           `tfsdk:"labels"`
	Annotations      models.Annotations      `tfsdk:"annotations"`
	Locked           types.Bool              `tfsdk:"locked"`
	Install          *models.MachineInstall  `tfsdk:"install"`
	Patches          []models.Patch          `tfsdk:"patches"`
	YAML             types.String            `tfsdk:"yaml"`
}

func NewOmniClusterMachinesTemplateResource() resource.Resource {
	return &omniClusterMachinesTemplate{}
}

func (r *omniClusterMachinesTemplate) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_machine_template"
}

func (r *omniClusterMachinesTemplate) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	omniClient, ok := req.ProviderData.(*client.Client)
	tflog.Debug(context.Background(), "Configuring omni machine template resource")
	if !ok {
		resp.Diagnostics.AddError(
			"failed to get image factory client",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.omniClient = omniClient
}

func (r *omniClusterMachinesTemplate) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Omni cluster machine set template definition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Name (ID) of cluster.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name (ID) of the machine.",
			},
			"kind": schema.StringAttribute{
				Computed:    true,
				Description: "Kind of Omni resource.",
			},
			"role": schema.StringAttribute{
				Required:    true,
				Description: "Role of the machine in the cluster.",
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
			"locked": schema.BoolAttribute{
				Optional:    true,
				Description: "Controls whether the machine is locked from configuration changes.",
			},
			"install": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"disk": schema.StringAttribute{
						Optional:    true,
						Description: "Disk the Talos system is installed to.",
					},
				},
				Optional:    true,
				Description: "Machine installation details.",
			},
			"patches": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
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
						"id_override": schema.StringAttribute{
							Optional:    true,
							Description: "The name (ID) of the patch.",
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

func (r *omniClusterMachinesTemplate) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OmniClusterMachinesTemplateModelV0

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.Encode(ClusterMachinesTemplate{
		Kind:             KindMachine,
		SystemExtensions: plan.SystemExtensions,
		Name:             plan.Name.ValueString(),
		Labels:           plan.Labels,
		Annotations:      plan.Annotations,
		Locked:           plan.Locked.ValueBool(),
		Install:          plan.Install,
		Patches:          convertPatchToYAML(plan.Patches),
	})
	yamlOutput := buf.String()

	tflog.Debug(ctx, fmt.Sprintf("machine template for %s:\n%s", plan.Name.ValueString(), yamlOutput))

	plan.ID = plan.Name
	plan.Kind = types.StringValue(KindMachine)
	plan.YAML = types.StringValue(string(yamlOutput))

	timeNow := types.StringValue(time.Now().Format(time.RFC850))
	plan.CreatedAt = timeNow
	plan.LastUpdated = timeNow

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *omniClusterMachinesTemplate) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.omniClient == nil {
		resp.Diagnostics.AddError("omni client is not configured", "Please report this issue to the provider developers.")
		return
	}

	var config OmniClusterMachinesTemplateModelV0
	resp.Diagnostics.Append(req.State.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Read cluster machine templates from state.")
	tflog.Trace(ctx, fmt.Sprintf("Cluster machines YAML from state:\n%s", config.YAML))

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (r *omniClusterMachinesTemplate) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OmniClusterMachinesTemplateModelV0
	var state OmniClusterMachinesTemplateModelV0

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	yamlOutput, err := yaml.Marshal(ClusterMachinesTemplate{
		Kind:             KindMachine,
		SystemExtensions: plan.SystemExtensions,
		Name:             plan.Name.ValueString(),
		Labels:           plan.Labels,
		Annotations:      plan.Annotations,
		Locked:           plan.Locked.ValueBool(),
		Install:          plan.Install,
		Patches:          convertPatchToYAML(plan.Patches),
	})
	if err != nil {
		resp.Diagnostics.AddError("Could not template YAML", "Error encountered generating YAML from inputs.")
		return
	}

	plan.Kind = types.StringValue(KindMachine)
	plan.ID = plan.Name
	plan.YAML = types.StringValue(string(yamlOutput))

	timeNow := types.StringValue(time.Now().Format(time.RFC850))
	plan.LastUpdated = timeNow

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *omniClusterMachinesTemplate) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OmniClusterMachinesTemplateModelV0
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
