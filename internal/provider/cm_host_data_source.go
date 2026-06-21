package provider

import (
	"context"
	"fmt"

	"github.com/asashv/terraform-provider-asashv/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &CMHostDataSource{}
var _ datasource.DataSourceWithConfigure = &CMHostDataSource{}

func NewCMHostDataSource() datasource.DataSource {
	return &CMHostDataSource{}
}

type CMHostDataSource struct {
	client *client.Client
}

type CMHostDataSourceModel struct {
	ID                 types.String  `tfsdk:"id"`
	Hostname           types.String  `tfsdk:"hostname"`
	IPAddress          types.String  `tfsdk:"ip_address"`
	ClusterID          types.String  `tfsdk:"cluster_id"`
	ClusterName        types.String  `tfsdk:"cluster_name"`
	Vendor             types.String  `tfsdk:"vendor"`
	Model              types.String  `tfsdk:"model"`
	CPUSockets         types.Int64   `tfsdk:"cpu_sockets"`
	CPUCoresPerSocket  types.Int64   `tfsdk:"cpu_cores_per_socket"`
	RAMGB              types.Float64 `tfsdk:"ram_gb"`
	CPUUsagePercent    types.Float64 `tfsdk:"cpu_usage_percent"`
	MemoryUsagePercent types.Float64 `tfsdk:"memory_usage_percent"`
	Status             types.String  `tfsdk:"status"`
	MaintenanceMode    types.Bool    `tfsdk:"maintenance_mode"`
}

func (d *CMHostDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cm_host"
}

func (d *CMHostDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up a LejamCM host by hostname.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "LejamCM host ID.",
			},
			"hostname": schema.StringAttribute{
				Required:    true,
				Description: "LejamCM host hostname.",
			},
			"ip_address": schema.StringAttribute{
				Computed: true,
			},
			"cluster_id": schema.StringAttribute{
				Computed: true,
			},
			"cluster_name": schema.StringAttribute{
				Computed: true,
			},
			"vendor": schema.StringAttribute{
				Computed: true,
			},
			"model": schema.StringAttribute{
				Computed: true,
			},
			"cpu_sockets": schema.Int64Attribute{
				Computed: true,
			},
			"cpu_cores_per_socket": schema.Int64Attribute{
				Computed: true,
			},
			"ram_gb": schema.Float64Attribute{
				Computed: true,
			},
			"cpu_usage_percent": schema.Float64Attribute{
				Computed: true,
			},
			"memory_usage_percent": schema.Float64Attribute{
				Computed: true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"maintenance_mode": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

func (d *CMHostDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CMHostDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CMHostDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hosts, err := d.client.ListCMHosts()
	if err != nil {
		resp.Diagnostics.AddError("Error listing LejamCM hosts", err.Error())
		return
	}

	wantedHostname := config.Hostname.ValueString()

	for _, host := range hosts {
		if host.Hostname == wantedHostname {
			config.ID = types.StringValue(host.ID)
			config.Hostname = types.StringValue(host.Hostname)
			config.IPAddress = types.StringValue(host.IPAddress)
			config.ClusterID = stringPtrValueOrNull(host.ClusterID)
			config.ClusterName = stringPtrValueOrNull(host.ClusterName)
			config.Vendor = types.StringValue(host.Vendor)
			config.Model = types.StringValue(host.Model)
			config.CPUSockets = types.Int64Value(host.CPUSockets)
			config.CPUCoresPerSocket = types.Int64Value(host.CPUCoresPerSocket)
			config.RAMGB = types.Float64Value(host.RAMGB)
			config.CPUUsagePercent = types.Float64Value(host.CPUUsagePercent)
			config.MemoryUsagePercent = types.Float64Value(host.MemoryUsagePercent)
			config.Status = types.StringValue(host.Status)
			config.MaintenanceMode = types.BoolValue(host.MaintenanceMode)

			resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
			return
		}
	}

	resp.Diagnostics.AddError(
		"LejamCM host not found",
		fmt.Sprintf("Could not find LejamCM host with hostname %q.", wantedHostname),
	)
}

func stringPtrValueOrNull(value *string) types.String {
	if value == nil || *value == "" {
		return types.StringNull()
	}

	return types.StringValue(*value)
}
