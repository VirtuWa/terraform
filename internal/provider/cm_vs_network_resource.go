package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/asashv/terraform-provider-asashv/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &CMVSNetworkResource{}
var _ resource.ResourceWithConfigure = &CMVSNetworkResource{}

func NewCMVSNetworkResource() resource.Resource {
	return &CMVSNetworkResource{}
}

type CMVSNetworkResource struct {
	client *client.Client
}

type CMVSNetworkHostModel struct {
	Hostname   types.String   `tfsdk:"hostname"`
	HostID     types.String   `tfsdk:"host_id"`
	Interfaces []types.String `tfsdk:"interfaces"`
}

type CMVSNetworkResourceModel struct {
	ID              types.String           `tfsdk:"id"`
	BridgeName      types.String           `tfsdk:"bridge_name"`
	BondMode        types.String           `tfsdk:"bond_mode"`
	BondName        types.String           `tfsdk:"bond_name"`
	ClusterName     types.String           `tfsdk:"cluster_name"`
	ClusterID       types.String           `tfsdk:"cluster_id"`
	Description     types.String           `tfsdk:"description"`
	Hosts           []CMVSNetworkHostModel `tfsdk:"host"`
	Status          types.String           `tfsdk:"status"`
	PhysicalUplinks types.List             `tfsdk:"physical_uplinks"`
}

func (r *CMVSNetworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cm_vs_network"
}

func (r *CMVSNetworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a VS/OVS network in LejamCM.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"bridge_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"bond_mode": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("active-backup"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"bond_name": schema.StringAttribute{
				Computed: true,
			},
			"cluster_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cluster_id": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("Created by Terraform"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"physical_uplinks": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"host": schema.ListNestedBlock{
				Description: "Host and physical interfaces for this VS network.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"hostname": schema.StringAttribute{
							Required: true,
						},
						"host_id": schema.StringAttribute{
							Computed: true,
						},
						"interfaces": schema.ListAttribute{
							Required:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (r *CMVSNetworkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CMVSNetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CMVSNetworkResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cluster, err := findCMCluster(r.client, "", plan.ClusterName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving LejamCM cluster", err.Error())
		return
	}

	hostIDs, interfaces, err := resolveCMVSNetworkHosts(r.client, plan.Hosts)
	if err != nil {
		resp.Diagnostics.AddError("Error resolving VS network hosts", err.Error())
		return
	}

	input := client.CreateCMVSNetworkRequest{
		BridgeName:  plan.BridgeName.ValueString(),
		BondMode:    plan.BondMode.ValueString(),
		ClusterID:   cluster.ID,
		HostIDs:     hostIDs,
		Interfaces:  interfaces,
		Description: plan.Description.ValueString(),
	}

	validation, err := r.client.ValidateCMVSNetwork(input)
	if err != nil {
		resp.Diagnostics.AddError("Error validating LejamCM VS network", err.Error())
		return
	}

	if validation != nil && !validation.Valid {
		resp.Diagnostics.AddError(
			"LejamCM VS network validation failed",
			fmt.Sprintf("errors=%v warnings=%v", validation.Errors, validation.Warnings),
		)
		return
	}

	created, err := r.client.CreateCMVSNetwork(input)
	if err != nil {
		resp.Diagnostics.AddError("Error creating LejamCM VS network", err.Error())
		return
	}

	network := created
	if network.ID == "" {
		network, err = waitForCMVSNetworkByBridgeName(ctx, r.client, plan.BridgeName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error finding created LejamCM VS network", err.Error())
			return
		}
	}

	applyCMVSNetworkToModel(&plan, network)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CMVSNetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CMVSNetworkResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	network, err := findCMVSNetwork(r.client, state.ID.ValueString(), state.BridgeName.ValueString())
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Error reading LejamCM VS network", err.Error())
		return
	}

	applyCMVSNetworkToModel(&state, network)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CMVSNetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CMVSNetworkResourceModel
	var state CMVSNetworkResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	network, err := findCMVSNetwork(r.client, state.ID.ValueString(), state.BridgeName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading LejamCM VS network during update", err.Error())
		return
	}

	plan.ID = state.ID
	applyCMVSNetworkToModel(&plan, network)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CMVSNetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CMVSNetworkResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	networkID := state.ID.ValueString()

	err := r.client.DeleteCMVSNetwork(networkID)
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(strings.ToLower(err.Error()), "not found") {
			return
		}

		resp.Diagnostics.AddError("Error deleting LejamCM VS network", err.Error())
		return
	}

	if err := waitForCMVSNetworkDeleted(ctx, r.client, networkID, state.BridgeName.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error waiting for LejamCM VS network delete", err.Error())
		return
	}
}

