package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/asashv/terraform-provider-asashv/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &CMVLANResource{}
var _ resource.ResourceWithConfigure = &CMVLANResource{}
var _ resource.ResourceWithImportState = &CMVLANResource{}

func NewCMVLANResource() resource.Resource {
	return &CMVLANResource{}
}

type CMVLANResource struct {
	client *client.Client
}

type CMVLANResourceModel struct {
	ID               types.String `tfsdk:"id"`
	VLANID           types.Int64  `tfsdk:"vlan_id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	BridgeName       types.String `tfsdk:"bridge_name"`
	ClusterID        types.String `tfsdk:"cluster_id"`
	ClusterName      types.String `tfsdk:"cluster_name"`
	PrimaryVSNetwork types.String `tfsdk:"primary_vs_network"`
	VSNetworkCount   types.Int64  `tfsdk:"vs_network_count"`
}

func (r *CMVLANResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cm_vlan"
}

func (r *CMVLANResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a VLAN/network in LejamCM.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"vlan_id": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
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
				Default:  stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"bridge_name": schema.StringAttribute{
				Computed: true,
			},
			"cluster_id": schema.StringAttribute{
				Computed: true,
			},
			"cluster_name": schema.StringAttribute{
				Computed: true,
			},
			"primary_vs_network": schema.StringAttribute{
				Computed: true,
			},
			"vs_network_count": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (r *CMVLANResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CMVLANResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CMVLANResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateCMVLAN(client.CreateVLANRequest{
		VLANID:      plan.VLANID.ValueInt64(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating LejamCM VLAN", err.Error())
		return
	}

	if created.ID == "" {
		found, err := waitForCMVLAN(ctx, r.client, plan.VLANID.ValueInt64(), plan.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error finding created LejamCM VLAN", err.Error())
			return
		}
		created = found
	}

	applyCMVLANToModel(&plan, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CMVLANResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CMVLANResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vlan, err := findCMVLAN(r.client, state.ID.ValueString(), state.VLANID.ValueInt64(), state.Name.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Error reading LejamCM VLAN", err.Error())
		return
	}

	applyCMVLANToModel(&state, vlan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CMVLANResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CMVLANResourceModel
	var state CMVLANResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vlan, err := findCMVLAN(r.client, state.ID.ValueString(), state.VLANID.ValueInt64(), state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading LejamCM VLAN during update", err.Error())
		return
	}

	plan.ID = state.ID
	applyCMVLANToModel(&plan, vlan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CMVLANResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CMVLANResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteID := state.ID.ValueString()
	if deleteID == "" {
		deleteID = fmt.Sprintf("%d", state.VLANID.ValueInt64())
	}

	err := r.client.DeleteCMVLAN(deleteID)
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(strings.ToLower(err.Error()), "not found") {
			return
		}

		resp.Diagnostics.AddError("Error deleting LejamCM VLAN", err.Error())
		return
	}

	timeout := time.After(60 * time.Second)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			resp.Diagnostics.AddError("Error waiting for LejamCM VLAN delete", ctx.Err().Error())
			return

		case <-timeout:
			resp.Diagnostics.AddError(
				"LejamCM VLAN still exists after delete",
				fmt.Sprintf("Delete request completed, but VLAN %q still exists after waiting and retrying.", deleteID),
			)
			return

		case <-ticker.C:
			_, err := findCMVLAN(r.client, deleteID, state.VLANID.ValueInt64(), state.Name.ValueString())
			if err != nil && strings.Contains(strings.ToLower(err.Error()), "not found") {
				return
			}

			// Retry delete because LejamCM may return success before the VLAN is actually removed.
			err = r.client.DeleteCMVLAN(deleteID)
			if err != nil {
				if strings.Contains(err.Error(), "404") || strings.Contains(strings.ToLower(err.Error()), "not found") {
					return
				}
			}
		}
	}
}

func (r *CMVLANResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	vlans, err := r.client.ListVLANs()
	if err != nil {
		resp.Diagnostics.AddError("Error importing LejamCM VLAN", err.Error())
		return
	}

	for _, vlan := range vlans {
		if vlan.ID == req.ID || strconv.FormatInt(vlan.VLANID, 10) == req.ID {
			var state CMVLANResourceModel
			applyCMVLANToModel(&state, &vlan)
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}

	resp.Diagnostics.AddError("LejamCM VLAN not found", fmt.Sprintf("Could not import VLAN %q.", req.ID))
}

func applyCMVLANToModel(model *CMVLANResourceModel, vlan *client.VLAN) {
	model.ID = types.StringValue(vlan.ID)
	model.VLANID = types.Int64Value(vlan.VLANID)
	model.Name = types.StringValue(vlan.Name)

	if model.Description.IsNull() || model.Description.IsUnknown() {
		model.Description = stringValueOrNull(vlan.Description)
	}

	model.BridgeName = stringValueOrNull(vlan.BridgeName)
	model.ClusterID = stringPtrValueOrNull(vlan.ClusterID)
	model.ClusterName = stringPtrValueOrNull(vlan.ClusterName)
	model.PrimaryVSNetwork = stringValueOrNull(vlan.PrimaryVSNetwork)
	model.VSNetworkCount = types.Int64Value(vlan.VSNetworkCount)
}

func findCMVLAN(c *client.Client, id string, vlanID int64, name string) (*client.VLAN, error) {
	vlans, err := c.ListVLANs()
	if err != nil {
		return nil, err
	}

	for _, vlan := range vlans {
		if vlan.ID == id {
			return &vlan, nil
		}

		if vlan.VLANID == vlanID && vlan.Name == name {
			return &vlan, nil
		}
	}

	return nil, fmt.Errorf("LejamCM VLAN not found")
}

func waitForCMVLAN(ctx context.Context, c *client.Client, vlanID int64, name string) (*client.VLAN, error) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeout := time.After(30 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, fmt.Errorf("timed out waiting for LejamCM VLAN %q", name)
		case <-ticker.C:
			vlan, err := findCMVLAN(c, "", vlanID, name)
			if err == nil {
				return vlan, nil
			}
		}
	}
}
