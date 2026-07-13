package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/martezr/terraform-provider-nightlight/internal/client"
)

var _ resource.Resource = &InstanceResource{}
var _ resource.ResourceWithImportState = &InstanceResource{}

func NewInstanceResource() resource.Resource { return &InstanceResource{} }

type InstanceResource struct{ c *client.Client }

// ── model ──────────────────────────────────────────────────────────────────

type InstanceResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	BootType          types.String `tfsdk:"boot_type"`
	CPUCores          types.Int64  `tfsdk:"cpu_cores"`
	CPUSockets        types.Int64  `tfsdk:"cpu_sockets"`
	MemoryMB          types.Int64  `tfsdk:"memory_mb"`
	SiteId            types.String `tfsdk:"site_id"`
	DatastoreId       types.String `tfsdk:"datastore_id"`
	ImageId           types.String `tfsdk:"image_id"`
	UserData          types.String `tfsdk:"user_data"`
	WinAutoattend     types.String `tfsdk:"win_autoattend"`
	IPXEScript        types.String `tfsdk:"ipxe_script"`
	SecureBoot        types.Bool   `tfsdk:"secure_boot"`
	TPM               types.Bool   `tfsdk:"tpm"`
	StartPoweredOff   types.Bool   `tfsdk:"start_powered_off"`
	InstanceType      types.String `tfsdk:"instance_type"`
	Tags              types.Map    `tfsdk:"tags"`
	WaitForGuest      types.Bool   `tfsdk:"wait_for_guest"`
	GuestReadyTimeout types.Int64  `tfsdk:"guest_ready_timeout"`
	// computed
	PowerState           types.String `tfsdk:"power_state"`
	InitializationStatus types.String `tfsdk:"initialization_status"`
	PrimaryIPAddress     types.String `tfsdk:"primary_ip_address"`
	PrimaryMacAddress    types.String `tfsdk:"primary_mac_address"`
	MetadataIPAddress    types.String `tfsdk:"metadata_ip_address"`
	VNCPort              types.Int64  `tfsdk:"vnc_port"`
	GuestIPAddresses     types.List   `tfsdk:"guest_ip_addresses"`
	// device lists
	NetworkInterfaces types.List `tfsdk:"network_interfaces"`
	StorageDisks      types.List `tfsdk:"storage_disks"`
	CDROMs            types.List `tfsdk:"cdroms"`
}

type networkInterfaceModel struct {
	ID          types.String `tfsdk:"id"`
	IndexNumber types.Int64  `tfsdk:"index_number"`
	BootOrder   types.Int64  `tfsdk:"boot_order"`
	Model       types.String `tfsdk:"model"`
	Connected   types.Bool   `tfsdk:"connected"`
	MacAddress  types.String `tfsdk:"mac_address"`
	BridgeName  types.String `tfsdk:"bridge_name"`
	SubnetId    types.String `tfsdk:"subnet_id"`
}

type storageDiskModel struct {
	ID           types.String `tfsdk:"id"`
	IndexNumber  types.Int64  `tfsdk:"index_number"`
	BootOrder    types.Int64  `tfsdk:"boot_order"`
	SizeGB       types.Int64  `tfsdk:"size_gb"`
	BusType      types.String `tfsdk:"bus_type"`
	Path         types.String `tfsdk:"path"`
	DatastoreId  types.String `tfsdk:"datastore_id"`
	ExistingPath types.String `tfsdk:"existing_path"`
	Clone        types.Bool   `tfsdk:"clone"`
}

type cdromModel struct {
	ID          types.String `tfsdk:"id"`
	IndexNumber types.Int64  `tfsdk:"index_number"`
	BootOrder   types.Int64  `tfsdk:"boot_order"`
	Connected   types.Bool   `tfsdk:"connected"`
	Path        types.String `tfsdk:"path"`
}

