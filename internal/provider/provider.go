// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/siderolabs/omni/client/pkg/client"
)

// Ensure OmniProvider satisfies various provider interfaces.
var _ provider.Provider = &OmniProvider{}

// OmniProvider defines the provider implementation.
type OmniProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// OmniProviderModel describes the provider data model.
type OmniProviderModelV0 struct {
	Endpoint          types.String `tfsdk:"endpoint"`
	ServiceAccountKey types.String `tfsdk:"service_account_key"`
}

func (p *OmniProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "omni"
	resp.Version = p.version
}

func (p *OmniProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "URL for the Omni API endpoint. May also be provided via the OMNI_ENDPOINT environment variable.",
				Optional:            true,
			},
			"service_account_key": schema.StringAttribute{
				MarkdownDescription: "Service account key for Omni. This is used to authenticate requests to the Omni API. May also be provided via the OMNI_SERVICE_ACCOUNT_KEY environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *OmniProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Omni client")

	endpoint := os.Getenv("OMNI_ENDPOINT")
	service_account_key := os.Getenv("OMNI_SERVICE_ACCOUNT_KEY")

	var config OmniProviderModelV0
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Using Omni API client configuration: %s", endpoint))

	if endpoint == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Missing Omni API Endpoint",
			"The provider cannot create the Omni API client as there is a missing or empty value for the Omni API endpoint. "+
				"Set the endpoint value in the configuration or use the OMNI_ENDPOINT environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if service_account_key == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("service_account_key"),
			"Missing Omni Service Account Key",
			"The provider cannot create the Omni API client as there is a missing or empty value for the Omni API service account key. "+
				"Set the service account key value in the configuration or use the OMNI_SERVICE_ACCOUNT_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "omni_endpoint", endpoint)
	ctx = tflog.SetField(ctx, "omni_service_account_key", service_account_key)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "omni_service_account_key")

	tflog.Debug(ctx, "Creating Omni API client")

	omniClient, err := client.New(endpoint, client.WithServiceAccount(service_account_key))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Omni API Client",
			"An unexpected error occurred when creating the Omni API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Omni Client Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = omniClient
	resp.ResourceData = omniClient

	tflog.Info(ctx, "Configured Omni client", map[string]any{"success": true})
}

func (p *OmniProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewOmniClusterResource,
		NewOmniClusterMachinesTemplateResource,
		NewOmniClusterMachineSetTemplateResource,
	}
}

func (p *OmniProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewOmniMachineStatusDataSource,
		NewOmniClusterTemplateDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &OmniProvider{
			version: version,
		}
	}
}
