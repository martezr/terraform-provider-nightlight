package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/martezr/terraform-provider-nightlight/internal/client"
)

var _ resource.Resource = &InstanceBootCommandsResource{}

func NewInstanceBootCommandsResource() resource.Resource { return &InstanceBootCommandsResource{} }

type InstanceBootCommandsResource struct{ c *client.Client }

type InstanceBootCommandsModel struct {
	ID                types.String `tfsdk:"id"`
	InstanceID        types.String `tfsdk:"instance_id"`
	WaitForGuest      types.Bool   `tfsdk:"wait_for_guest"`
	GuestReadyTimeout types.Int64  `tfsdk:"guest_ready_timeout"`
	BootCommands      types.List   `tfsdk:"boot_commands"`
	PostBootCommands  types.List   `tfsdk:"post_boot_commands"`
}

func (r *InstanceBootCommandsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance_boot_commands"
}

func (r *InstanceBootCommandsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	bootCmdNested := schema.ListNestedAttribute{
		Optional: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"keys": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Key sequence to send (e.g. `root`, `<enter>`). Use an empty string for a pure wait step.",
				},
				"count": schema.Int64Attribute{
					Optional:            true,
					Computed:            true,
					Default:             int64default.StaticInt64(1),
					MarkdownDescription: "Number of times to send `keys`. Defaults to 1.",
				},
				"pause_between_ms": schema.Int64Attribute{
					Optional:            true,
					Computed:            true,
					Default:             int64default.StaticInt64(0),
					MarkdownDescription: "Milliseconds to pause between repetitions when `count` > 1. Defaults to 0.",
				},
				"pause_after_ms": schema.Int64Attribute{
					Optional:            true,
					Computed:            true,
					Default:             int64default.StaticInt64(0),
					MarkdownDescription: "Milliseconds to pause after all repetitions. Defaults to 0.",
				},
			},
		},
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Sends console key sequences to a Nightlight instance. Use this resource instead of embedding `boot_commands`/`post_boot_commands` directly on `nightlight_instance` when the key sequences need to reference computed instance attributes (e.g. `primary_ip_address`), since those values are only known after the instance is created.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"instance_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the instance to send console commands to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"wait_for_guest": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "When true (default), waits for the guest agent to report `ready` before sending `post_boot_commands`.",
			},
			"guest_ready_timeout": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(300),
				MarkdownDescription: "Maximum seconds to wait for the guest to be ready. Defaults to 300 (5 minutes).",
			},
			"boot_commands": bootCmdNested,
			"post_boot_commands": func() schema.ListNestedAttribute {
				cp := bootCmdNested
				cp.MarkdownDescription = "Ordered list of console key sequences sent after `wait_for_guest` completes (or immediately after `boot_commands` when `wait_for_guest` is false)."
				return cp
			}(),
		},
	}
}

func (r *InstanceBootCommandsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *InstanceBootCommandsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data InstanceBootCommandsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	instanceID := data.InstanceID.ValueString()
	data.ID = types.StringValue(instanceID)

	tflog.Info(ctx, "sending boot_commands", map[string]any{"instance_id": instanceID})
	if err := sendConsoleCommands(ctx, r.c, instanceID, data.BootCommands, "boot_commands"); err != nil {
		resp.Diagnostics.AddError("Error sending boot_commands", err.Error())
		return
	}

	if data.WaitForGuest.ValueBool() {
		timeout := time.Duration(data.GuestReadyTimeout.ValueInt64()) * time.Second
		tflog.Info(ctx, "waiting for guest agent", map[string]any{"instance_id": instanceID, "timeout_s": data.GuestReadyTimeout.ValueInt64()})
		deadline := time.Now().Add(timeout)
		for time.Now().Before(deadline) {
			time.Sleep(5 * time.Second)
			inst, err := r.c.GetInstance(instanceID)
			if err != nil {
				break
			}
			tflog.Info(ctx, "guest status poll", map[string]any{"instance_id": instanceID, "guest_status": inst.GuestStatus})
			if inst.GuestStatus == "ready" {
				break
			}
		}
	}

	tflog.Info(ctx, "sending post_boot_commands", map[string]any{"instance_id": instanceID})
	if err := sendConsoleCommands(ctx, r.c, instanceID, data.PostBootCommands, "post_boot_commands"); err != nil {
		resp.Diagnostics.AddError("Error sending post_boot_commands", err.Error())
		return
	}
	tflog.Info(ctx, "post_boot_commands complete", map[string]any{"instance_id": instanceID})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InstanceBootCommandsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data InstanceBootCommandsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.c.GetInstance(data.InstanceID.ValueString())
	if err == client.ErrNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading instance", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update is a no-op: boot commands are one-shot provisioning actions.
func (r *InstanceBootCommandsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data InstanceBootCommandsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InstanceBootCommandsResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Nothing to undo for console provisioning.
}

// sendConsoleCommands sends an ordered list of boot command entries to the instance VNC console.
func sendConsoleCommands(ctx context.Context, c *client.Client, instanceID string, list types.List, phase string) error {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	var cmds []bootCommandModel
	if diags := list.ElementsAs(ctx, &cmds, false); diags.HasError() {
		return fmt.Errorf("failed to decode console commands")
	}
	total := len(cmds)
	for i, cmd := range cmds {
		n := cmd.Count.ValueInt64()
		if n < 1 {
			n = 1
		}
		keys := cmd.Keys.ValueString()
		if keys == "" {
			tflog.Info(ctx, fmt.Sprintf("%s: step %d/%d", phase, i+1, total), map[string]any{
				"instance_id":    instanceID,
				"pause_after_ms": cmd.PauseAfterMs.ValueInt64(),
			})
		} else {
			tflog.Info(ctx, fmt.Sprintf("%s: step %d/%d", phase, i+1, total), map[string]any{
				"instance_id":      instanceID,
				"keys":             keys,
				"count":            n,
				"pause_between_ms": cmd.PauseBetweenMs.ValueInt64(),
				"pause_after_ms":   cmd.PauseAfterMs.ValueInt64(),
			})
		}
		if keys != "" {
			for j := int64(0); j < n; j++ {
				if err := c.SendInstanceConsoleKeys(instanceID, client.Command{KeyCode: keys}); err != nil {
					return err
				}
				if cmd.PauseBetweenMs.ValueInt64() > 0 && j < n-1 {
					time.Sleep(time.Duration(cmd.PauseBetweenMs.ValueInt64()) * time.Millisecond)
				}
			}
		}
		if cmd.PauseAfterMs.ValueInt64() > 0 {
			time.Sleep(time.Duration(cmd.PauseAfterMs.ValueInt64()) * time.Millisecond)
		}
	}
	tflog.Info(ctx, fmt.Sprintf("%s: complete", phase), map[string]any{"instance_id": instanceID, "steps": total})
	return nil
}