type bootCommandModel struct {
	Keys           types.String `tfsdk:"keys"`
	Count          types.Int64  `tfsdk:"count"`
	PauseBetweenMs types.Int64  `tfsdk:"pause_between_ms"`
	PauseAfterMs   types.Int64  `tfsdk:"pause_after_ms"`
}

var networkInterfaceAttrTypes = map[string]attr.Type{
	"id":           types.StringType,
	"index_number": types.Int64Type,
	"boot_order":   types.Int64Type,
	"model":        types.StringType,
	"connected":    types.BoolType,
	"mac_address":  types.StringType,
	"bridge_name":  types.StringType,
	"subnet_id":    types.StringType,
}

var storageDiskAttrTypes = map[string]attr.Type{
	"id":            types.StringType,
	"index_number":  types.Int64Type,
	"boot_order":    types.Int64Type,
	"size_gb":       types.Int64Type,
	"bus_type":      types.StringType,
	"path":          types.StringType,
	"datastore_id":  types.StringType,
	"existing_path": types.StringType,
	"clone":         types.BoolType,
}

var cdromAttrTypes = map[string]attr.Type{
	"id":           types.StringType,
	"index_number": types.Int64Type,
	"boot_order":   types.Int64Type,
	"connected":    types.BoolType,
	"path":         types.StringType,
}

var bootCommandAttrTypes = map[string]attr.Type{
	"keys":            types.StringType,
	"count":           types.Int64Type,
	"pause_between_ms": types.Int64Type,
	"pause_after_ms":  types.Int64Type,
}

// ── metadata / schema ──────────────────────────────────────────────────────

func (r *InstanceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

func (r *InstanceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Nightlight Cloud virtual machine instance.",
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
			"boot_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("bios"),
			},
			"cpu_cores":   schema.Int64Attribute{Required: true},
			"cpu_sockets": schema.Int64Attribute{Required: true},
			"memory_mb":   schema.Int64Attribute{Required: true},
			"site_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"datastore_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"image_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_data": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"win_autoattend": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Windows unattend.xml content for automated Windows installations.",
			},
			"ipxe_script": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"secure_boot": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"tpm": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"instance_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("virtualmachine"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"start_powered_off": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"tags": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"wait_for_guest": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "When true (default), Terraform waits for the guest agent to report `ready` before completing the create step.",
			},
			"guest_ready_timeout": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(300),
				MarkdownDescription: "Maximum seconds to wait for `guest_status` to reach `ready`. Defaults to 300 (5 minutes).",
			},
			// computed-only
			"power_state": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"initialization_status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Transient status set by the server during operations such as image capture (e.g. `capturing-image`). Empty when no operation is in progress.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"primary_ip_address": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"primary_mac_address": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"metadata_ip_address": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"vnc_port": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"guest_ip_addresses": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "IP addresses reported by the guest agent.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			// device lists
			"network_interfaces": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
					preserveNICComputed{},
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:      true,
							PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						},
						"index_number": schema.Int64Attribute{Required: true},
						"boot_order":   schema.Int64Attribute{Required: true},
						"model": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString("virtio"),
						},
						"connected": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(true),
						},
						"mac_address": schema.StringAttribute{
							Computed:      true,
							PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						},
						"bridge_name": schema.StringAttribute{Required: true},
						"subnet_id": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString(""),
						},
					},
				},
			},
			"storage_disks": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
					preserveDiskComputed{},
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:      true,
							PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						},
						"index_number": schema.Int64Attribute{Required: true},
						"boot_order":   schema.Int64Attribute{Required: true},
						"size_gb":      schema.Int64Attribute{Required: true},
						"bus_type": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString("virtio"),
						},
						"path": schema.StringAttribute{
							Computed:      true,
							PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						},
						"datastore_id": schema.StringAttribute{Required: true},
						"existing_path": schema.StringAttribute{
							Optional:      true,
							Computed:      true,
							Default:       stringdefault.StaticString(""),
							PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						},
						"clone": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(false),
						},
					},
				},
			},
			"cdroms": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
					preserveCDROMComputed{},
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:      true,
							PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						},
						"index_number": schema.Int64Attribute{Required: true},
						"boot_order":   schema.Int64Attribute{Required: true},
						"connected": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(true),
						},
						"path": schema.StringAttribute{Required: true},
					},
				},
			},
		},
	}
}

