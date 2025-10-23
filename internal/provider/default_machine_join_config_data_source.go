package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/siderolabs/omni/client/pkg/client"
)

type omniDefaultMachineJoinConfigDataSource struct {
	omniClient *client.Client
}

type OmniDefaultMachineJoinConfigModelV0 struct {
	ID         types.String `tfsdk:"id"`
	KernelArgs types.List   `tfsdk:"kernel_args"`
	ConfigYAML types.String `tfsdk:"config_yaml"`
}

var _ datasource.DataSource = &omniDefaultMachineJoinConfigDataSource{}

func NewOmniDefaultMachineJoinConfigDataSource() datasource.DataSource {
	return &omniDefaultMachineJoinConfigDataSource{}
}

func (d *omniDefaultMachineJoinConfigDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_default_machine_join_config"
}

func (d *omniDefaultMachineJoinConfigDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Omni default machine join config definition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Name (ID) of the join token.",
				Computed:    true,
			},
			"kernel_args": schema.ListAttribute{
				ElementType: types.StringType,
				Description: "Kernel arguments for default machine join to Omni instance.",
				Computed:    true,
				Sensitive:   true,
			},
			"config_yaml": schema.StringAttribute{
				Description: "Returned full default machine join config in YAML.",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func (d *omniDefaultMachineJoinConfigDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	omniClient, ok := req.ProviderData.(*client.Client)
	tflog.Debug(context.Background(), "Configuring default omni machine join config datasource")
	if !ok {
		resp.Diagnostics.AddError(
			"failed to get omni client",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.omniClient = omniClient
}

func (d *omniDefaultMachineJoinConfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.omniClient == nil {
		resp.Diagnostics.AddError("omni client is not configured", "Please report this issue to the provider developers.")
		return
	}

	var config OmniDefaultMachineJoinConfigModelV0
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	managementClient := d.omniClient.Management()
	joinConfig, err := managementClient.GetMachineJoinConfig(ctx, "", true)
	if err != nil {
		resp.Diagnostics.AddError("Error retrieving machine join confg", err.Error())
	}

	tflog.Debug(ctx, fmt.Sprintf("machine join config (kernel args):\n%s", joinConfig.GetKernelArgs()))
	tflog.Debug(ctx, fmt.Sprintf("machine join config (YAML):\n%s", joinConfig.GetConfig()))

	config.KernelArgs, diags = types.ListValueFrom(ctx, types.StringType, joinConfig.GetKernelArgs())
	if diags.HasError() {
		return
	}
	config.ConfigYAML = types.StringValue(joinConfig.GetConfig())
	config.ID = types.StringValue("default")

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
