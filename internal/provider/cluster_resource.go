package provider

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"terraform-provider-omni/internal/models"
	"time"

	cosiresource "github.com/cosi-project/runtime/pkg/resource"
	"github.com/cosi-project/runtime/pkg/safe"
	"github.com/cosi-project/runtime/pkg/state"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/siderolabs/omni/client/pkg/client"
	"github.com/siderolabs/omni/client/pkg/omni/resources"
	"github.com/siderolabs/omni/client/pkg/omni/resources/omni"
	"github.com/siderolabs/omni/client/pkg/template"
	"github.com/siderolabs/omni/client/pkg/template/operations"
	"gopkg.in/yaml.v3"
)

var (
	_ resource.Resource = &omniClusterResource{}
	// _ resource.ResourceWithConfigure   = &omniClusterResource{}
	_ resource.ResourceWithImportState = &omniClusterResource{}
	// _ resource.ResourceWithModifyPlan  = &omniClusterResource{}
)

type omniClusterResource struct {
	omniClient *client.Client
	context    context.Context
}

type OmniClusterResourceModelV0 struct {
	ID                   types.String   `tfsdk:"id"`
	CreatedAt            types.String   `tfsdk:"created_at"`
	LastUpdated          types.String   `tfsdk:"last_updated"`
	DeleteMachineLinks   types.Bool     `tfsdk:"delete_machine_links"`
	ClusterTemplate      types.String   `tfsdk:"cluster_template"`
	ControlPlaneTemplate types.String   `tfsdk:"control_plane_template"`
	WorkersTemplate      []types.String `tfsdk:"workers_template"`
	MachinesTemplate     []types.String `tfsdk:"machines_template"`
	YAML                 types.String   `tfsdk:"yaml"`
}

func NewOmniClusterResource() resource.Resource {
	return &omniClusterResource{}
}

func (r *omniClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (r *omniClusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *omniClusterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Omni cluster template definition.",
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
			"delete_machine_links": schema.BoolAttribute{
				Optional:    true,
				Description: "Controls if machine links are deleted when cluster is deleted.",
				// Default:     booldefault.StaticBool(false),
			},
			"cluster_template": schema.StringAttribute{
				Description: "YAML document describing cluster.",
				Required:    true,
			},
			"control_plane_template": schema.StringAttribute{
				Description: "YAML document describing control plane machine set.",
				Required:    true,
			},
			"workers_template": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "YAML document describing workers machine sets.",
				Required:    true,
			},
			"machines_template": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "YAML document describing machines.",
				Required:    true,
			},
			"yaml": schema.StringAttribute{
				Description: "Full YAML document descripting cluster template.",
				Computed:    true,
			},
		},
	}
}

