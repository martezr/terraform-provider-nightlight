package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/martezr/terraform-provider-nightlight/internal/client"
)

var _ datasource.DataSource = &ImageDataSource{}

func NewImageDataSource() datasource.DataSource { return &ImageDataSource{} }

type ImageDataSource struct{ c *client.Client }

type ImageDataSourceModel struct {
	ID               types.String  `tfsdk:"id"`
	Name             types.String  `tfsdk:"name"`
	Description      types.String  `tfsdk:"description"`
	Format           types.String  `tfsdk:"format"`
	SizeGB           types.Float64 `tfsdk:"size_gb"`
	OperatingSystem  types.String  `tfsdk:"operating_system"`
	Status           types.String  `tfsdk:"status"`
	Path             types.String  `tfsdk:"path"`
	DownloadProgress types.Int64   `tfsdk:"download_progress"`
	DatastoreId      types.String  `tfsdk:"datastore_id"`
	CreatedAt        types.String  `tfsdk:"created_at"`
}

func (d *ImageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image"
}

func (d *ImageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Looks up a Nightlight Cloud image by ID or name. Use the `id` output with `nightlight_instance.image_id`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Image ID. Provide to look up by ID, or leave unset to look up by `name`.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Image name. Used to look up the image when `id` is not provided.",
			},
			"description":       schema.StringAttribute{Computed: true},
			"path":              schema.StringAttribute{Computed: true},
			"format":            schema.StringAttribute{Computed: true},
			"size_gb":           schema.Float64Attribute{Computed: true},
			"operating_system":  schema.StringAttribute{Computed: true},
			"status":            schema.StringAttribute{Computed: true},
			"download_progress": schema.Int64Attribute{Computed: true},
			"datastore_id":      schema.StringAttribute{Computed: true},
			"created_at":        schema.StringAttribute{Computed: true},
		},
	}
}

func (d *ImageDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ImageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ImageDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.ID.IsNull() && data.Name.IsNull() {
		resp.Diagnostics.AddError("Missing filter", "At least one of `id` or `name` must be set.")
		return
	}

	var img *client.Image

	if !data.ID.IsNull() {
		found, err := d.c.GetImage(data.ID.ValueString())
		if err == client.ErrNotFound {
			resp.Diagnostics.AddError("Image not found", fmt.Sprintf("No image with ID %q exists.", data.ID.ValueString()))
			return
		}
		if err != nil {
			resp.Diagnostics.AddError("Error reading image", err.Error())
			return
		}
		img = found
	} else {
		all, err := d.c.ListImages()
		if err != nil {
			resp.Diagnostics.AddError("Error listing images", err.Error())
			return
		}
		for i := range all {
			if all[i].Name == data.Name.ValueString() {
				img = &all[i]
				break
			}
		}
		if img == nil {
			resp.Diagnostics.AddError("Image not found", fmt.Sprintf("No image named %q exists.", data.Name.ValueString()))
			return
		}
	}

	data.ID = types.StringValue(img.ID)
	data.Name = types.StringValue(img.Name)
	data.Description = types.StringValue(img.Description)
	data.Path = types.StringValue(img.Path)
	data.Format = types.StringValue(img.Format)
	data.SizeGB = types.Float64Value(img.SizeGB)
	data.OperatingSystem = types.StringValue(img.OperatingSystem)
	data.Status = types.StringValue(img.Status)
	data.DownloadProgress = types.Int64Value(int64(img.DownloadProgress))
	data.DatastoreId = types.StringValue(img.DatastoreId)
	data.CreatedAt = types.StringValue(img.CreatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