// ── configure ──────────────────────────────────────────────────────────────

func (r *InstanceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ── CRUD ───────────────────────────────────────────────────────────────────

func (r *InstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data InstanceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inst := instanceFromModel(ctx, data)
	created, err := r.c.CreateInstance(inst)
	if err != nil {
		resp.Diagnostics.AddError("Error creating instance", err.Error())
		return
	}

	instanceToModel(ctx, created, &data)

	if data.WaitForGuest.ValueBool() {
		timeout := time.Duration(data.GuestReadyTimeout.ValueInt64()) * time.Second
		deadline := time.Now().Add(timeout)
		for time.Now().Before(deadline) {
			time.Sleep(5 * time.Second)
			current, err := r.c.GetInstance(created.ID)
			if err != nil {
				break
			}
			instanceToModel(ctx, current, &data)
			if current.GuestStatus == "ready" {
				break
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data InstanceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inst, err := r.c.GetInstance(data.ID.ValueString())
	if err == client.ErrNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading instance", err.Error())
		return
	}

	instanceToModel(ctx, inst, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data InstanceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	// Preserve computed fields from state.
	var state InstanceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = state.ID

	inst := instanceFromModel(ctx, data)
	inst.ID = data.ID.ValueString()
	updated, err := r.c.UpdateInstance(data.ID.ValueString(), inst)
	if err != nil {
		resp.Diagnostics.AddError("Error updating instance", err.Error())
		return
	}

	instanceToModel(ctx, updated, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data InstanceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.c.DeleteInstance(data.ID.ValueString()); err != nil && err != client.ErrNotFound {
		resp.Diagnostics.AddError("Error deleting instance", err.Error())
	}
}

func (r *InstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ── converters ─────────────────────────────────────────────────────────────

func instanceFromModel(ctx context.Context, m InstanceResourceModel) client.Instance {
	inst := client.Instance{
		Name:            m.Name.ValueString(),
		Description:     m.Description.ValueString(),
		BootType:        m.BootType.ValueString(),
		CPUCores:        int(m.CPUCores.ValueInt64()),
		CPUSockets:      int(m.CPUSockets.ValueInt64()),
		MemoryMB:        int(m.MemoryMB.ValueInt64()),
		SiteId:          m.SiteId.ValueString(),
		DatastoreId:     m.DatastoreId.ValueString(),
		ImageId:         m.ImageId.ValueString(),
		UserData:        m.UserData.ValueString(),
		WinAutoattend:   m.WinAutoattend.ValueString(),
		IPXEScript:      m.IPXEScript.ValueString(),
		SecureBoot:      m.SecureBoot.ValueBool(),
		TPM:             m.TPM.ValueBool(),
		StartPoweredOff: m.StartPoweredOff.ValueBool(),
		InstanceType:    m.InstanceType.ValueString(),
		Tags:            tagsFromModel(ctx, m.Tags),
	}

	// network interfaces
	if !m.NetworkInterfaces.IsNull() && !m.NetworkInterfaces.IsUnknown() {
		var nics []networkInterfaceModel
		m.NetworkInterfaces.ElementsAs(ctx, &nics, false)
		for _, n := range nics {
			inst.Devices.NetworkInterfaces = append(inst.Devices.NetworkInterfaces, client.NetworkInterface{
				ID:          n.ID.ValueString(),
				IndexNumber: int(n.IndexNumber.ValueInt64()),
				BootOrder:   int(n.BootOrder.ValueInt64()),
				Model:       n.Model.ValueString(),
				Connected:   n.Connected.ValueBool(),
				MacAddress:  n.MacAddress.ValueString(),
				BridgeName:  n.BridgeName.ValueString(),
				SubnetId:    n.SubnetId.ValueString(),
			})
		}
	}

	// storage disks
	if !m.StorageDisks.IsNull() && !m.StorageDisks.IsUnknown() {
		var disks []storageDiskModel
		m.StorageDisks.ElementsAs(ctx, &disks, false)
		for _, d := range disks {
			inst.Devices.StorageDisks = append(inst.Devices.StorageDisks, client.StorageDisk{
				ID:           d.ID.ValueString(),
				IndexNumber:  int(d.IndexNumber.ValueInt64()),
				BootOrder:    int(d.BootOrder.ValueInt64()),
				SizeGB:       int(d.SizeGB.ValueInt64()),
				BusType:      d.BusType.ValueString(),
				Path:         d.Path.ValueString(),
				DatastoreId:  d.DatastoreId.ValueString(),
				ExistingPath: d.ExistingPath.ValueString(),
				Clone:        d.Clone.ValueBool(),
			})
		}
	}

	// cdroms
	if !m.CDROMs.IsNull() && !m.CDROMs.IsUnknown() {
		var cds []cdromModel
		m.CDROMs.ElementsAs(ctx, &cds, false)
		for _, c := range cds {
			inst.Devices.CDROMs = append(inst.Devices.CDROMs, client.CDROM{
				ID:          c.ID.ValueString(),
				IndexNumber: int(c.IndexNumber.ValueInt64()),
				BootOrder:   int(c.BootOrder.ValueInt64()),
				Connected:   c.Connected.ValueBool(),
				Path:        c.Path.ValueString(),
			})
		}
	}

	return inst
}

func instanceToModel(ctx context.Context, inst *client.Instance, m *InstanceResourceModel) {
	m.ID = types.StringValue(inst.ID)
	m.Name = types.StringValue(inst.Name)
	m.Description = types.StringValue(inst.Description)
	m.BootType = types.StringValue(inst.BootType)
	m.CPUCores = types.Int64Value(int64(inst.CPUCores))
	m.CPUSockets = types.Int64Value(int64(inst.CPUSockets))
	m.MemoryMB = types.Int64Value(int64(inst.MemoryMB))
	m.SiteId = types.StringValue(inst.SiteId)
	m.DatastoreId = types.StringValue(inst.DatastoreId)
	m.ImageId = types.StringValue(inst.ImageId)
	m.UserData = types.StringValue(inst.UserData)
	m.WinAutoattend = types.StringValue(inst.WinAutoattend)
	m.IPXEScript = types.StringValue(inst.IPXEScript)
	m.SecureBoot = types.BoolValue(inst.SecureBoot)
	m.TPM = types.BoolValue(inst.TPM)
	m.StartPoweredOff = types.BoolValue(inst.StartPoweredOff)
	m.InstanceType = types.StringValue(inst.InstanceType)
	m.PowerState = types.StringValue(inst.PowerState)
	m.InitializationStatus = types.StringValue(inst.InitializationStatus)
	m.PrimaryIPAddress = types.StringValue(inst.PrimaryIPAddress)
	m.PrimaryMacAddress = types.StringValue(inst.PrimaryMacAddress)
	m.MetadataIPAddress = types.StringValue(inst.MetadataIPAddress)
	m.VNCPort = types.Int64Value(int64(inst.VNCPort))
	m.Tags = tagsToModel(ctx, inst.Tags)

	// guest IP addresses
	guestIPs := make([]attr.Value, len(inst.GuestIPAddresses))
	for i, ip := range inst.GuestIPAddresses {
		guestIPs[i] = types.StringValue(ip)
	}
	m.GuestIPAddresses, _ = types.ListValue(types.StringType, guestIPs)

	// network interfaces
	nicObjs := make([]attr.Value, len(inst.Devices.NetworkInterfaces))
	for i, n := range inst.Devices.NetworkInterfaces {
		obj, _ := types.ObjectValue(networkInterfaceAttrTypes, map[string]attr.Value{
			"id":           types.StringValue(n.ID),
			"index_number": types.Int64Value(int64(n.IndexNumber)),
			"boot_order":   types.Int64Value(int64(n.BootOrder)),
			"model":        types.StringValue(n.Model),
			"connected":    types.BoolValue(n.Connected),
			"mac_address":  types.StringValue(n.MacAddress),
			"bridge_name":  types.StringValue(n.BridgeName),
			"subnet_id":    types.StringValue(n.SubnetId),
		})
		nicObjs[i] = obj
	}
	m.NetworkInterfaces, _ = types.ListValue(types.ObjectType{AttrTypes: networkInterfaceAttrTypes}, nicObjs)

	// storage disks
	diskObjs := make([]attr.Value, len(inst.Devices.StorageDisks))
	for i, d := range inst.Devices.StorageDisks {
		obj, _ := types.ObjectValue(storageDiskAttrTypes, map[string]attr.Value{
			"id":            types.StringValue(d.ID),
			"index_number":  types.Int64Value(int64(d.IndexNumber)),
			"boot_order":    types.Int64Value(int64(d.BootOrder)),
			"size_gb":       types.Int64Value(int64(d.SizeGB)),
			"bus_type":      types.StringValue(d.BusType),
			"path":          types.StringValue(d.Path),
			"datastore_id":  types.StringValue(d.DatastoreId),
			"existing_path": types.StringValue(d.ExistingPath),
			"clone":         types.BoolValue(d.Clone),
		})
		diskObjs[i] = obj
	}
	m.StorageDisks, _ = types.ListValue(types.ObjectType{AttrTypes: storageDiskAttrTypes}, diskObjs)

	// cdroms
	cdObjs := make([]attr.Value, len(inst.Devices.CDROMs))
	for i, c := range inst.Devices.CDROMs {
		obj, _ := types.ObjectValue(cdromAttrTypes, map[string]attr.Value{
			"id":           types.StringValue(c.ID),
			"index_number": types.Int64Value(int64(c.IndexNumber)),
			"boot_order":   types.Int64Value(int64(c.BootOrder)),
			"connected":    types.BoolValue(c.Connected),
			"path":         types.StringValue(c.Path),
		})
		cdObjs[i] = obj
	}
	m.CDROMs, _ = types.ListValue(types.ObjectType{AttrTypes: cdromAttrTypes}, cdObjs)

}

// ── list plan modifiers ────────────────────────────────────────────────────
//
// The framework cannot correlate list elements when a nested attribute value is
// unknown (e.g. existing_path references an image that is (known after apply)).
// Without element correlation, UseStateForUnknown on individual nested attrs is
// a no-op — there is no prior-state element to copy from.  These custom list-
// level modifiers use positional matching to copy computed fields from state
// before RequiresReplace evaluates, so unchanged instances are not replaced.

type preserveNICComputed struct{}

func (preserveNICComputed) Description(_ context.Context) string {
	return "Preserves computed NIC fields (id, mac_address) from prior state."
}
func (preserveNICComputed) MarkdownDescription(ctx context.Context) string {
	return preserveNICComputed{}.Description(ctx)
}
func (preserveNICComputed) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() || req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}
	var stateElems, planElems []networkInterfaceModel
	if req.StateValue.ElementsAs(ctx, &stateElems, false).HasError() {
		return
	}
	if req.PlanValue.ElementsAs(ctx, &planElems, false).HasError() {
		return
	}
	if len(planElems) != len(stateElems) {
		return
	}
	changed := false
	for i := range planElems {
		s, p := stateElems[i], &planElems[i]
		if p.ID.IsUnknown() && !s.ID.IsUnknown() && !s.ID.IsNull() {
			p.ID = s.ID
			changed = true
		}
		if p.MacAddress.IsUnknown() && !s.MacAddress.IsUnknown() && !s.MacAddress.IsNull() {
			p.MacAddress = s.MacAddress
			changed = true
		}
	}
	if changed {
		newList, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: networkInterfaceAttrTypes}, planElems)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			resp.PlanValue = newList
		}
	}
}

type preserveDiskComputed struct{}

func (preserveDiskComputed) Description(_ context.Context) string {
	return "Preserves computed disk fields (id, path, existing_path) from prior state."
}
func (preserveDiskComputed) MarkdownDescription(ctx context.Context) string {
	return preserveDiskComputed{}.Description(ctx)
}
func (preserveDiskComputed) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() || req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}
	var stateElems, planElems []storageDiskModel
	if req.StateValue.ElementsAs(ctx, &stateElems, false).HasError() {
		return
	}
	if req.PlanValue.ElementsAs(ctx, &planElems, false).HasError() {
		return
	}
	if len(planElems) != len(stateElems) {
		return
	}
	changed := false
	for i := range planElems {
		s, p := stateElems[i], &planElems[i]
		if p.ID.IsUnknown() && !s.ID.IsUnknown() && !s.ID.IsNull() {
			p.ID = s.ID
			changed = true
		}
		if p.Path.IsUnknown() && !s.Path.IsUnknown() && !s.Path.IsNull() {
			p.Path = s.Path
			changed = true
		}
		if p.ExistingPath.IsUnknown() && !s.ExistingPath.IsUnknown() && !s.ExistingPath.IsNull() {
			p.ExistingPath = s.ExistingPath
			changed = true
		}
	}
	if changed {
		newList, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: storageDiskAttrTypes}, planElems)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			resp.PlanValue = newList
		}
	}
}