func (r *omniClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.omniClient == nil {
		resp.Diagnostics.AddError("omni client is not configured", "Please report this issue to the provider developers.")
		return
	}

	st := r.omniClient.Omni().State()

	var config OmniClusterResourceModelV0
	resp.Diagnostics.Append(req.State.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterName := config.ID

	buf := bytes.Buffer{}

	_, err := operations.ExportTemplate(ctx, st, clusterName.ValueString(), &buf)
	if err != nil {
		resp.Diagnostics.AddError("problem exporting tempalte", fmt.Sprintf("Encountered a problem exporting the cluster template for %s from Omni. Error: %s", clusterName.ValueString(), err))
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("exported yaml template on read:\n%s", buf.String()))

	config.YAML = types.StringValue(buf.String())

	cluster, controlPlane, workers, machines, err := SplitYAMLByKind(buf.String())
	if err != nil {
		// handle error
	}

	tflog.Debug(ctx, fmt.Sprintf("yaml split into kind (cluster):\n%s", cluster))
	tflog.Debug(ctx, fmt.Sprintf("yaml split into kind (controlplane):\n%s", controlPlane))
	tflog.Debug(ctx, fmt.Sprintf("yaml split into kind (workers):\n%s", workers))
	tflog.Debug(ctx, fmt.Sprintf("yaml split into kind (machines):\n%s", machines))

	config.ClusterTemplate = types.StringValue(cluster)
	config.ControlPlaneTemplate = types.StringValue(controlPlane)
	config.WorkersTemplate = workers
	config.MachinesTemplate = machines

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (r *omniClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OmniClusterResourceModelV0

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// set sane default for deleting machine links
	if plan.DeleteMachineLinks.IsUnknown() || plan.DeleteMachineLinks.IsNull() {
		plan.DeleteMachineLinks = types.BoolValue(false)
	}

	var clusterTemplateUnmarshaled models.ClusterYAML
	yaml.Unmarshal([]byte(plan.ClusterTemplate.ValueString()), &clusterTemplateUnmarshaled)

	plan.ID = types.StringValue(clusterTemplateUnmarshaled.Name)
	planYAML, err := constructYAMLTemplate(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("eror encountered constructing yaml", fmt.Sprintf("error: %s", err))
	}
	syncError := SyncClusterTemplateAndWaitForReady(ctx, r.omniClient.Omni().State(), strings.NewReader(planYAML))
	if syncError != nil {
		resp.Diagnostics.AddError("eror encountered syncing template", fmt.Sprintf("error: %s", syncError))
		return
	}

	plan.YAML = types.StringValue(planYAML)
	timeNow := types.StringValue(time.Now().Format(time.RFC850))
	plan.CreatedAt = timeNow
	plan.LastUpdated = timeNow

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *omniClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "you reached the delete function")
	st := r.omniClient.Omni().State()
	var state OmniClusterResourceModelV0
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "getting machine IDs now")

	var machineIDList []string
	for _, machineYAML := range state.MachinesTemplate {
		var machineSpec models.MachinesYAML
		// machineYAML.ValueString()
		yaml.Unmarshal([]byte(machineYAML.ValueString()), &machineSpec)
		machineIDList = append(machineIDList, machineSpec.Name)
	}

	tflog.Debug(ctx, fmt.Sprintf("machine IDs get:\n%s", machineIDList))

	machinesList, err := safe.StateList[*omni.MachineStatus](ctx, st, omni.NewMachineStatus(resources.DefaultNamespace, "").Metadata())
	if err != nil {
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("machines get:\n%s", machinesList))
	var clusterMachines []*omni.MachineStatus
	for _, machineID := range machineIDList {
		filteredMachine, found := machinesList.Find(func(e *omni.MachineStatus) bool {
			if e.Metadata().ID() != machineID {
				return false
			}
			return true
		})
		if found != false {
			clusterMachines = append(clusterMachines, filteredMachine)
			tflog.Debug(ctx, fmt.Sprintf("found machine:\n%s", filteredMachine.Metadata().ID()))
		}
	}

	clusterDeleteErr := operations.DeleteCluster(ctx, state.ID.ValueString(), io.Discard, st, operations.SyncOptions{})
	if clusterDeleteErr != nil {
		resp.Diagnostics.AddError("Error deleting cluster", fmt.Sprintf("error: %s", clusterDeleteErr))
		return
	}

	tflog.Debug(ctx, "sent deletion, monitoring status")

	for {
		_, err := safe.StateGetByID[*omni.ClusterDestroyStatus](ctx, st, omni.NewClusterDestroyStatus(resources.DefaultNamespace, "").Metadata().ID())
		if err != nil {
			break
		}
	}

	tflog.Debug(ctx, "it go deleted, checking if we want to delete machine links")

	if state.DeleteMachineLinks.ValueBool() {
		tflog.Debug(ctx, "turns out we do want to delete machine links")
		for _, machine := range clusterMachines {
			destroyReady, err := st.Teardown(
				ctx,
				cosiresource.NewMetadata(machine.Metadata().Namespace(),
					"Links.omni.sidero.dev",
					machine.Metadata().ID(),
					machine.Metadata().Version(),
				))
			if err != nil {
				resp.Diagnostics.AddError("Error during teardown", fmt.Sprintf("error: %s", err))
				return
			}

			if destroyReady {
				if err = st.Destroy(
					ctx,
					cosiresource.NewMetadata(machine.Metadata().Namespace(),
						"Links.omni.sidero.dev",
						machine.Metadata().ID(),
						machine.Metadata().Version(),
					)); err != nil {
					resp.Diagnostics.AddError("Error during destroy", fmt.Sprintf("error: %s", err))
					return
				}
			}
		}
	}
}

