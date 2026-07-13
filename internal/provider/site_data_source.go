package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/martezr/terraform-provider-nightlight/internal/client"
)

var _ datasource.DataSource = &SiteDataSource{}

func NewSiteDataSource() datasource.DataSource { return &SiteDataSource{} }

type SiteDataSource struct{ c *client.Client }

type SiteDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Location    types.String `tfsdk:"location"`
	Type        types.String `tfsdk:"type"`
	Topology    types.String `tfsdk:"topology"`
	Description types.String `tfsdk:"description"`
	Bridges     types.List   `tfsdk:"bridges"`
}

func (d *SiteDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_site"
}

func (d *SiteDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Looks up a Nightlight Cloud site by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Site ID. Provide to look up by ID, or leave unset to look up by `name`.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Site name. Used to look up the site when `id` is not provided.",
			},
			"location":    schema.StringAttribute{Computed: true},
			"type":        schema.StringAttribute{Computed: true},
			"topology":    schema.StringAttribute{Computed: true},
			"description": schema.StringAttribute{Computed: true},
			"bridges": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

func (d *SiteDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type",
			fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	d.c = c
}

func (d *SiteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SiteDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.IsNull() && data.Name.IsNull() {
		resp.Diagnostics.AddError("Missing filter", "At least one of `id` or `name` must be set.")
		return
	}

	var site *client.Site

	if !data.ID.IsNull() {
		found, err := d.c.GetSite(data.ID.ValueString())
		if err == client.ErrNotFound {
			resp.Diagnostics.AddError("Site not found", fmt.Sprintf("No site with ID %q exists.", data.ID.ValueString()))
			return
		}
		if err != nil {
			resp.Diagnostics.AddError("Error reading site", err.Error())
			return
		}
		site = found
	} else {
		all, err := d.c.ListSites()
		if err != nil {
			resp.Diagnostics.AddError("Error listing sites", err.Error())
			return
		}
		for i := range all {
			if all[i].Name == data.Name.ValueString() {
				site = &all[i]
				break
			}
		}
		if site == nil {
			resp.Diagnostics.AddError("Site not found", fmt.Sprintf("No site named %q exists.", data.Name.ValueString()))
			return
		}
	}

	data.ID = types.StringValue(site.ID)
	data.Name = types.StringValue(site.Name)
	data.Location = types.StringValue(site.Location)
	data.Type = types.StringValue(site.Type)
	data.Topology = types.StringValue(site.Topology)
	data.Description = types.StringValue(site.Description)
	bridges, _ := types.ListValueFrom(ctx, types.StringType, site.Bridges)
	data.Bridges = bridges

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