type preserveCDROMComputed struct{}

func (preserveCDROMComputed) Description(_ context.Context) string {
	return "Preserves computed CDROM fields (id) from prior state."
}
func (preserveCDROMComputed) MarkdownDescription(ctx context.Context) string {
	return preserveCDROMComputed{}.Description(ctx)
}
func (preserveCDROMComputed) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() || req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}
	var stateElems, planElems []cdromModel
	if req.StateValue.ElementsAs(ctx, &stateElems, false).HasError() {
		return
	}
	if req.PlanValue.ElementsAs(ctx, &planElems, false).HasError() {
		return
	}
	if len(planElems) != len(stateElems) {
		return
	}
	changed := false
	for i := range planElems {
		s, p := stateElems[i], &planElems[i]
		if p.ID.IsUnknown() && !s.ID.IsUnknown() && !s.ID.IsNull() {
			p.ID = s.ID
			changed = true
		}
	}
	if changed {
		newList, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: cdromAttrTypes}, planElems)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			resp.PlanValue = newList
		}
	}
}

// ── tag helpers ────────────────────────────────────────────────────────────

// tagsFromModel converts a Terraform map[string]string into the API's []map[string]interface{}.
func tagsFromModel(ctx context.Context, m types.Map) []map[string]interface{} {
	if m.IsNull() || m.IsUnknown() {
		return []map[string]interface{}{}
	}
	var flat map[string]string
	m.ElementsAs(ctx, &flat, false)
	tags := make([]map[string]interface{}, 0, len(flat))
	for k, v := range flat {
		tags = append(tags, map[string]interface{}{k: v})
	}
	return tags
}

// tagsToModel converts the API's []map[string]interface{} into a Terraform map[string]string.
func tagsToModel(_ context.Context, tags []map[string]interface{}) types.Map {
	flat := make(map[string]attr.Value, len(tags))
	for _, tag := range tags {
		for k, v := range tag {
			flat[k] = types.StringValue(fmt.Sprintf("%v", v))
		}
	}
	m, _ := types.MapValue(types.StringType, flat)
	return m
}
