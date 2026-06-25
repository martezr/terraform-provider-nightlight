package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/martezr/terraform-provider-nightlight/internal/client"
)

var _ resource.Resource = &SubnetResource{}
var _ resource.ResourceWithImportState = &SubnetResource{}

func NewSubnetResource() resource.Resource { return &SubnetResource{} }

type SubnetResource struct{ c *client.Client }

type SubnetResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	CIDRBlock   types.String `tfsdk:"cidr_block"`
	SiteId      types.String `tfsdk:"site_id"`
	VLANId      types.Int64  `tfsdk:"vlan_id"`
	DNSServers  types.List   `tfsdk:"dns_servers"`
	NTPServers  types.List   `tfsdk:"ntp_servers"`
	DomainName  types.String `tfsdk:"domain_name"`
	BridgeName  types.String `tfsdk:"bridge_name"`
	DHCPServer  types.Bool   `tfsdk:"dhcp_server"`
	IPPoolRange types.String `tfsdk:"ip_pool_range"`
	Gateway     types.String `tfsdk:"gateway"`
	Tags        types.Map    `tfsdk:"tags"`
}

func (r *SubnetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

func (r *SubnetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Nightlight Cloud subnet.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name":       schema.StringAttribute{Required: true},
			"cidr_block": schema.StringAttribute{Required: true},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"site_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"vlan_id": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"dns_servers": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"ntp_servers": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"domain_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"bridge_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"dhcp_server": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"ip_pool_range": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"gateway": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"tags": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *SubnetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SubnetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SubnetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.c.CreateSubnet(subnetFromModel(ctx, data))
	if err != nil {
		resp.Diagnostics.AddError("Error creating subnet", err.Error())
		return
	}

	subnetToModel(ctx, created, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubnetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SubnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sub, err := r.c.GetSubnet(data.ID.ValueString())
	if err == client.ErrNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading subnet", err.Error())
		return
	}

	subnetToModel(ctx, sub, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubnetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SubnetResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	var state SubnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = state.ID

	updated, err := r.c.UpdateSubnet(data.ID.ValueString(), subnetFromModel(ctx, data))
	if err != nil {
		resp.Diagnostics.AddError("Error updating subnet", err.Error())
		return
	}

	subnetToModel(ctx, updated, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubnetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SubnetResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.c.DeleteSubnet(data.ID.ValueString()); err != nil && err != client.ErrNotFound {
		resp.Diagnostics.AddError("Error deleting subnet", err.Error())
	}
}

func (r *SubnetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func subnetFromModel(ctx context.Context, m SubnetResourceModel) client.Subnet {
	sub := client.Subnet{
		ID:          m.ID.ValueString(),
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueString(),
		CIDRBlock:   m.CIDRBlock.ValueString(),
		SiteId:      m.SiteId.ValueString(),
		VLANId:      m.VLANId.ValueInt64(),
		DomainName:  m.DomainName.ValueString(),
		BridgeName:  m.BridgeName.ValueString(),
		DHCPServer:  m.DHCPServer.ValueBool(),
		IPPoolRange: m.IPPoolRange.ValueString(),
		Gateway:     m.Gateway.ValueString(),
		Tags:        tagsFromModel(ctx, m.Tags),
	}
	if !m.DNSServers.IsNull() && !m.DNSServers.IsUnknown() {
		m.DNSServers.ElementsAs(ctx, &sub.DNSServers, false)
	}
	if !m.NTPServers.IsNull() && !m.NTPServers.IsUnknown() {
		m.NTPServers.ElementsAs(ctx, &sub.NTPServers, false)
	}
	return sub
}

func subnetToModel(ctx context.Context, sub *client.Subnet, m *SubnetResourceModel) {
	m.ID = types.StringValue(sub.ID)
	m.Name = types.StringValue(sub.Name)
	m.Description = types.StringValue(sub.Description)
	m.CIDRBlock = types.StringValue(sub.CIDRBlock)
	m.SiteId = types.StringValue(sub.SiteId)
	m.VLANId = types.Int64Value(sub.VLANId)
	m.DomainName = types.StringValue(sub.DomainName)
	m.BridgeName = types.StringValue(sub.BridgeName)
	m.DHCPServer = types.BoolValue(sub.DHCPServer)
	m.IPPoolRange = types.StringValue(sub.IPPoolRange)
	m.Gateway = types.StringValue(sub.Gateway)
	m.Tags = tagsToModel(ctx, sub.Tags)

	dns, _ := types.ListValueFrom(ctx, types.StringType, sub.DNSServers)
	m.DNSServers = dns
	ntp, _ := types.ListValueFrom(ctx, types.StringType, sub.NTPServers)
	m.NTPServers = ntp
}
