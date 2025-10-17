package provider

import (
	"context"
	"fmt"
	"terraform-provider-omni/internal/models"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/siderolabs/omni/client/pkg/client"
	"github.com/siderolabs/omni/client/pkg/client/management"
	"gopkg.in/yaml.v3"
)

var (
	_ resource.Resource              = &omniClusterKubeConfig{}
	_ resource.ResourceWithConfigure = &omniClusterKubeConfig{}
)

type omniClusterKubeConfig struct {
	omniClient *client.Client
	context    context.Context
}

type KubeConfigClusterModel struct {
	Cluster KubeConfigClusterDetailsModel `tfsdk:"cluster"`
	Name    types.String                  `tfsdk:"name"`
}

type KubeConfigClusterDetailsModel struct {
	Server types.String `tfsdk:"server"`
}

type KubeConfigContextModel struct {
	Context KubeConfigContextDetailsModel `tfsdk:"context"`
	Name    types.String                  `tfsdk:"name"`
}

type KubeConfigContextDetailsModel struct {
	Cluster   types.String `tfsdk:"cluster"`
	Namespace types.String `tfsdk:"namespace"`
	User      types.String `tfsdk:"user"`
}

type KubeConfigUserModel struct {
	Name types.String               `tfsdk:"name"`
	User KubeConfigUserDetailsModel `tfsdk:"user"`
}

type KubeConfigUserDetailsModel struct {
	Token types.String `tfsdk:"token"`
}

type KubeConfigModel struct {
	Clusters []KubeConfigClusterModel `tfsdk:"clusters"`
	Contexts []KubeConfigContextModel `tfsdk:"contexts"`
	Users    []KubeConfigUserModel    `tfsdk:"users"`
}

type OmniClusterKubeConfigModelV0 struct {
	ID          types.String `tfsdk:"id"`
	CreatedAt   types.String `tfsdk:"created_at"`
	LastUpdated types.String `tfsdk:"last_updated"`
	User        types.String `tfsdk:"user"`
	Groups      types.List   `tfsdk:"groups"`
	Cluster     types.String `tfsdk:"cluster"`
	Clusters    types.List   `tfsdk:"clusters"`
	Contexts    types.List   `tfsdk:"contexts"`
	Users       types.List   `tfsdk:"users"`
	YAML        types.String `tfsdk:"yaml"`
}

func NewOmniClusterKubeConfigResource() resource.Resource {
	return &omniClusterKubeConfig{}
}

func (r *omniClusterKubeConfig) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_kubeconfig"
}

func (r *omniClusterKubeConfig) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	omniClient, ok := req.ProviderData.(*client.Client)
	tflog.Debug(context.Background(), "Configuring omni kubeconfig resource")
	if !ok {
		resp.Diagnostics.AddError(
			"failed to get image factory client",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.omniClient = omniClient
}

func (r *omniClusterKubeConfig) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Omni cluster kubeconfig definition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Name (ID) of cluster that kubeconfig belongs to.",
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
			"cluster": schema.StringAttribute{
				Required:    true,
				Description: "Name of the cluster.",
			},
			"user": schema.StringAttribute{
				Optional:    true,
				Description: "User to use for generated kubeconfig.",
			},
			"groups": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Groups to use for generated kubeconfig.",
			},
			"clusters": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the Kubernetes cluster.",
						},
						"cluster": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"server": schema.StringAttribute{
									Computed:    true,
									Description: "Endpoint for Kubernetes cluster.",
								},
							},
							Computed:    true,
							Description: "Kubernetes cluster information.",
						},
					},
				},
				Computed:    true,
				Description: "Clusters defined in kubeconfig.",
			},
			"contexts": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the Kubernetes cluster.",
						},
						"context": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"cluster": schema.StringAttribute{
									Computed:    true,
									Description: "Kubernetes cluster associated with context.",
								},
								"namespace": schema.StringAttribute{
									Computed:    true,
									Description: "Default namespace for context.",
								},
								"user": schema.StringAttribute{
									Computed:    true,
									Description: "Authenticated user associated with context.",
								},
							},
							Computed:    true,
							Description: "A context associated with kubeconfig.",
						},
					},
				},
				Computed:    true,
				Description: "Clusters defined in kubeconfig.",
			},
			"users": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the user.",
						},
						"user": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"token": schema.StringAttribute{
									Computed:    true,
									Description: "Token used by user to authenticate to Kubernetes cluster.",
								},
							},
							Computed:    true,
							Description: "User information.",
						},
					},
				},
				Computed:    true,
				Description: "Clusters defined in kubeconfig.",
			},
			"yaml": schema.StringAttribute{
				Description: "Returned kubeconfig in YAML.",
				Computed:    true,
			},
		},
	}
}

