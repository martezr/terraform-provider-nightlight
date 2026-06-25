package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/martezr/terraform-provider-nightlight/internal/client"
)

var _ resource.Resource = &SiteResource{}
var _ resource.ResourceWithImportState = &SiteResource{}

func NewSiteResource() resource.Resource { return &SiteResource{} }

type SiteResource struct{ c *client.Client }

type SiteResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Location    types.String `tfsdk:"location"`
	Type        types.String `tfsdk:"type"`
	Topology    types.String `tfsdk:"topology"`
	Description types.String `tfsdk:"description"`
	Bridges     types.List   `tfsdk:"bridges"`
}

func (r *SiteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_site"
}

func (r *SiteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Nightlight Cloud site.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{Required: true},
			"location": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"topology": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("single-bridge"),
				MarkdownDescription: "Network topology: `single-bridge`, `spine-leaf`, or `robo`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"bridges": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

func (r *SiteResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type",
			fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	r.c = c
}

func (r *SiteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SiteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.c.CreateSite(client.Site{
		Name:        data.Name.ValueString(),
		Location:    data.Location.ValueString(),
		Type:        data.Type.ValueString(),
		Topology:    data.Topology.ValueString(),
		Description: data.Description.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating site", err.Error())
		return
	}

	siteToModel(ctx, created, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SiteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SiteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	site, err := r.c.GetSite(data.ID.ValueString())
	if err == client.ErrNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading site", err.Error())
		return
	}

	siteToModel(ctx, site, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SiteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SiteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	var state SiteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = state.ID

	updated, err := r.c.UpdateSite(data.ID.ValueString(), client.Site{
		ID:          data.ID.ValueString(),
		Name:        data.Name.ValueString(),
		Location:    data.Location.ValueString(),
		Type:        data.Type.ValueString(),
		Topology:    data.Topology.ValueString(),
		Description: data.Description.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating site", err.Error())
		return
	}

	siteToModel(ctx, updated, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SiteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SiteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.c.DeleteSite(data.ID.ValueString()); err != nil && err != client.ErrNotFound {
		resp.Diagnostics.AddError("Error deleting site", err.Error())
	}
}

func (r *SiteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func siteToModel(ctx context.Context, s *client.Site, m *SiteResourceModel) {
	m.ID = types.StringValue(s.ID)
	m.Name = types.StringValue(s.Name)
	m.Location = types.StringValue(s.Location)
	m.Type = types.StringValue(s.Type)
	m.Topology = types.StringValue(s.Topology)
	m.Description = types.StringValue(s.Description)
	bridges, _ := types.ListValueFrom(ctx, types.StringType, s.Bridges)
	m.Bridges = bridges
}
