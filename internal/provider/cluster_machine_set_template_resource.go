package provider

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"terraform-provider-omni/internal/models"
	"terraform-provider-omni/internal/validators"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/siderolabs/omni/client/pkg/client"
	yamlv3 "gopkg.in/yaml.v3"
)

var (
	_ resource.Resource              = &omniClusterMachineSetTemplate{}
	_ resource.ResourceWithConfigure = &omniClusterMachineSetTemplate{}
)

type omniClusterMachineSetTemplate struct {
	omniClient *client.Client
	context    context.Context
}

type OmniClusterMachineSetTemplateModelV0 struct {
	ID               types.String            `tfsdk:"id"`
	CreatedAt        types.String            `tfsdk:"created_at"`
	LastUpdated      types.String            `tfsdk:"last_updated"`
	Name             types.String            `tfsdk:"name"`
	Kind             types.String            `tfsdk:"kind"`
	SystemExtensions models.SystemExtensions `tfsdk:"system_extensions"`
	Labels           models.Labels           `tfsdk:"labels"`
	Annotations      models.Annotations      `tfsdk:"annotations"`
	Machines         models.MachineIDList    `tfsdk:"machines"`
	Patches          []models.Patch          `tfsdk:"patches"`
	YAML             types.String            `tfsdk:"yaml"`
}

func NewOmniClusterMachineSetTemplateResource() resource.Resource {
	return &omniClusterMachineSetTemplate{}
}

func (r *omniClusterMachineSetTemplate) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_machine_set_template"
}

func (r *omniClusterMachineSetTemplate) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	omniClient, ok := req.ProviderData.(*client.Client)
	tflog.Debug(context.Background(), "Configuring omni machine status data source")
	if !ok {
		resp.Diagnostics.AddError(
			"failed to get image factory client",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.context = context.Background()
	r.omniClient = omniClient
}

func (r *omniClusterMachineSetTemplate) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				Optional:    true,
				Description: "Name (ID) of the machine. Only required on Worker machine sets.",
				Validators: []validator.String{
					validators.RequiredWhenValueIs(path.MatchRoot("kind"), types.StringValue("worker")),
				},
			},
			"kind": schema.StringAttribute{
				Required:    true,
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
			"machines": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "List of machines belonging to machine set.",
				Required:    true,
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

func (r *omniClusterMachineSetTemplate) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OmniClusterMachineSetTemplateModelV0

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	finalPlan, err := compileMachineSetTemplate(plan)
	if err != nil {
		resp.Diagnostics.AddError("Error encountered compiling machine set template.", fmt.Sprintf("Error: %s", err))
	}

	plan = finalPlan

	timeNow := types.StringValue(time.Now().Format(time.RFC850))
	plan.CreatedAt = timeNow
	plan.LastUpdated = timeNow

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *omniClusterMachineSetTemplate) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.omniClient == nil {
		resp.Diagnostics.AddError("omni client is not configured", "Please report this issue to the provider developers.")
		return
	}

	var config OmniClusterMachineSetTemplateModelV0
	resp.Diagnostics.Append(req.State.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Read cluster machine set template from state.")
	tflog.Trace(ctx, fmt.Sprintf("Cluster machine set YAML from state:\n%s", config.YAML))

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (r *omniClusterMachineSetTemplate) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OmniClusterMachineSetTemplateModelV0
	var state OmniClusterMachineSetTemplateModelV0

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	finalPlan, err := compileMachineSetTemplate(plan)
	if err != nil {
		resp.Diagnostics.AddError("Error encountered compiling machine set template.", fmt.Sprintf("Error: %s", err))
	}

	plan = finalPlan

	timeNow := types.StringValue(time.Now().Format(time.RFC850))
	plan.LastUpdated = timeNow

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *omniClusterMachineSetTemplate) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OmniClusterMachineSetTemplateModelV0
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func sanitizeMachineIDs(machineIDs []string) ([]string, error) {
	machineIDList := make([]string, 0, len(machineIDs))
	for _, machine := range machineIDs {
		sanitized := strings.TrimSpace(machine)
		sanitized = strings.ReplaceAll(sanitized, "\n", "")
		sanitized = strings.ReplaceAll(sanitized, "\r", "")
		if sanitized != "" {
			machineIDList = append(machineIDList, sanitized)
		} else {
			return nil, fmt.Errorf("Encountered an error sanitizing a machine ID of %s", machine)
		}
	}

	return machineIDList, nil
}

func compileMachineSetTemplate(plan OmniClusterMachineSetTemplateModelV0) (OmniClusterMachineSetTemplateModelV0, error) {
	var machineSetKind string
	if plan.Kind.ValueString() == "controlplane" {
		machineSetKind = KindControlPlane
	} else if plan.Kind.ValueString() == "worker" {
		machineSetKind = KindWorkers
	} else {
		// resp.Diagnostics.AddError("kind not determined", "Could not determine the Kind of the machine set.")
		return plan, fmt.Errorf("Could not determine the Kind of the machine set. Machine set kind: %s", plan.Kind.ValueString())
	}

	sanitizedMachines, err := sanitizeMachineIDs(plan.Machines)
	if err != nil {
		return plan, err
	}

	var buf bytes.Buffer
	enc := yamlv3.NewEncoder(&buf)
	enc.Encode(models.MachineSetYAML{
		Kind:             string(machineSetKind),
		SystemExtensions: plan.SystemExtensions,
		Name:             plan.Name.ValueString(),
		Labels:           plan.Labels,
		Annotations:      plan.Annotations,
		Machines:         sanitizedMachines,
		Patches:          convertPatchToYAML(plan.Patches),
	})
	yamlOutput := buf.String()

	plan.ID = plan.Name
	plan.YAML = types.StringValue(string(yamlOutput))

	return plan, nil
}
