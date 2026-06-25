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

var _ resource.Resource = &SwitchResource{}
var _ resource.ResourceWithImportState = &SwitchResource{}

func NewSwitchResource() resource.Resource { return &SwitchResource{} }

type SwitchResource struct{ c *client.Client }

type SwitchResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	SiteId      types.String `tfsdk:"site_id"`
	BridgeName  types.String `tfsdk:"bridge_name"`
	Type        types.String `tfsdk:"type"`
	Tags        types.List   `tfsdk:"tags"`
}

func (r *SwitchResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_switch"
}

func (r *SwitchResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Nightlight Cloud OVS switch.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{Required: true},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"site_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"bridge_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("core"),
				MarkdownDescription: "Switch type: `core`, `leaf`, or `access`.",
			},
			"tags": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *SwitchResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SwitchResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SwitchResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sw := switchFromModel(ctx, data)
	created, err := r.c.CreateSwitch(sw)
	if err != nil {
		resp.Diagnostics.AddError("Error creating switch", err.Error())
		return
	}

	switchToModel(ctx, created, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SwitchResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SwitchResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sw, err := r.c.GetSwitch(data.ID.ValueString())
	if err == client.ErrNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading switch", err.Error())
		return
	}

	switchToModel(ctx, sw, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SwitchResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SwitchResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	var state SwitchResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = state.ID

	sw := switchFromModel(ctx, data)
	sw.ID = data.ID.ValueString()
	updated, err := r.c.UpdateSwitch(data.ID.ValueString(), sw)
	if err != nil {
		resp.Diagnostics.AddError("Error updating switch", err.Error())
		return
	}

	switchToModel(ctx, updated, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SwitchResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SwitchResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.c.DeleteSwitch(data.ID.ValueString()); err != nil && err != client.ErrNotFound {
		resp.Diagnostics.AddError("Error deleting switch", err.Error())
	}
}

func (r *SwitchResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func switchFromModel(ctx context.Context, m SwitchResourceModel) client.Switch {
	sw := client.Switch{
		ID:          m.ID.ValueString(),
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueString(),
		SiteId:      m.SiteId.ValueString(),
		BridgeName:  m.BridgeName.ValueString(),
		Type:        m.Type.ValueString(),
	}
	if !m.Tags.IsNull() && !m.Tags.IsUnknown() {
		m.Tags.ElementsAs(ctx, &sw.Tags, false)
	}
	return sw
}

func switchToModel(ctx context.Context, sw *client.Switch, m *SwitchResourceModel) {
	m.ID = types.StringValue(sw.ID)
	m.Name = types.StringValue(sw.Name)
	m.Description = types.StringValue(sw.Description)
	m.SiteId = types.StringValue(sw.SiteId)
	m.BridgeName = types.StringValue(sw.BridgeName)
	m.Type = types.StringValue(sw.Type)
	tags, _ := types.ListValueFrom(ctx, types.StringType, sw.Tags)
	m.Tags = tags
}
