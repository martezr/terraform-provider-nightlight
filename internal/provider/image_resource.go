package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/martezr/terraform-provider-nightlight/internal/client"
)

var _ resource.Resource = &ImageResource{}
var _ resource.ResourceWithImportState = &ImageResource{}

func NewImageResource() resource.Resource { return &ImageResource{} }

type ImageResource struct{ c *client.Client }

type ImageResourceModel struct {
	ID               types.String  `tfsdk:"id"`
	Name             types.String  `tfsdk:"name"`
	Description      types.String  `tfsdk:"description"`
	OperatingSystem  types.String  `tfsdk:"operating_system"`
	Format           types.String  `tfsdk:"format"`
	SizeGB           types.Float64 `tfsdk:"size_gb"`
	DatastoreId      types.String  `tfsdk:"datastore_id"`
	SourceType       types.String  `tfsdk:"source_type"`
	SourcePath       types.String  `tfsdk:"source_path"`
	SourceURL        types.String  `tfsdk:"source_url"`
	FileName         types.String  `tfsdk:"file_name"`
	Path             types.String  `tfsdk:"path"`
	Status           types.String  `tfsdk:"status"`
	DownloadProgress types.Int64   `tfsdk:"download_progress"`
	Tags             types.Map     `tfsdk:"tags"`
	CreatedAt        types.String  `tfsdk:"created_at"`
}

func (r *ImageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image"
}

func (r *ImageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replaceOnChange := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Nightlight Cloud image. Supports registering images from a URL, a datastore file, or a manual path.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:      true,
				PlanModifiers: replaceOnChange,
			},
			"description": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Default:       stringdefault.StaticString(""),
				PlanModifiers: replaceOnChange,
			},
			"operating_system": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Default:       stringdefault.StaticString(""),
				PlanModifiers: replaceOnChange,
			},
			"format": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("qcow2"),
				MarkdownDescription: "Disk image format. Defaults to `qcow2`.",
				PlanModifiers:       replaceOnChange,
			},
			"size_gb": schema.Float64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Disk size in GB. Computed for URL and file source types.",
			},
			"datastore_id": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Default:       stringdefault.StaticString(""),
				PlanModifiers: replaceOnChange,
			},
			"source_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "How to populate the image: `url` to download from a remote URL, `file` to copy from a datastore path, or empty for a manual record.",
				PlanModifiers:       replaceOnChange,
			},
			"source_path": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Filename inside the datastore to copy (required when `source_type` is `file`).",
				PlanModifiers:       replaceOnChange,
			},
			"source_url": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Remote URL to download the image from (required when `source_type` is `url`).",
				PlanModifiers:       replaceOnChange,
			},
			"file_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Destination filename for URL downloads. Defaults to the basename of `source_url`.",
				PlanModifiers:       replaceOnChange,
			},
			"path": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"download_progress": schema.Int64Attribute{
				Computed: true,
			},
			"tags": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ImageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ImageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ImageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.c.CreateImage(client.CreateImageRequest{
		Name:            data.Name.ValueString(),
		Description:     data.Description.ValueString(),
		OperatingSystem: data.OperatingSystem.ValueString(),
		Format:          data.Format.ValueString(),
		SizeGB:          data.SizeGB.ValueFloat64(),
		DatastoreId:     data.DatastoreId.ValueString(),
		SourceType:      data.SourceType.ValueString(),
		SourcePath:      data.SourcePath.ValueString(),
		SourceURL:       data.SourceURL.ValueString(),
		FileName:        data.FileName.ValueString(),
		Tags:            tagsFromModel(ctx, data.Tags),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating image", err.Error())
		return
	}

	// For URL-based downloads, poll until the image is no longer downloading.
	if created.Status == "downloading" {
		img, pollErr := r.pollUntilReady(created.ID)
		if pollErr != nil {
			resp.Diagnostics.AddError("Error waiting for image download", pollErr.Error())
			return
		}
		created = img
	}

	imageToModel(ctx, created, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ImageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ImageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	img, err := r.c.GetImage(data.ID.ValueString())
	if err == client.ErrNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Error reading image", err.Error())
		return
	}

	imageToModel(ctx, img, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Images have no PUT/PATCH endpoint — all changes require replacement.
func (r *ImageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Image resources cannot be updated in place.")
}

func (r *ImageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ImageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.c.DeleteImage(data.ID.ValueString()); err != nil && err != client.ErrNotFound {
		resp.Diagnostics.AddError("Error deleting image", err.Error())
	}
}

func (r *ImageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// pollUntilReady polls the image until its status leaves "downloading" or a 30-minute timeout is reached.
func (r *ImageResource) pollUntilReady(id string) (*client.Image, error) {
	deadline := time.Now().Add(30 * time.Minute)
	for time.Now().Before(deadline) {
		time.Sleep(5 * time.Second)
		img, err := r.c.GetImage(id)
		if err != nil {
			return nil, err
		}
		if img.Status != "downloading" {
			if img.Status == "error" {
				return nil, fmt.Errorf("image download failed (status=error)")
			}
			return img, nil
		}
	}
	return nil, fmt.Errorf("timed out waiting for image to become ready")
}

func imageToModel(ctx context.Context, img *client.Image, m *ImageResourceModel) {
	m.ID = types.StringValue(img.ID)
	m.Name = types.StringValue(img.Name)
	m.Description = types.StringValue(img.Description)
	m.OperatingSystem = types.StringValue(img.OperatingSystem)
	m.Format = types.StringValue(img.Format)
	m.SizeGB = types.Float64Value(img.SizeGB)
	m.DatastoreId = types.StringValue(img.DatastoreId)
	m.Path = types.StringValue(img.Path)
	m.Status = types.StringValue(img.Status)
	m.DownloadProgress = types.Int64Value(int64(img.DownloadProgress))
	m.Tags = tagsToModel(ctx, img.Tags)
	m.CreatedAt = types.StringValue(img.CreatedAt)
}
