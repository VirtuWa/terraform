package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/asashv/terraform-provider-asashv/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &CMClusterResource{}
var _ resource.ResourceWithConfigure = &CMClusterResource{}

func NewCMClusterResource() resource.Resource {
	return &CMClusterResource{}
}

type CMClusterResource struct {
	client *client.Client
}

type CMClusterResourceModel struct {
	ID                types.String   `tfsdk:"id"`
	Name              types.String   `tfsdk:"name"`
	Description       types.String   `tfsdk:"description"`
	VirtualIP         types.String   `tfsdk:"virtual_ip"`
	Hostnames         []types.String `tfsdk:"hostnames"`
	HostIDs           types.List     `tfsdk:"host_ids"`
	SharedStorageName types.String   `tfsdk:"shared_storage_name"`
	SharedStorageID   types.String   `tfsdk:"shared_storage_id"`
	HAEnabled         types.Bool     `tfsdk:"ha_enabled"`
	DRSEnabled        types.Bool     `tfsdk:"drs_enabled"`

	Status             types.String  `tfsdk:"status"`
	HostCount          types.Int64   `tfsdk:"host_count"`
	VMCount            types.Int64   `tfsdk:"vm_count"`
	CPUUsagePercent    types.Float64 `tfsdk:"cpu_usage_percent"`
	MemoryUsagePercent types.Float64 `tfsdk:"memory_usage_percent"`
}

func (r *CMClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cm_cluster"
}

func (r *CMClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a LejamCM HA cluster.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("Created by Terraform"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"virtual_ip": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"hostnames": schema.ListAttribute{
				Required:    true,
				Description: "LejamCM hostnames for the cluster. Minimum 2 hosts.",
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"host_ids": schema.ListAttribute{
				Computed:    true,
				Description: "Resolved LejamCM host IDs.",
				ElementType: types.StringType,
			},
			"shared_storage_name": schema.StringAttribute{
				Required:    true,
				Description: "Shared storage pool name, for example NFS.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"shared_storage_id": schema.StringAttribute{
				Computed: true,
			},
			"ha_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"drs_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"host_count": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"vm_count": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
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
		},
	}
}

func (r *CMClusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CMClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CMClusterResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hostnames := expandStringList(plan.Hostnames)

	if len(hostnames) < 2 {
		resp.Diagnostics.AddError("Invalid LejamCM cluster hostnames", "A LejamCM cluster requires at least 2 hostnames.")
		return
	}

	hostIDs, err := resolveCMHostnamesToIDs(r.client, hostnames)
	if err != nil {
		resp.Diagnostics.AddError("Error resolving LejamCM hostnames", err.Error())
		return
	}

	storage, err := findCMStorage(r.client, "", plan.SharedStorageName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving LejamCM shared storage", err.Error())
		return
	}

	created, err := r.client.CreateCMCluster(client.CreateCMClusterRequest{
		Name:            plan.Name.ValueString(),
		Description:     plan.Description.ValueString(),
		VirtualIP:       plan.VirtualIP.ValueString(),
		HostIDs:         hostIDs,
		SharedStorageID: storage.UUID,
		HAEnabled:       plan.HAEnabled.ValueBool(),
		DRSEnabled:      plan.DRSEnabled.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating LejamCM cluster", err.Error())
		return
	}

	cluster := created
	if cluster.ID == "" || cluster.HostCount == 0 {
		cluster, err = waitForCMClusterByName(ctx, r.client, plan.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error finding created LejamCM cluster", err.Error())
			return
		}
	}

	applyCMClusterToModel(&plan, cluster)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CMClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CMClusterResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cluster, err := findCMCluster(r.client, state.ID.ValueString(), state.Name.ValueString())
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Error reading LejamCM cluster", err.Error())
		return
	}

	applyCMClusterToModel(&state, cluster)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CMClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CMClusterResourceModel
	var state CMClusterResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cluster, err := findCMCluster(r.client, state.ID.ValueString(), state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading LejamCM cluster during update", err.Error())
		return
	}

	plan.ID = state.ID
	applyCMClusterToModel(&plan, cluster)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CMClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CMClusterResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := state.ID.ValueString()

	err := r.client.DeleteCMCluster(clusterID)
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(strings.ToLower(err.Error()), "not found") {
			return
		}

		resp.Diagnostics.AddError("Error deleting LejamCM cluster", err.Error())
		return
	}

	if err := waitForCMClusterDeleted(ctx, r.client, clusterID, state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error waiting for LejamCM cluster delete", err.Error())
		return
	}
}

