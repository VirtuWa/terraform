package provider

import (
	"context"
	"fmt"

	"github.com/asashv/terraform-provider-asashv/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &CMISODataSource{}
var _ datasource.DataSourceWithConfigure = &CMISODataSource{}

func NewCMISODataSource() datasource.DataSource {
	return &CMISODataSource{}
}

type CMISODataSource struct {
	client *client.Client
}

type CMISODataSourceModel struct {
	ID        types.String  `tfsdk:"id"`
	Name      types.String  `tfsdk:"name"`
	SizeBytes types.Int64   `tfsdk:"size_bytes"`
	SizeGB    types.Float64 `tfsdk:"size_gb"`
}

func (d *CMISODataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cm_iso"
}

func (d *CMISODataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up a LejamCM ISO by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "LejamCM ISO ID.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "ISO filename.",
			},
			"size_bytes": schema.Int64Attribute{
				Computed: true,
			},
			"size_gb": schema.Float64Attribute{
				Computed: true,
			},
		},
	}
}

func (d *CMISODataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CMISODataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CMISODataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	isos, err := d.client.ListCMISOs()
	if err != nil {
		resp.Diagnostics.AddError("Error listing LejamCM ISOs", err.Error())
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
		"LejamCM ISO not found",
		fmt.Sprintf("Could not find LejamCM ISO with name %q.", wantedName),
	)
}
