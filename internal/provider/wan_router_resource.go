package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/martezr/terraform-provider-nightlight/internal/client"
)

var _ resource.Resource = &WanRouterResource{}
var _ resource.ResourceWithImportState = &WanRouterResource{}

func NewWanRouterResource() resource.Resource { return &WanRouterResource{} }

type WanRouterResource struct{ c *client.Client }

type WanRouterResourceModel struct {
	ID           types.String `tfsdk:"id"`
	WANIPAddress types.String `tfsdk:"wan_ip_address"`
	WANNetmask   types.String `tfsdk:"wan_netmask"`
	WANGateway   types.String `tfsdk:"wan_gateway"`
	StaticRoutes types.List   `tfsdk:"static_routes"`
	NATEnabled   types.Bool   `tfsdk:"nat_enabled"`
}

type staticRouteModel struct {
	Destination types.String `tfsdk:"destination"`
	Gateway     types.String `tfsdk:"gateway"`
	Interface   types.String `tfsdk:"interface"`
}

var staticRouteAttrTypes = map[string]attr.Type{
	"destination": types.StringType,
	"gateway":     types.StringType,
	"interface":   types.StringType,
}

func (r *WanRouterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wan_router"
}

func (r *WanRouterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the Nightlight Cloud WAN router configuration. This is a singleton resource — only one WAN router exists per deployment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"wan_ip_address": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Public IP address assigned to the WAN interface.",
			},
			"wan_netmask": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Subnet mask for the WAN interface (e.g. `255.255.255.0`).",
			},
			"wan_gateway": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Default gateway IP address for the WAN interface.",
			},
			"nat_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Enable NAT/masquerade on the WAN interface.",
			},
			"static_routes": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				Default:  listdefault.StaticValue(types.ListValueMust(types.ObjectType{AttrTypes: staticRouteAttrTypes}, []attr.Value{})),
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"destination": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Destination CIDR block (e.g. `10.0.0.0/24`).",
						},
						"gateway": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Next-hop gateway IP address.",
						},
						"interface": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							MarkdownDescription: "MAC address of the egress network interface.",
						},
					},
				},
			},
		},
	}
}

func (r *WanRouterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create applies the desired WAN router config. The router already exists as a system singleton.
func (r *WanRouterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WanRouterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, diags := wanRouterFromModel(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.c.UpdateWanRouterConfig(cfg)
	if err != nil {
		resp.Diagnostics.AddError("Error configuring WAN router", err.Error())
		return
	}

	// Also push to the running VM's RouterConfig so the change takes effect
	// without requiring a VM restart. Skipped if the VM doesn't exist yet.
	if syncErr := r.syncRouterConfig(cfg); syncErr != nil {
		resp.Diagnostics.AddWarning("WAN router config saved but running VM not updated", syncErr.Error())
	}

	resp.Diagnostics.Append(wanRouterToModel(ctx, updated, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WanRouterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WanRouterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.c.GetWanRouterConfig()
	if err == client.ErrNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading WAN router config", err.Error())
		return
	}

	resp.Diagnostics.Append(wanRouterToModel(ctx, cfg, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WanRouterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data WanRouterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	var state WanRouterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = state.ID

	cfg, diags := wanRouterFromModel(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.c.UpdateWanRouterConfig(cfg)
	if err != nil {
		resp.Diagnostics.AddError("Error updating WAN router config", err.Error())
		return
	}

	if syncErr := r.syncRouterConfig(cfg); syncErr != nil {
		resp.Diagnostics.AddWarning("WAN router config saved but running VM not updated", syncErr.Error())
	}

	resp.Diagnostics.Append(wanRouterToModel(ctx, updated, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete is a no-op — the WAN router is a system singleton and cannot be destroyed.
func (r *WanRouterResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *WanRouterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// syncRouterConfig pushes WAN IP/netmask/gateway, NAT, and static routes into
// the running VM's RouterConfig so changes take effect without a VM restart.
// Returns an error only if the VM exists but the update fails; a 404 (VM not
// yet created) is silently ignored.
func (r *WanRouterResource) syncRouterConfig(wan client.WanRouterConfig) error {
	const wanRouterID = "wanrouter"
	current, err := r.c.GetRouterConfig(wanRouterID)
	if err == client.ErrNotFound {
		return nil
	}
	if err != nil {
		return err
	}

	// Update only the WAN-facing interface (index 0); leave the transit
	// interface (index 1, fixed 172.16.100.1) untouched.
	if len(current.Interfaces) > 0 {
		current.Interfaces[0].IPAddress = wan.WANIPAddress
		current.Interfaces[0].Netmask = wan.WANNetmask
		current.Interfaces[0].Gateway = wan.WANGateway
	}

	current.NAT.Enabled = wan.NATEnabled

	routes := make([]client.StaticRoute, len(wan.StaticRoutes))
	copy(routes, wan.StaticRoutes)
	current.StaticRoutes = routes

	_, err = r.c.UpdateRouterConfig(wanRouterID, *current)
	return err
}

func wanRouterFromModel(ctx context.Context, m WanRouterResourceModel) (client.WanRouterConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	cfg := client.WanRouterConfig{
		WANIPAddress: m.WANIPAddress.ValueString(),
		WANNetmask:   m.WANNetmask.ValueString(),
		WANGateway:   m.WANGateway.ValueString(),
		NATEnabled:   m.NATEnabled.ValueBool(),
	}
	if !m.StaticRoutes.IsNull() && !m.StaticRoutes.IsUnknown() {
		var routes []staticRouteModel
		diags.Append(m.StaticRoutes.ElementsAs(ctx, &routes, false)...)
		for _, r := range routes {
			cfg.StaticRoutes = append(cfg.StaticRoutes, client.StaticRoute{
				Destination: r.Destination.ValueString(),
				Gateway:     r.Gateway.ValueString(),
				Interface:   r.Interface.ValueString(),
			})
		}
	}
	return cfg, diags
}

func wanRouterToModel(_ context.Context, cfg *client.WanRouterConfig, m *WanRouterResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	m.ID = types.StringValue(cfg.ID)
	m.WANIPAddress = types.StringValue(cfg.WANIPAddress)
	m.WANNetmask = types.StringValue(cfg.WANNetmask)
	m.WANGateway = types.StringValue(cfg.WANGateway)
	m.NATEnabled = types.BoolValue(cfg.NATEnabled)

	routes := make([]attr.Value, 0, len(cfg.StaticRoutes))
	for _, sr := range cfg.StaticRoutes {
		obj, d := types.ObjectValue(staticRouteAttrTypes, map[string]attr.Value{
			"destination": types.StringValue(sr.Destination),
			"gateway":     types.StringValue(sr.Gateway),
			"interface":   types.StringValue(sr.Interface),
		})
		diags.Append(d...)
		routes = append(routes, obj)
	}
	list, d := types.ListValue(types.ObjectType{AttrTypes: staticRouteAttrTypes}, routes)
	diags.Append(d...)
	m.StaticRoutes = list
	return diags
}