func applyCMClusterToModel(model *CMClusterResourceModel, cluster *client.CMCluster) {
	model.ID = types.StringValue(cluster.ID)
	model.Name = types.StringValue(cluster.Name)
	model.Description = stringValueOrNull(cluster.Description)
	model.VirtualIP = types.StringValue(cluster.VirtualIP)

	// Keep the user-provided hostname order. LejamCM may return nodes in a different order,
	// and Terraform list attributes are order-sensitive.
	currentHostnames := expandStringList(model.Hostnames)
	if len(currentHostnames) > 0 {
		model.Hostnames = flattenStringList(currentHostnames)
		model.HostIDs = stringListToTypesList(clusterNodeIDsByHostnames(cluster.Nodes, currentHostnames))
	} else {
		model.Hostnames = flattenStringList(clusterNodeHostnames(cluster.Nodes))
		model.HostIDs = stringListToTypesList(clusterNodeIDs(cluster.Nodes))
	}

	model.SharedStorageID = types.StringValue(cluster.SharedStorageID)
	model.SharedStorageName = stringValueOrNull(cluster.SharedStorageName)

	model.HAEnabled = types.BoolValue(cluster.HAEnabled)
	model.DRSEnabled = types.BoolValue(cluster.DRSEnabled)
	model.Status = types.StringValue(cluster.Status)
	model.HostCount = types.Int64Value(cluster.HostCount)
	model.VMCount = types.Int64Value(cluster.VMCount)
	model.CPUUsagePercent = types.Float64Value(cluster.CPUUsagePercent)
	model.MemoryUsagePercent = types.Float64Value(cluster.MemoryUsagePercent)
}

func clusterNodeIDsByHostnames(nodes []client.CMClusterNode, hostnames []string) []string {
	result := make([]string, 0, len(hostnames))

	for _, wantedHostname := range hostnames {
		for _, node := range nodes {
			if node.Hostname == wantedHostname && node.ID != "" {
				result = append(result, node.ID)
				break
			}
		}
	}

	return result
}

func clusterNodeIDs(nodes []client.CMClusterNode) []string {
	result := make([]string, 0, len(nodes))

	for _, node := range nodes {
		if node.ID != "" {
			result = append(result, node.ID)
		}
	}

	return result
}

func clusterNodeHostnames(nodes []client.CMClusterNode) []string {
	result := make([]string, 0, len(nodes))

	for _, node := range nodes {
		if node.Hostname != "" {
			result = append(result, node.Hostname)
		}
	}

	return result
}

func findCMCluster(c *client.Client, id string, name string) (*client.CMCluster, error) {
	clusters, err := c.ListCMClusters()
	if err != nil {
		return nil, err
	}

	for _, cluster := range clusters {
		if id != "" && cluster.ID == id {
			return &cluster, nil
		}

		if name != "" && cluster.Name == name {
			return &cluster, nil
		}
	}

	return nil, fmt.Errorf("LejamCM cluster not found")
}

func waitForCMClusterByName(ctx context.Context, c *client.Client, name string) (*client.CMCluster, error) {
	timeout := time.After(90 * time.Second)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, fmt.Errorf("timed out waiting for LejamCM cluster %q", name)
		case <-ticker.C:
			cluster, err := findCMCluster(c, "", name)
			if err == nil && cluster.ID != "" && cluster.HostCount > 0 {
				return cluster, nil
			}
		}
	}
}

func waitForCMClusterDeleted(ctx context.Context, c *client.Client, id string, name string) error {
	timeout := time.After(90 * time.Second)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timed out waiting for LejamCM cluster %q to be deleted", name)
		case <-ticker.C:
			_, err := findCMCluster(c, id, name)
			if err != nil && strings.Contains(strings.ToLower(err.Error()), "not found") {
				return nil
			}
		}
	}
}
