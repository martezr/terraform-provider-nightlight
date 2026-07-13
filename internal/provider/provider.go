package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/martezr/terraform-provider-nightlight/internal/client"
)

var _ provider.Provider = &NightlightProvider{}
var _ provider.ProviderWithFunctions = &NightlightProvider{}

type NightlightProvider struct {
	version string
}

type NightlightProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (p *NightlightProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "nightlight"
	resp.Version = p.version
}

func (p *NightlightProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Interact with a Nightlight Cloud hypervisor.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Nightlight Cloud API endpoint (e.g. http://192.168.1.10). Defaults to NIGHTLIGHT_ENDPOINT env var.",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for authentication. Defaults to NIGHTLIGHT_USERNAME env var or 'root'.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for authentication. Defaults to NIGHTLIGHT_PASSWORD env var.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *NightlightProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data NightlightProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := os.Getenv("NIGHTLIGHT_ENDPOINT")
	if !data.Endpoint.IsNull() {
		endpoint = data.Endpoint.ValueString()
	}
	if endpoint == "" {
		resp.Diagnostics.AddError("Missing endpoint", "Set the endpoint attribute or NIGHTLIGHT_ENDPOINT environment variable.")
		return
	}

	username := os.Getenv("NIGHTLIGHT_USERNAME")
	if username == "" {
		username = "root"
	}
	if !data.Username.IsNull() {
		username = data.Username.ValueString()
	}

	password := os.Getenv("NIGHTLIGHT_PASSWORD")
	if password == "" {
		password = "nightlight"
	}
	if !data.Password.IsNull() {
		password = data.Password.ValueString()
	}

	c, err := client.NewClient(endpoint, username, password)
	if err != nil {
		resp.Diagnostics.AddError("Client configuration error", err.Error())
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *NightlightProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewInstanceResource,
		NewInstanceBootCommandsResource,
		NewSubnetResource,
		NewDatastoreResource,
		NewSiteResource,
		NewSwitchResource,
		NewImageResource,
		NewWanRouterResource,
		NewContentLibraryResource,
	}
}

func (p *NightlightProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewImageDataSource,
		NewSiteDataSource,
		NewSubnetDataSource,
	}
}

func (p *NightlightProvider) Functions(_ context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &NightlightProvider{version: version}
	}
}
