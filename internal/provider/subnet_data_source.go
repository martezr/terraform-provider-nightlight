package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/martezr/terraform-provider-nightlight/internal/client"
)

var _ datasource.DataSource = &SubnetDataSource{}

func NewSubnetDataSource() datasource.DataSource { return &SubnetDataSource{} }

type SubnetDataSource struct{ c *client.Client }

type SubnetDataSourceModel struct {
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

func (d *SubnetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

func (d *SubnetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Looks up a Nightlight Cloud subnet by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Subnet ID. Provide to look up by ID, or leave unset to look up by `name`.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Subnet name. Used to look up the subnet when `id` is not provided.",
			},
			"description":   schema.StringAttribute{Computed: true},
			"cidr_block":    schema.StringAttribute{Computed: true},
			"site_id":       schema.StringAttribute{Computed: true},
			"vlan_id":       schema.Int64Attribute{Computed: true},
			"domain_name":   schema.StringAttribute{Computed: true},
			"bridge_name":   schema.StringAttribute{Computed: true},
			"dhcp_server":   schema.BoolAttribute{Computed: true},
			"ip_pool_range": schema.StringAttribute{Computed: true},
			"gateway":       schema.StringAttribute{Computed: true},
			"dns_servers": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"ntp_servers": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"tags": schema.MapAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

func (d *SubnetDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SubnetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SubnetDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.IsNull() && data.Name.IsNull() {
		resp.Diagnostics.AddError("Missing filter", "At least one of `id` or `name` must be set.")
		return
	}

	var sub *client.Subnet

	if !data.ID.IsNull() {
		found, err := d.c.GetSubnet(data.ID.ValueString())
		if err == client.ErrNotFound {
			resp.Diagnostics.AddError("Subnet not found", fmt.Sprintf("No subnet with ID %q exists.", data.ID.ValueString()))
			return
		}
		if err != nil {
			resp.Diagnostics.AddError("Error reading subnet", err.Error())
			return
		}
		sub = found
	} else {
		all, err := d.c.ListSubnets()
		if err != nil {
			resp.Diagnostics.AddError("Error listing subnets", err.Error())
			return
		}
		for i := range all {
			if all[i].Name == data.Name.ValueString() {
				sub = &all[i]
				break
			}
		}
		if sub == nil {
			resp.Diagnostics.AddError("Subnet not found", fmt.Sprintf("No subnet named %q exists.", data.Name.ValueString()))
			return
		}
	}

	data.ID = types.StringValue(sub.ID)
	data.Name = types.StringValue(sub.Name)
	data.Description = types.StringValue(sub.Description)
	data.CIDRBlock = types.StringValue(sub.CIDRBlock)
	data.SiteId = types.StringValue(sub.SiteId)
	data.VLANId = types.Int64Value(sub.VLANId)
	data.DomainName = types.StringValue(sub.DomainName)
	data.BridgeName = types.StringValue(sub.BridgeName)
	data.DHCPServer = types.BoolValue(sub.DHCPServer)
	data.IPPoolRange = types.StringValue(sub.IPPoolRange)
	data.Gateway = types.StringValue(sub.Gateway)
	data.Tags = tagsToModel(ctx, sub.Tags)

	dns, _ := types.ListValueFrom(ctx, types.StringType, sub.DNSServers)
	data.DNSServers = dns
	ntp, _ := types.ListValueFrom(ctx, types.StringType, sub.NTPServers)
	data.NTPServers = ntp

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
