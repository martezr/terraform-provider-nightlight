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

var _ resource.Resource = &DatastoreResource{}
var _ resource.ResourceWithImportState = &DatastoreResource{}

func NewDatastoreResource() resource.Resource { return &DatastoreResource{} }

type DatastoreResource struct{ c *client.Client }

type DatastoreResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	Path        types.String `tfsdk:"path"`
	LocalPath   types.String `tfsdk:"local_path"`
	Tags        types.Map    `tfsdk:"tags"`
}

func (r *DatastoreResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datastore"
}

func (r *DatastoreResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Nightlight Cloud datastore.",
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
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Datastore type: `local` or `nfs`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"path": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "NFS export path (e.g. `192.168.1.5:/exports/data`).",
			},
			"local_path": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tags": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *DatastoreResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DatastoreResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DatastoreResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.c.CreateDatastore(client.Datastore{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Type:        data.Type.ValueString(),
		Path:        data.Path.ValueString(),
		Tags:        tagsFromModel(ctx, data.Tags),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating datastore", err.Error())
		return
	}

	datastoreToModel(ctx, created, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatastoreResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DatastoreResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ds, err := r.c.GetDatastore(data.ID.ValueString())
	if err == client.ErrNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading datastore", err.Error())
		return
	}

	datastoreToModel(ctx, ds, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Datastores have no PUT endpoint — any change to type forces replacement.
func (r *DatastoreResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Datastore resources cannot be updated in place.")
}

func (r *DatastoreResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DatastoreResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.c.DeleteDatastore(data.ID.ValueString()); err != nil && err != client.ErrNotFound {
		resp.Diagnostics.AddError("Error deleting datastore", err.Error())
	}
}

func (r *DatastoreResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func datastoreToModel(ctx context.Context, ds *client.Datastore, m *DatastoreResourceModel) {
	m.ID = types.StringValue(ds.ID)
	m.Name = types.StringValue(ds.Name)
	m.Description = types.StringValue(ds.Description)
	m.Type = types.StringValue(ds.Type)
	m.Path = types.StringValue(ds.Path)
	m.LocalPath = types.StringValue(ds.LocalPath)
	m.Tags = tagsToModel(ctx, ds.Tags)
}
