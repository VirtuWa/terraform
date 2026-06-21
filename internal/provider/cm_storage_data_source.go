package provider

import (
	"context"
	"fmt"

	"github.com/asashv/terraform-provider-asashv/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &CMStorageDataSource{}
var _ datasource.DataSourceWithConfigure = &CMStorageDataSource{}

func NewCMStorageDataSource() datasource.DataSource {
	return &CMStorageDataSource{}
}

type CMStorageDataSource struct {
	client *client.Client
}

type CMStorageDataSourceModel struct {
	ID       types.String  `tfsdk:"id"`
	Name     types.String  `tfsdk:"name"`
	Type     types.String  `tfsdk:"type"`
	Path     types.String  `tfsdk:"path"`
	Server   types.String  `tfsdk:"server"`
	Vendor   types.String  `tfsdk:"vendor"`
	Protocol types.String  `tfsdk:"protocol"`
	Status   types.String  `tfsdk:"status"`
	SizeGB   types.Float64 `tfsdk:"size_gb"`
	UsedGB   types.Float64 `tfsdk:"used_gb"`
	FreeGB   types.Float64 `tfsdk:"free_gb"`
	Usage    types.String  `tfsdk:"usage"`
	Active   types.Bool    `tfsdk:"active"`
}

func (d *CMStorageDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cm_storage"
}

func (d *CMStorageDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up a LejamCM storage pool by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "LejamCM storage UUID.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Storage pool name, for example NFS.",
			},
			"type": schema.StringAttribute{
				Computed: true,
			},
			"path": schema.StringAttribute{
				Computed: true,
			},
			"server": schema.StringAttribute{
				Computed: true,
			},
			"vendor": schema.StringAttribute{
				Computed: true,
			},
			"protocol": schema.StringAttribute{
				Computed: true,
			},
			"status": schema.StringAttribute{
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
			"usage": schema.StringAttribute{
				Computed: true,
			},
			"active": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

func (d *CMStorageDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CMStorageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CMStorageDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pools, err := d.client.ListStoragePools()
	if err != nil {
		resp.Diagnostics.AddError("Error listing LejamCM storage pools", err.Error())
		return
	}

	wantedName := config.Name.ValueString()

	for _, pool := range pools {
		if pool.Name == wantedName {
			config.ID = types.StringValue(pool.UUID)
			config.Name = types.StringValue(pool.Name)
			config.Type = types.StringValue(pool.Type)
			config.Path = types.StringValue(pool.Path)
			config.Server = types.StringValue(pool.Server)
			config.Vendor = types.StringValue(pool.Vendor)
			config.Protocol = types.StringValue(pool.Protocol)
			config.Status = types.StringValue(pool.Status)
			config.SizeGB = types.Float64Value(pool.SizeGB)
			config.UsedGB = types.Float64Value(pool.UsedGB)
			config.FreeGB = types.Float64Value(pool.FreeGB)
			config.Usage = stringValueOrNull(pool.Usage)
			config.Active = types.BoolValue(pool.Status == "active")

			resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
			return
		}
	}

	resp.Diagnostics.AddError(
		"LejamCM storage pool not found",
		fmt.Sprintf("Could not find LejamCM storage pool with name %q.", wantedName),
	)
}