func resolveCMVSNetworkHosts(c *client.Client, hosts []CMVSNetworkHostModel) ([]string, map[string][]string, error) {
	cmHosts, err := c.ListCMHosts()
	if err != nil {
		return nil, nil, err
	}

	hostIDs := make([]string, 0, len(hosts))
	interfaces := map[string][]string{}

	for _, wanted := range hosts {
		hostname := wanted.Hostname.ValueString()
		foundID := ""

		for _, host := range cmHosts {
			if host.Hostname == hostname {
				foundID = host.ID
				break
			}
		}

		if foundID == "" {
			return nil, nil, fmt.Errorf("LejamCM host with hostname %q not found", hostname)
		}

		ifaceList := expandStringList(wanted.Interfaces)
		if len(ifaceList) == 0 {
			return nil, nil, fmt.Errorf("host %q must have at least one interface", hostname)
		}

		hostIDs = append(hostIDs, foundID)
		interfaces[foundID] = ifaceList
	}

	return hostIDs, interfaces, nil
}

func applyCMVSNetworkToModel(model *CMVSNetworkResourceModel, network *client.CMVSNetwork) {
	model.ID = types.StringValue(network.ID)
	model.BridgeName = types.StringValue(network.BridgeName)
	model.BondMode = types.StringValue(network.BondMode)
	model.BondName = stringValueOrNull(network.BondName)
	model.ClusterID = stringValueOrNull(network.ClusterID)
	model.ClusterName = stringValueOrNull(network.ClusterName)
	model.Description = stringValueOrNull(network.Description)
	model.Status = stringValueOrNull(network.Status)
	model.PhysicalUplinks = stringListToTypesList(network.PhysicalUplinks)

	// Preserve config host order.
	currentHosts := model.Hosts
	if len(currentHosts) > 0 {
		for i := range currentHosts {
			hostname := currentHosts[i].Hostname.ValueString()
			for _, apiHost := range network.Hosts {
				if apiHost.Hostname == hostname {
					currentHosts[i].HostID = types.StringValue(apiHost.ID)
					break
				}
			}
		}
		model.Hosts = currentHosts
		return
	}

	model.Hosts = make([]CMVSNetworkHostModel, 0, len(network.Hosts))
	for _, host := range network.Hosts {
		model.Hosts = append(model.Hosts, CMVSNetworkHostModel{
			Hostname:   types.StringValue(host.Hostname),
			HostID:     types.StringValue(host.ID),
			Interfaces: flattenStringList(host.Interfaces),
		})
	}
}

func findCMVSNetwork(c *client.Client, id string, bridgeName string) (*client.CMVSNetwork, error) {
	networks, err := c.ListCMVSNetworks()
	if err != nil {
		return nil, err
	}

	for _, network := range networks {
		if id != "" && network.ID == id {
			return &network, nil
		}

		if bridgeName != "" && network.BridgeName == bridgeName {
			return &network, nil
		}
	}

	return nil, fmt.Errorf("LejamCM VS network not found")
}

func waitForCMVSNetworkByBridgeName(ctx context.Context, c *client.Client, bridgeName string) (*client.CMVSNetwork, error) {
	timeout := time.After(60 * time.Second)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, fmt.Errorf("timed out waiting for LejamCM VS network %q", bridgeName)
		case <-ticker.C:
			network, err := findCMVSNetwork(c, "", bridgeName)
			if err == nil {
				return network, nil
			}
		}
	}
}

func waitForCMVSNetworkDeleted(ctx context.Context, c *client.Client, id string, bridgeName string) error {
	timeout := time.After(60 * time.Second)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timed out waiting for LejamCM VS network %q to be deleted", bridgeName)
		case <-ticker.C:
			_, err := findCMVSNetwork(c, id, bridgeName)
			if err != nil && strings.Contains(strings.ToLower(err.Error()), "not found") {
				return nil
			}
		}
	}
}
