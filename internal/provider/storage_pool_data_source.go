package provider

import (
	"context"
	"fmt"

	"github.com/asashv/terraform-provider-asashv/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &StoragePoolDataSource{}
var _ datasource.DataSourceWithConfigure = &StoragePoolDataSource{}

func NewStoragePoolDataSource() datasource.DataSource {
	return &StoragePoolDataSource{}
}

type StoragePoolDataSource struct {
	client *client.Client
}

type StoragePoolDataSourceModel struct {
	Name     types.String  `tfsdk:"name"`
	UUID     types.String  `tfsdk:"uuid"`
	State    types.String  `tfsdk:"state"`
	Status   types.String  `tfsdk:"status"`
	Active   types.Bool    `tfsdk:"active"`
	Vendor   types.String  `tfsdk:"vendor"`
	Server   types.String  `tfsdk:"server"`
	Protocol types.String  `tfsdk:"protocol"`
	Type     types.String  `tfsdk:"type"`
	SizeGB   types.Float64 `tfsdk:"size_gb"`
	UsedGB   types.Float64 `tfsdk:"used_gb"`
	FreeGB   types.Float64 `tfsdk:"free_gb"`
	Path     types.String  `tfsdk:"path"`
	Usage    types.String  `tfsdk:"usage"`
}

func (d *StoragePoolDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_pool"
}

func (d *StoragePoolDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up an ASASHV storage pool by name.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Storage pool name, for example NFS-Storage.",
			},
			"uuid": schema.StringAttribute{
				Computed: true,
			},
			"state": schema.StringAttribute{
				Computed: true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"active": schema.BoolAttribute{
				Computed: true,
			},
			"vendor": schema.StringAttribute{
				Computed: true,
			},
			"server": schema.StringAttribute{
				Computed: true,
			},
			"protocol": schema.StringAttribute{
				Computed: true,
			},
			"type": schema.StringAttribute{
				Computed: true,
			},
			"size_gb": schema.Float64Attribute{
				Computed: true,
			},
			"used_gb": schema.Float64Attribute{
				Computed: true,
			},
			"free_gb": schema.Float64Attribute{
				Computed: true,
			},
			"path": schema.StringAttribute{
				Computed: true,
			},
			"usage": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *StoragePoolDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}

	d.client = c
}

func (d *StoragePoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config StoragePoolDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pools, err := d.client.ListStoragePools()
	if err != nil {
		resp.Diagnostics.AddError("Error listing ASASHV storage pools", err.Error())
		return
	}

	wantedName := config.Name.ValueString()

	for _, pool := range pools {
		if pool.Name == wantedName {
			config.Name = types.StringValue(pool.Name)
			config.UUID = types.StringValue(pool.UUID)
			config.State = types.StringValue(pool.State)
			config.Status = types.StringValue(pool.Status)
			config.Active = types.BoolValue(pool.Active)
			config.Vendor = types.StringValue(pool.Vendor)
			config.Server = types.StringValue(pool.Server)
			config.Protocol = types.StringValue(pool.Protocol)
			config.Type = types.StringValue(pool.Type)
			config.SizeGB = types.Float64Value(pool.SizeGB)
			config.UsedGB = types.Float64Value(pool.UsedGB)
			config.FreeGB = types.Float64Value(pool.FreeGB)
			config.Path = types.StringValue(pool.Path)
			config.Usage = types.StringValue(pool.Usage)

			resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
			return
		}
	}

	resp.Diagnostics.AddError(
		"ASASHV storage pool not found",
		fmt.Sprintf("Could not find storage pool with name %q.", wantedName),
	)
}
