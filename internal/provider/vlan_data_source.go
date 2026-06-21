package provider

import (
	"context"
	"fmt"

	"github.com/asashv/terraform-provider-asashv/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &VLANDataSource{}
var _ datasource.DataSourceWithConfigure = &VLANDataSource{}

func NewVLANDataSource() datasource.DataSource {
	return &VLANDataSource{}
}

type VLANDataSource struct {
	client *client.Client
}

type VLANDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	VLANID      types.Int64  `tfsdk:"vlan_id"`
	BridgeName  types.String `tfsdk:"bridge_name"`
}

func (d *VLANDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlan"
}

func (d *VLANDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up an ASASHV VLAN/network by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "VLAN ID as string.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "VLAN/network name, for example default.",
			},
			"description": schema.StringAttribute{
				Computed: true,
			},
			"vlan_id": schema.Int64Attribute{
				Computed: true,
			},
			"bridge_name": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *VLANDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VLANDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config VLANDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vlans, err := d.client.ListVLANs()
	if err != nil {
		resp.Diagnostics.AddError("Error listing ASASHV VLANs", err.Error())
		return
	}

	wantedName := config.Name.ValueString()

	for _, vlan := range vlans {
		if vlan.Name == wantedName {
			config.ID = types.StringValue(vlan.ID)
			config.Name = types.StringValue(vlan.Name)
			config.Description = types.StringValue(vlan.Description)
			config.VLANID = types.Int64Value(vlan.VLANID)
			config.BridgeName = types.StringValue(vlan.BridgeName)

			resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
			return
		}
	}

	resp.Diagnostics.AddError(
		"ASASHV VLAN not found",
		fmt.Sprintf("Could not find VLAN/network with name %q.", wantedName),
	)
}
