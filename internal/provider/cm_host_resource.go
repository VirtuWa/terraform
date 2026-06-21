package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/asashv/terraform-provider-asashv/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &CMHostResource{}
var _ resource.ResourceWithConfigure = &CMHostResource{}

func NewCMHostResource() resource.Resource {
	return &CMHostResource{}
}

type CMHostResource struct {
	client *client.Client
}

type CMHostResourceModel struct {
	ID          types.String `tfsdk:"id"`
	IPAddress   types.String `tfsdk:"ip_address"`
	SSHUsername types.String `tfsdk:"ssh_username"`
	SSHPassword types.String `tfsdk:"ssh_password"`

	Hostname           types.String  `tfsdk:"hostname"`
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

func (r *CMHostResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cm_host"
}

func (r *CMHostResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Adds and manages a LejamCM host.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"ip_address": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ssh_username": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ssh_password": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"hostname": schema.StringAttribute{
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
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"cpu_cores_per_socket": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"ram_gb": schema.Float64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
			"cpu_usage_percent": schema.Float64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
			"memory_usage_percent": schema.Float64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
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

func (r *CMHostResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = c
}

func (r *CMHostResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CMHostResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateCMHost(client.CreateCMHostRequest{
		IPAddress:   plan.IPAddress.ValueString(),
		SSHUsername: plan.SSHUsername.ValueString(),
		SSHPassword: plan.SSHPassword.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error adding LejamCM host", err.Error())
		return
	}

	host := created
	if host.ID == "" || host.Status == "provisioning" {
		host, err = waitForCMHostByIP(ctx, r.client, plan.IPAddress.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error waiting for LejamCM host", err.Error())
			return
		}
	}

	applyCMHostToResourceModel(&plan, host)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CMHostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CMHostResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	host, err := findCMHost(r.client, state.ID.ValueString(), state.IPAddress.ValueString())
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Error reading LejamCM host", err.Error())
		return
	}

	applyCMHostToResourceModel(&state, host)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CMHostResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CMHostResourceModel
	var state CMHostResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	host, err := findCMHost(r.client, state.ID.ValueString(), state.IPAddress.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading LejamCM host during update", err.Error())
		return
	}

	plan.ID = state.ID
	applyCMHostToResourceModel(&plan, host)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CMHostResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CMHostResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hostID := state.ID.ValueString()

	err := r.client.DeleteCMHost(hostID)
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(strings.ToLower(err.Error()), "not found") {
			return
		}

		resp.Diagnostics.AddError("Error deleting LejamCM host", err.Error())
		return
	}

	if err := waitForCMHostDeleted(ctx, r.client, hostID, state.IPAddress.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error waiting for LejamCM host delete", err.Error())
		return
	}
}

func applyCMHostToResourceModel(model *CMHostResourceModel, host *client.CMHost) {
	model.ID = types.StringValue(host.ID)
	model.IPAddress = types.StringValue(host.IPAddress)
	model.Hostname = stringValueOrNull(host.Hostname)
	model.ClusterID = stringPtrValueOrNull(host.ClusterID)
	model.ClusterName = stringPtrValueOrNull(host.ClusterName)
	model.Vendor = stringValueOrNull(host.Vendor)
	model.Model = stringValueOrNull(host.Model)
	model.CPUSockets = types.Int64Value(host.CPUSockets)
	model.CPUCoresPerSocket = types.Int64Value(host.CPUCoresPerSocket)
	model.RAMGB = types.Float64Value(host.RAMGB)
	model.CPUUsagePercent = types.Float64Value(host.CPUUsagePercent)
	model.MemoryUsagePercent = types.Float64Value(host.MemoryUsagePercent)
	model.Status = stringValueOrNull(host.Status)
	model.MaintenanceMode = types.BoolValue(host.MaintenanceMode)
}

func findCMHost(c *client.Client, id string, ipAddress string) (*client.CMHost, error) {
	hosts, err := c.ListCMHosts()
	if err != nil {
		return nil, err
	}

	for _, host := range hosts {
		if id != "" && host.ID == id {
			return &host, nil
		}

		if ipAddress != "" && host.IPAddress == ipAddress {
			return &host, nil
		}
	}

	return nil, fmt.Errorf("LejamCM host not found")
}

func waitForCMHostByIP(ctx context.Context, c *client.Client, ipAddress string) (*client.CMHost, error) {
	timeout := time.After(180 * time.Second)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, fmt.Errorf("timed out waiting for LejamCM host %q", ipAddress)
		case <-ticker.C:
			host, err := findCMHost(c, "", ipAddress)
			if err == nil && host.ID != "" && host.Status != "provisioning" {
				return host, nil
			}
		}
	}
}

func waitForCMHostDeleted(ctx context.Context, c *client.Client, id string, ipAddress string) error {
	timeout := time.After(120 * time.Second)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timed out waiting for LejamCM host %q to be deleted", ipAddress)
		case <-ticker.C:
			_, err := findCMHost(c, id, ipAddress)
			if err != nil && strings.Contains(strings.ToLower(err.Error()), "not found") {
				return nil
			}
		}
	}
}