func (r *omniClusterKubeConfig) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OmniClusterKubeConfigModelV0

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var user string
	if plan.User.IsNull() {
		user = "admin"
	} else {
		user = plan.User.ValueString()
	}

	var groups []string
	if plan.Groups.IsNull() {
		groups = []string{"system:masters"}
	}

	kubeconfig, err := r.omniClient.Management().WithCluster(plan.Cluster.ValueString()).Kubeconfig(ctx, management.WithServiceAccount(24*time.Hour, user, groups...))
	if err != nil {
		resp.Diagnostics.AddError("Error encountered getting cluster kubeconfig", err.Error())
	}

	tflog.Debug(ctx, "Now unmarshalling....")

	var kubeconfigUnmarshalled models.KubeConfig
	yamlErr := yaml.Unmarshal(kubeconfig, &kubeconfigUnmarshalled)
	if yamlErr != nil {
		resp.Diagnostics.AddError("Could not unmarshall kubeconfig", yamlErr.Error())
		return
	}

	tflog.Debug(ctx, "Unmarshalled....")

	tflog.Debug(ctx, fmt.Sprintf("Clusters:\n%s", kubeconfigUnmarshalled.Clusters[0].Name))
	tflog.Debug(ctx, fmt.Sprintf("Contexts:\n%s", kubeconfigUnmarshalled.Contexts))
	tflog.Debug(ctx, fmt.Sprintf("Users:\n%s", kubeconfigUnmarshalled.Users))

	var clusterObjects []KubeConfigClusterModel
	for _, clusterObject := range kubeconfigUnmarshalled.Clusters {
		clusterObjects = append(clusterObjects, KubeConfigClusterModel{
			Name: types.StringValue(clusterObject.Name),
			Cluster: KubeConfigClusterDetailsModel{
				Server: types.StringValue(clusterObject.Cluster.Server),
			},
		})
	}
	plan.Clusters, diags = types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name": types.StringType,
			"cluster": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"server": types.StringType,
				},
			},
		},
	}, clusterObjects)
	if diags.HasError() {
		return
	}

	var contextObjects []KubeConfigContextModel
	for _, contextObject := range kubeconfigUnmarshalled.Contexts {
		contextObjects = append(contextObjects, KubeConfigContextModel{
			Name: types.StringValue(contextObject.Name),
			Context: KubeConfigContextDetailsModel{
				Cluster:   types.StringValue(contextObject.Context.Cluster),
				Namespace: types.StringValue(contextObject.Context.Namespace),
				User:      types.StringValue(contextObject.Context.User),
			},
		})
	}
	plan.Contexts, diags = types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name": types.StringType,
			"context": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"cluster":   types.StringType,
					"namespace": types.StringType,
					"user":      types.StringType,
				},
			},
		},
	}, contextObjects)
	if diags.HasError() {
		return
	}

	var userObjects []KubeConfigUserModel
	for _, userObject := range kubeconfigUnmarshalled.Users {
		userObjects = append(userObjects, KubeConfigUserModel{
			Name: types.StringValue(userObject.Name),
			User: KubeConfigUserDetailsModel{
				Token: types.StringValue(userObject.User.Token),
			},
		})
	}
	plan.Users, diags = types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name": types.StringType,
			"user": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"token": types.StringType,
				},
			},
		},
	}, userObjects)
	if diags.HasError() {
		return
	}

	plan.YAML = types.StringValue(string(kubeconfig))

	tflog.Debug(ctx, fmt.Sprintf("kubeconfig for cluster %s:\n%s", plan.Cluster.ValueString(), plan.YAML.ValueString()))

	plan.ID = plan.Cluster
	timeNow := types.StringValue(time.Now().Format(time.RFC850))
	plan.CreatedAt = timeNow
	plan.LastUpdated = timeNow

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *omniClusterKubeConfig) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.omniClient == nil {
		resp.Diagnostics.AddError("omni client is not configured", "Please report this issue to the provider developers.")
		return
	}

	var config OmniClusterKubeConfigModelV0
	resp.Diagnostics.Append(req.State.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Read cluster kubeconfig from state.")
	tflog.Trace(ctx, fmt.Sprintf("kubeconfig:\n%s", config.YAML))

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (r *omniClusterKubeConfig) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OmniClusterKubeConfigModelV0
	var state OmniClusterKubeConfigModelV0

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("kubeconfig for %s in state:\n%s", state.Cluster, state.YAML.ValueString()))

	kubeconfig, err := r.omniClient.Management().WithCluster(plan.Cluster.ValueString()).Kubeconfig(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error encountered getting cluster kubeconfig", err.Error())
	}

	var kubeconfigUnmarshalled models.KubeConfig
	yamlErr := yaml.Unmarshal(kubeconfig, &kubeconfigUnmarshalled)
	if yamlErr != nil {
		resp.Diagnostics.AddError("Could not unmarshall kubeconfig", err.Error())
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Clusters:\n%s", kubeconfigUnmarshalled.Clusters))
	tflog.Debug(ctx, fmt.Sprintf("Contexts:\n%s", kubeconfigUnmarshalled.Contexts))
	tflog.Debug(ctx, fmt.Sprintf("Users:\n%s", kubeconfigUnmarshalled.Users))

	// plan.Clusters, _ = types.ListValueFrom(ctx, types.StringType, kubeconfigUnmarshalled.Clusters)
	// plan.Contexts, _ = types.ListValueFrom(ctx, types.StringType, kubeconfigUnmarshalled.Contexts)
	// plan.Users, _ = types.ListValueFrom(ctx, types.StringType, kubeconfigUnmarshalled.Users)
	plan.YAML = types.StringValue(string(kubeconfig))
	plan.ID = plan.Cluster
	timeNow := types.StringValue(time.Now().Format(time.RFC850))
	plan.LastUpdated = timeNow

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *omniClusterKubeConfig) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OmniClusterKubeConfigModelV0
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
