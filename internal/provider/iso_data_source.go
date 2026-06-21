package provider

import (
	"context"
	"fmt"

	"github.com/asashv/terraform-provider-asashv/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ISODataSource{}
var _ datasource.DataSourceWithConfigure = &ISODataSource{}

func NewISODataSource() datasource.DataSource {
	return &ISODataSource{}
}

type ISODataSource struct {
	client *client.Client
}

type ISODataSourceModel struct {
	ID        types.String  `tfsdk:"id"`
	Name      types.String  `tfsdk:"name"`
	SizeBytes types.Int64   `tfsdk:"size_bytes"`
	SizeGB    types.Float64 `tfsdk:"size_gb"`
}

func (d *ISODataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iso"
}

func (d *ISODataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up an ISO file in ASASHV by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ISO ID used by ASASHV.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "ISO filename, for example ubuntu-26.04-live-server-amd64.iso.",
			},
			"size_bytes": schema.Int64Attribute{
				Computed:    true,
				Description: "ISO size in bytes.",
			},
			"size_gb": schema.Float64Attribute{
				Computed:    true,
				Description: "ISO size in GB.",
			},
		},
	}
}

func (d *ISODataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ISODataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ISODataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	isos, err := d.client.ListISOs()
	if err != nil {
		resp.Diagnostics.AddError("Error listing ASASHV ISOs", err.Error())
		return
	}

	wantedName := config.Name.ValueString()

	for _, iso := range isos {
		if iso.Name == wantedName {
			config.ID = types.StringValue(iso.ID)
			config.Name = types.StringValue(iso.Name)
			config.SizeBytes = types.Int64Value(iso.SizeBytes)
			config.SizeGB = types.Float64Value(iso.SizeGB)

			resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
			return
		}
	}

	resp.Diagnostics.AddError(
		"ASASHV ISO not found",
		fmt.Sprintf("Could not find ISO with name %q.", wantedName),
	)
}