func (r *omniClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OmniClusterResourceModelV0
	var state OmniClusterResourceModelV0

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// set sane default for deleting machine links
	if plan.DeleteMachineLinks.IsUnknown() || plan.DeleteMachineLinks.IsNull() {
		plan.DeleteMachineLinks = types.BoolValue(false)
	}

	finalYAML, err := constructYAMLTemplate(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("error constructing YAML template", fmt.Sprintf("encountered an error constructing yaml template: %s", err))
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Constructed YAML during update:\n%s", finalYAML))

	syncError := SyncClusterTemplateAndWaitForReady(ctx, r.omniClient.Omni().State(), strings.NewReader(finalYAML))
	if syncError != nil {
		resp.Diagnostics.AddError("error syncing template", fmt.Sprintf("full error: %s", syncError))
	}

	plan.YAML = types.StringValue(finalYAML)
	tflog.Debug(ctx, fmt.Sprintf("plan YAML value after construct:\n%s", plan.YAML.ValueString()))

	var planClusterTemplateUnmarshaled models.ClusterYAML
	yaml.Unmarshal([]byte(plan.ClusterTemplate.ValueString()), &planClusterTemplateUnmarshaled)

	var stateClusterTemplateUnmarshaled models.ClusterYAML
	yaml.Unmarshal([]byte(state.ClusterTemplate.ValueString()), &stateClusterTemplateUnmarshaled)

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *omniClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func constructYAMLTemplate(ctx context.Context, plan OmniClusterResourceModelV0) (string, error) {
	buf := &bytes.Buffer{}

	encoder := yaml.NewEncoder(buf)
	encoder.SetIndent(2)

	encodeYAMLDoc := func(yamlStr types.String) error {
		if yamlStr.ValueString() == "" {
			return nil
		}
		var node yaml.Node
		if err := yaml.Unmarshal([]byte(*yamlStr.ValueStringPointer()), &node); err != nil {
			return err
		}
		return encoder.Encode(&node)
	}

	if err := encodeYAMLDoc(plan.ClusterTemplate); err != nil {
		return buf.String(), err
	}

	if err := encodeYAMLDoc(plan.ControlPlaneTemplate); err != nil {
		return buf.String(), err
	}

	for _, machineSet := range plan.WorkersTemplate {
		if err := encodeYAMLDoc(machineSet); err != nil {
			return buf.String(), err
		}
	}

	for _, machine := range plan.MachinesTemplate {
		if err := encodeYAMLDoc(machine); err != nil {
			return buf.String(), err
		}
	}

	encoder.Close()
	tflog.Debug(ctx, fmt.Sprintf("full constructed YAML:\n%s", buf.String()))

	return buf.String(), nil
}

func SplitYAMLByKind(yamlStr string) (clusterTemplateYAML string, controlPlaneTemplateYAML string, workersTemplateYAML []types.String, machinesTemplateYAML []types.String, err error) {
	dec := yaml.NewDecoder(bytes.NewReader([]byte(yamlStr)))
	buf := &bytes.Buffer{}
	for {
		var node yaml.Node
		if err := dec.Decode(&node); err != nil {
			if err == io.EOF {
				break
			}
			return "", "", nil, nil, err
		}

		// Find the "kind" field in the root mapping node
		var kind string
		if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
			root := node.Content[0]
			if root.Kind == yaml.MappingNode {
				for i := 0; i < len(root.Content)-1; i += 2 {
					k := root.Content[i]
					v := root.Content[i+1]
					if k.Value == "kind" {
						kind = v.Value
						break
					}
				}
			}
		}

		// Marshal the node back to YAML to preserve formatting
		enc := yaml.NewEncoder(buf)
		enc.SetIndent(0)
		if err := enc.Encode(&node); err != nil {
			return "", "", nil, nil, err
		}
		enc.Close()
		docStr := buf.String()

		switch kind {
		case "Cluster":
			clusterTemplateYAML = docStr
		case "ControlPlane":
			controlPlaneTemplateYAML = docStr
		case "Workers":
			workersTemplateYAML = append(workersTemplateYAML, types.StringValue(docStr))
		case "Machine":
			machinesTemplateYAML = append(machinesTemplateYAML, types.StringValue(docStr))
		}

		buf.Reset()
	}
	return
}

func ValidateClusterTemplate(input io.Reader) error {
	buf := &bytes.Buffer{}
	tee := io.TeeReader(input, buf)

	err := operations.ValidateTemplate(tee)
	if err != nil {
		return err
	}

	return nil
}

func SyncClusterTemplateAndWaitForReady(ctx context.Context, state state.State, input io.Reader) error {
	buf := &bytes.Buffer{}
	tee := io.TeeReader(input, buf)

	err := operations.SyncTemplate(ctx, tee, io.Discard, state, operations.SyncOptions{})
	if err != nil {
		tflog.Debug(ctx, "error actually encountered during operations.SyncTemplate")
		return err
	}

	t, err := template.Load(buf)
	if err != nil {
		tflog.Debug(ctx, "error actually encountered during template.Load")
		return err
	}

	name, err := t.ClusterName()
	if err != nil {
		tflog.Debug(ctx, "error actually encountered during ClusterName lookup")
		return err
	}

	for {
		cluster, err := safe.StateGetByID[*omni.ClusterStatus](ctx, state, name)
		if err != nil {
			tflog.Debug(ctx, "error actually encountered during ClusterStatus check")
			return err
		}
		if cluster.TypedSpec().Value.Ready {
			break
		}
	}

	return nil
}
