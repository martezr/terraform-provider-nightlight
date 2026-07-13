package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/martezr/terraform-provider-nightlight/internal/client"
)

var _ resource.Resource = &ContentLibraryResource{}
var _ resource.ResourceWithImportState = &ContentLibraryResource{}

func NewContentLibraryResource() resource.Resource { return &ContentLibraryResource{} }

type ContentLibraryResource struct{ c *client.Client }

type ContentLibraryResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	DepotURL     types.String `tfsdk:"depot_url"`
	DepotToken   types.String `tfsdk:"depot_token"`
	DatastoreId  types.String `tfsdk:"datastore_id"`
	SyncInterval types.String `tfsdk:"sync_interval"`
	SyncStatus   types.String `tfsdk:"sync_status"`
	SyncError    types.String `tfsdk:"sync_error"`
	LastSyncAt   types.String `tfsdk:"last_sync_at"`
	ItemCount    types.Int64  `tfsdk:"item_count"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

func (r *ContentLibraryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_content_library"
}

func (r *ContentLibraryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Nightlight Cloud content library backed by a remote Depot server.",
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
			"depot_url": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "URL of the remote Depot server (e.g. `http://depot.example.com`).",
			},
			"depot_token": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "Authentication token for the Depot server.",
			},
			"datastore_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the datastore where imported images will be stored.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sync_interval": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("manual"),
				MarkdownDescription: "How often to sync with the Depot: `manual`, `1h`, `6h`, or `24h`.",
			},
			"sync_status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Current sync status: `idle`, `syncing`, `synced`, or `error`.",
			},
			"sync_error": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Last sync error message, if any.",
			},
			"last_sync_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "RFC3339 timestamp of the last successful sync.",
			},
			"item_count": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of items currently in the library.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *ContentLibraryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ContentLibraryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ContentLibraryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.c.CreateContentLibrary(client.ContentLibrary{
		Name:         data.Name.ValueString(),
		Description:  data.Description.ValueString(),
		DepotURL:     data.DepotURL.ValueString(),
		DepotToken:   data.DepotToken.ValueString(),
		DatastoreId:  data.DatastoreId.ValueString(),
		SyncInterval: data.SyncInterval.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating content library", err.Error())
		return
	}

	contentLibraryToModel(created, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContentLibraryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ContentLibraryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lib, err := r.c.GetContentLibrary(data.ID.ValueString())
	if err == client.ErrNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading content library", err.Error())
		return
	}

	contentLibraryToModel(lib, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContentLibraryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ContentLibraryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.c.UpdateContentLibrary(data.ID.ValueString(), client.UpdateContentLibraryRequest{
		Name:         data.Name.ValueString(),
		Description:  data.Description.ValueString(),
		DepotURL:     data.DepotURL.ValueString(),
		DepotToken:   data.DepotToken.ValueString(),
		SyncInterval: data.SyncInterval.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating content library", err.Error())
		return
	}

	contentLibraryToModel(updated, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContentLibraryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ContentLibraryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.c.DeleteContentLibrary(data.ID.ValueString()); err != nil && err != client.ErrNotFound {
		resp.Diagnostics.AddError("Error deleting content library", err.Error())
	}
}

func (r *ContentLibraryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func contentLibraryToModel(lib *client.ContentLibrary, m *ContentLibraryResourceModel) {
	m.ID = types.StringValue(lib.ID)
	m.Name = types.StringValue(lib.Name)
	m.Description = types.StringValue(lib.Description)
	m.DepotURL = types.StringValue(lib.DepotURL)
	m.DatastoreId = types.StringValue(lib.DatastoreId)
	m.SyncInterval = types.StringValue(lib.SyncInterval)
	m.SyncStatus = types.StringValue(lib.SyncStatus)
	m.SyncError = types.StringValue(lib.SyncError)
	m.LastSyncAt = types.StringValue(lib.LastSyncAt)
	m.ItemCount = types.Int64Value(int64(lib.ItemCount))
	m.CreatedAt = types.StringValue(lib.CreatedAt)
	m.UpdatedAt = types.StringValue(lib.UpdatedAt)
	// depot_token is write-only; the API never returns it, so preserve the plan value.
}
